package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type BankStatementLineRequest struct {
	Date           *time.Time
	Amount         float64
	CurrencyCode   *string
	FeeAmount      float64
	InterestAmount float64
	ContactID      *uint64
	Ref            string
	Narration      string
}

type CreateBankStatementRequest struct {
	OrganizationID uint64
	JournalID      uint64
	Date           time.Time
	BalanceStart   float64
	BalanceEnd     float64
	CurrencyCode   *string
	Lines          []BankStatementLineRequest
}

type BankStatementService struct {
	statements BankStatementDAO
	lines      BankStatementLineDAO
	payments   PaymentDAO
	journals   dao.CRUD[reference.Journal]
	tx         db.Transactioner
	now        func() time.Time
	poster     Poster
	charges    BankChargeResolver
}

func (s BankStatementService) WithPosting(poster Poster, charges BankChargeResolver) BankStatementService {
	s.poster = poster
	s.charges = charges
	return s
}

func NewBankStatementService(
	statements BankStatementDAO,
	lines BankStatementLineDAO,
	payments PaymentDAO,
	journals dao.CRUD[reference.Journal],
	tx db.Transactioner,
) BankStatementService {
	return BankStatementService{
		statements: statements,
		lines:      lines,
		payments:   payments,
		journals:   journals,
		tx:         tx,
		now:        time.Now,
	}
}

func (s BankStatementService) Create(ctx context.Context, request CreateBankStatementRequest) (*BankStatement, error) {
	if len(request.Lines) == 0 {
		return nil, ErrStatementNoLines
	}
	if _, err := s.journals.Find(ctx, request.JournalID); err != nil {
		return nil, err
	}

	statementDate := request.Date
	if statementDate.IsZero() {
		statementDate = s.now().UTC()
	}
	statement := &BankStatement{
		JournalID:    helper.Ptr(request.JournalID),
		Name:         helper.Ptr("Bank statement"),
		Date:         &statementDate,
		BalanceStart: amount.FromFloat64(request.BalanceStart).Round(4),
		BalanceEnd:   amount.FromFloat64(request.BalanceEnd).Round(4),
		CurrencyCode: request.CurrencyCode,
		State:        BankStatementStateOpen,
	}
	lines := make([]*BankStatementLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &BankStatementLine{
			Date:           line.Date,
			Amount:         amount.FromFloat64(line.Amount).Round(4),
			CurrencyCode:   line.CurrencyCode,
			FeeAmount:      amount.FromFloat64(line.FeeAmount).Round(4),
			InterestAmount: amount.FromFloat64(line.InterestAmount).Round(4),
			ContactID:      line.ContactID,
			Ref:            helper.Ptr(line.Ref),
			Narration:      helper.Ptr(line.Narration),
		}
	}

	var created *BankStatement
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		var err error
		created, err = s.statements.CreateWithLinesTx(ctx, tx, statement, lines)
		return err
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s BankStatementService) Find(ctx context.Context, id uint64) (*BankStatement, error) {
	return s.statements.Find(ctx, id)
}

func (s BankStatementService) List(ctx context.Context, q *query.Query) (*query.Page[BankStatement], error) {
	return s.statements.List(ctx, q)
}

func (s BankStatementService) ListLines(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	return s.lines.ListByStatement(ctx, statementID)
}

