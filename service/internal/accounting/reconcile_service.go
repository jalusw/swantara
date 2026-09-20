package accounting

import (
	"context"
	"fmt"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

type ReconcileRequest struct {
	OrganizationID    uint64
	DebitLineID       uint64
	CreditLineID      uint64
	Amount            float64
	WriteOffAccountID *uint64
	Tolerance         float64
}

type ReconcileResult struct {
	Partial    *AccountPartialReconcile
	Full       *AccountFullReconcile
	Reconciled bool
}

type ReconcileService struct {
	lines     JournalLineDAO
	partials  AccountPartialReconcileDAO
	fulls     AccountFullReconcileDAO
	movements JournalEntryDAO
	poster    Poster
	tx        db.Transactioner
}

func NewReconcileService(
	lines JournalLineDAO,
	partials AccountPartialReconcileDAO,
	fulls AccountFullReconcileDAO,
	movements JournalEntryDAO,
	tx db.Transactioner,
) ReconcileService {
	return ReconcileService{lines: lines, partials: partials, fulls: fulls, movements: movements, tx: tx}
}

func (s ReconcileService) WithPoster(poster Poster) ReconcileService {
	s.poster = poster
	return s
}

func (s ReconcileService) Reconcile(ctx context.Context, request ReconcileRequest) (*ReconcileResult, error) {
	reqAmount := amount.FromFloat64(request.Amount)
	if !reqAmount.IsPositive() {
		return nil, ErrReconcileAmount
	}
	debitLine, err := s.lines.Find(ctx, request.DebitLineID)
	if err != nil {
		return nil, err
	}
	if debitLine == nil {
		return nil, ErrLineNotFound
	}
	creditLine, err := s.lines.Find(ctx, request.CreditLineID)
	if err != nil {
		return nil, err
	}
	if creditLine == nil {
		return nil, ErrLineNotFound
	}
	debitMove, err := s.movements.Find(ctx, debitLine.EntryID)
	if err != nil {
		return nil, err
	}
	if debitMove == nil || debitMove.OrganizationID != request.OrganizationID {
		return nil, ErrLineNotInOrganization
	}
	creditMove, err := s.movements.Find(ctx, creditLine.EntryID)
	if err != nil {
		return nil, err
	}
	if creditMove == nil || creditMove.OrganizationID != request.OrganizationID {
		return nil, ErrLineNotInOrganization
	}
	if debitLine.AccountID != creditLine.AccountID {
		return nil, ErrLinesDifferentAccount
	}

	debitRemaining, err := s.remaining(ctx, debitLine, true)
	if err != nil {
		return nil, err
	}
	creditRemaining, err := s.remaining(ctx, creditLine, false)
	if err != nil {
		return nil, err
	}
	tolerance := amount.FromFloat64(request.Tolerance).Abs().Round(postingPrecision)
	allowsTolerance := func(remaining amount.Amount) bool {
		if tolerance.IsZero() {
			return false
		}
		diff := remaining.Sub(reqAmount).Abs().Round(postingPrecision)
		return diff.LessThan(tolerance) || diff.Equal(tolerance)
	}
	if reqAmount.GreaterThan(debitRemaining) && !allowsTolerance(debitRemaining) {
		return nil, ErrReconcileExceedsBalance
	}
	if reqAmount.GreaterThan(creditRemaining) && !allowsTolerance(creditRemaining) {
		return nil, ErrReconcileExceedsBalance
	}
	if reqAmount.GreaterThan(debitRemaining) {
		reqAmount = debitRemaining
	}
	if reqAmount.GreaterThan(creditRemaining) {
		reqAmount = creditRemaining
	}

	result := &ReconcileResult{Partial: &AccountPartialReconcile{}}
	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		partial := &AccountPartialReconcile{
			DebitLineID:  request.DebitLineID,
			CreditLineID: request.CreditLineID,
			Amount:       reqAmount.Float64(),
		}
		created, err := s.partials.CreateTx(ctx, tx, partial)
		if err != nil {
			return err
		}
		result.Partial = created

		if reqAmount.Equal(debitRemaining) && reqAmount.Equal(creditRemaining) {
			full := &AccountFullReconcile{Name: helper.Ptr[string](fmt.Sprintf("Reconciliation %d", created.ID))}
			createdFull, err := s.fulls.CreateTx(ctx, tx, full)
			if err != nil {
				return err
			}
			created.FullReconcileID = &createdFull.ID
			if _, err := s.partials.UpdateTx(ctx, tx, created); err != nil {
				return err
			}
			debitLine.Reconciled = true
			debitLine.FullReconcileID = &createdFull.ID
			if _, err := s.lines.UpdateTx(ctx, tx, debitLine); err != nil {
				return err
			}
			creditLine.Reconciled = true
			creditLine.FullReconcileID = &createdFull.ID
			if _, err := s.lines.UpdateTx(ctx, tx, creditLine); err != nil {
				return err
			}
			result.Full = createdFull
			result.Reconciled = true
			return nil
		}

		if request.WriteOffAccountID != nil {
			debitRem := debitRemaining.Sub(reqAmount)
			creditRem := creditRemaining.Sub(reqAmount)
			diff := debitRem.Sub(creditRem).Abs()
			if !diff.IsZero() && s.poster != nil {
				journalID := debitMove.JournalID
				if journalID == 0 {
					journalID = creditMove.JournalID
				}
				accountID := debitLine.AccountID
				var writeOffLines []PostingLine
				if debitRem.GreaterThan(creditRem) {
					writeOffLines = []PostingLine{
						{AccountID: accountID, Name: "Write-off balance", Credit: diff},
						{AccountID: *request.WriteOffAccountID, Name: "Write-off", Debit: diff},
					}
				} else {
					writeOffLines = []PostingLine{
						{AccountID: accountID, Name: "Write-off balance", Debit: diff},
						{AccountID: *request.WriteOffAccountID, Name: "Write-off", Credit: diff},
					}
				}
				if _, err := s.poster.PostTx(ctx, tx, PostRequest{
					OrganizationID: request.OrganizationID,
					JournalID:      journalID,
					Date:           debitMove.Date,
					Description:    "Reconcile write-off",
					Lines:          writeOffLines,
				}); err != nil {
					return err
				}
			}
			full := &AccountFullReconcile{Name: helper.Ptr[string](fmt.Sprintf("Reconciliation %d", created.ID))}
			createdFull, err := s.fulls.CreateTx(ctx, tx, full)
			if err != nil {
				return err
			}
			created.FullReconcileID = &createdFull.ID
			if _, err := s.partials.UpdateTx(ctx, tx, created); err != nil {
				return err
			}
			debitLine.Reconciled = true
			debitLine.FullReconcileID = &createdFull.ID
			if _, err := s.lines.UpdateTx(ctx, tx, debitLine); err != nil {
				return err
			}
			creditLine.Reconciled = true
			creditLine.FullReconcileID = &createdFull.ID
			if _, err := s.lines.UpdateTx(ctx, tx, creditLine); err != nil {
				return err
			}
			result.Full = createdFull
			result.Reconciled = true
			return nil
		}

		if reqAmount.Equal(debitRemaining) {
			debitLine.Reconciled = true
			if _, err := s.lines.UpdateTx(ctx, tx, debitLine); err != nil {
				return err
			}
		}
		if reqAmount.Equal(creditRemaining) {
			creditLine.Reconciled = true
			if _, err := s.lines.UpdateTx(ctx, tx, creditLine); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s ReconcileService) remaining(ctx context.Context, line *JournalLine, debit bool) (amount.Amount, error) {
	var nominal amount.Amount
	if debit && line.Debit.IsPositive() {
		nominal = line.Debit
	} else if !debit && line.Credit.IsPositive() {
		nominal = line.Credit
	} else {
		return nominal, ErrInvalidLine
	}
	total, err := s.partials.TotalByLine(ctx, line.ID)
	if err != nil {
		return nominal, err
	}
	remaining := nominal.Sub(amount.FromFloat64(total))
	if remaining.IsNegative() {
		return nominal, ErrReconcileExceedsBalance
	}
	return remaining, nil
}