func (s BankStatementService) Match(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	statement, err := s.statements.Find(ctx, statementID)
	if err != nil {
		return nil, err
	}
	if statement == nil {
		return nil, ErrStatementNotFound
	}
	if statement.State == BankStatementStateCancelled {
		return nil, ErrStatementCancelled
	}

	lines, err := s.lines.ListUnreconciledByStatement(ctx, statementID)
	if err != nil {
		return nil, err
	}

	unreconciled := make([]*BankStatementLine, 0, len(lines))
	for _, line := range lines {
		if line.ContactID == nil {
			unreconciled = append(unreconciled, line)
			continue
		}
		payments, err := s.payments.ListPostedByContact(ctx, *line.ContactID)
		if err != nil {
			return nil, err
		}
		matched := false
		gross := StatementGross(line)
		lineAbs := gross.Abs().Round(4)
		tolerance := amount.FromFloat64(0.02).Round(4)
		if !gross.IsZero() {
			tol := gross.Abs().Mul(amount.FromFloat64(0.01)).Round(4)
			if tol.GreaterThan(tolerance) {
				tolerance = tol
			}
		}
		for _, payment := range payments {
			if line.CurrencyCode != nil && payment.CurrencyCode != nil && *line.CurrencyCode != *payment.CurrencyCode {
				continue
			}
			payAmt := amount.FromFloat64(payment.Amount).Abs().Round(4)
			diff := payAmt.Sub(lineAbs).Abs()
			if diff.GreaterThan(tolerance) && !payAmt.Equal(lineAbs) {
				continue
			}
			if line.Date != nil {
				diffHours := payment.Date.Sub(*line.Date).Hours()
				if diffHours > 72 || diffHours < -72 {
					continue
				}
			}
			if payment.EntryID == nil {
				continue
			}
			err = s.tx.Run(ctx, func(tx *gorm.DB) error {
				line.PaymentID = helper.Ptr(payment.ID)
				line.Reconciled = true
				if err := s.postFeeInterestTx(ctx, tx, statement, line); err != nil {
					return err
				}
				_, err := s.lines.UpdateTx(ctx, tx, line)
				return err
			})
			if err != nil {
				return nil, err
			}
			matched = true
			break
		}
		if !matched && line.Narration != nil {
			for _, payment := range payments {
				if payment.Reference != nil && *payment.Reference != "" && line.Ref != nil && *payment.Reference == *line.Ref {
					err = s.tx.Run(ctx, func(tx *gorm.DB) error {
						line.PaymentID = helper.Ptr(payment.ID)
						line.Reconciled = true
						if err := s.postFeeInterestTx(ctx, tx, statement, line); err != nil {
							return err
						}
						_, err := s.lines.UpdateTx(ctx, tx, line)
						return err
					})
					if err != nil {
						return nil, err
					}
					matched = true
					break
				}
			}
		}
		if !matched {
			unreconciled = append(unreconciled, line)
		}
	}

	if len(unreconciled) == 0 && len(lines) > 0 {
		statement.State = BankStatementStateReconciled
		if _, err := s.statements.Update(ctx, statement); err != nil {
			return nil, err
		}
	}

	return unreconciled, nil
}

func (s BankStatementService) postFeeInterestTx(ctx context.Context, tx *gorm.DB, statement *BankStatement, line *BankStatementLine) error {
	if s.poster == nil || s.charges == nil {
		return nil
	}
	if statement.JournalID == nil {
		return nil
	}
	journal, err := s.journals.Find(ctx, *statement.JournalID)
	if err != nil || journal == nil {
		return err
	}
	bankAccount := journal.DefaultAccountID
	if bankAccount == nil {
		bankAccount = journal.BankAccountID
	}
	if bankAccount == nil {
		return nil
	}
	date := s.now().UTC()
	if line.Date != nil && !line.Date.IsZero() {
		date = *line.Date
	} else if statement.Date != nil && !statement.Date.IsZero() {
		date = *statement.Date
	}
	if !line.FeeAmount.IsZero() && line.FeeEntryID == nil {
		feeAccount, err := s.charges.FeeAccountID(ctx, journal.OrganizationID)
		if err != nil {
			return nil
		}
		fee := line.FeeAmount.Abs().Round(4)
		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: journal.OrganizationID,
			JournalID:      *statement.JournalID,
			Date:           date,
			Ref:            refOr("", line.Ref) + "/FEE",
			OriginType:     OriginTypeBankFee,
			OriginID:       line.ID,
			Description:    "Bank fee",
			Lines: []PostingLine{
				{AccountID: feeAccount, Name: "Bank Fee", Debit: fee},
				{AccountID: *bankAccount, Name: "Bank", Credit: fee},
			},
		})
		if err != nil {
			return err
		}
		line.FeeEntryID = helper.Ptr(entry.ID)
	}
	if !line.InterestAmount.IsZero() && line.InterestEntryID == nil {
		interestAccount, err := s.charges.InterestAccountID(ctx, journal.OrganizationID)
		if err != nil {
			return nil
		}
		interest := line.InterestAmount.Abs().Round(4)
		entry, err := s.poster.PostTx(ctx, tx, PostRequest{
			OrganizationID: journal.OrganizationID,
			JournalID:      *statement.JournalID,
			Date:           date,
			Ref:            refOr("", line.Ref) + "/INT",
			OriginType:     OriginTypeBankInterest,
			OriginID:       line.ID,
			Description:    "Bank interest",
			Lines: []PostingLine{
				{AccountID: *bankAccount, Name: "Bank", Debit: interest},
				{AccountID: interestAccount, Name: "Bank Interest", Credit: interest},
			},
		})
		if err != nil {
			return err
		}
		line.InterestEntryID = helper.Ptr(entry.ID)
	}
	return nil
}

func refOr(fallback string, ref *string) string {
	if ref != nil && *ref != "" {
		return *ref
	}
	return fallback
}
