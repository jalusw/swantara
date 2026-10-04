package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type PaymentBatchService struct {
	batches    PaymentBatchDAO
	batchLines PaymentBatchLineDAO
	payments   PaymentDAO
	journals   JournalResolver
	tx         db.Transactioner
	now        func() time.Time
}

func NewPaymentBatchService(
	batches PaymentBatchDAO,
	batchLines PaymentBatchLineDAO,
	payments PaymentDAO,
	journals JournalResolver,
	tx db.Transactioner,
) PaymentBatchService {
	return PaymentBatchService{
		batches:    batches,
		batchLines: batchLines,
		payments:   payments,
		journals:   journals,
		tx:         tx,
		now:        time.Now,
	}
}

type CreatePaymentBatchRequest struct {
	OrganizationID uint64
	JournalID      uint64
	Name           string
	Date           time.Time
	PaymentIDs     []uint64
}

func (s PaymentBatchService) List(ctx context.Context, q *query.Query) (*query.Page[PaymentBatch], error) {
	return s.batches.List(ctx, q)
}

func (s PaymentBatchService) Create(ctx context.Context, request CreatePaymentBatchRequest) (*PaymentBatch, error) {
	if len(request.PaymentIDs) == 0 {
		return nil, ErrPaymentNoInvoices
	}

	var totalAmount float64
	batchLines := make([]*PaymentBatchLine, 0, len(request.PaymentIDs))

	var createdID uint64
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, paymentID := range request.PaymentIDs {
			payment, err := s.payments.Find(ctx, paymentID)
			if err != nil {
				return err
			}
			if payment == nil {
				return ErrPaymentNotFound
			}
			if payment.State != PaymentStatePosted {
				return ErrEntryNotPosted
			}
			totalAmount += payment.Amount
			batchLines = append(batchLines, &PaymentBatchLine{PaymentID: paymentID})
		}

		batchDate := request.Date
		if batchDate.IsZero() {
			batchDate = s.now().UTC()
		}

		batch := &PaymentBatch{
			OrganizationID: request.OrganizationID,
			Name:           helper.Ptr(request.Name),
			JournalID:      request.JournalID,
			TotalAmount:    totalAmount,
			PaymentCount:   len(batchLines),
			State:          PaymentBatchStateDraft,
			BatchDate:      &batchDate,
		}

		created, err := s.batches.CreateTx(ctx, tx, batch)
		if err != nil {
			return err
		}
		createdID = created.ID

		if err := s.batchLines.CreateBatchLinesTx(ctx, tx, batchLines, created.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	batch, err := s.batches.Find(ctx, createdID)
	if err != nil {
		return nil, err
	}
	return batch, nil
}

func (s PaymentBatchService) GenerateSEPA(ctx context.Context, batchID, organizationID uint64) (string, error) {
	batch, err := s.batches.Find(ctx, batchID)
	if err != nil {
		return "", err
	}
	if batch == nil || batch.OrganizationID != organizationID {
		return "", ErrEntryNotFound
	}
	if batch.State != PaymentBatchStateDraft {
		return "", ErrInvalidPeriodState
	}

	lines, err := s.batchLines.ListByBatch(ctx, batchID)
	if err != nil {
		return "", err
	}

	amounts := make(map[uint64]float64, len(lines))
	for _, line := range lines {
		payment, err := s.payments.Find(ctx, line.PaymentID)
		if err != nil {
			return "", err
		}
		if payment != nil {
			amounts[line.PaymentID] = payment.Amount
		}
	}

	xml := s.buildSEPAXML(batch, lines, amounts)

	now := s.now().UTC()
	batch.State = PaymentBatchStateConfirmed
	batch.GeneratedAt = &now
	if _, err := s.batches.Update(ctx, batch); err != nil {
		return "", err
	}

	return xml, nil
}

func (s PaymentBatchService) buildSEPAXML(batch *PaymentBatch, lines []*PaymentBatchLine, amounts map[uint64]float64) string {
	batchDate := s.now().UTC().Format("2006-01-02")
	if batch.BatchDate != nil && !batch.BatchDate.IsZero() {
		batchDate = batch.BatchDate.Format("2006-01-02")
	}
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pain.001.001.09">
  <CstmrCdtTrfInitn>
    <GrpHdr>
      <MsgId>` + fmt.Sprintf("BATCH/%d", batch.ID) + `</MsgId>
      <CreDtTm>` + s.now().UTC().Format("2006-01-02T15:04:05") + `</CreDtTm>
      <NbOfTxs>` + fmt.Sprintf("%d", batch.PaymentCount) + `</NbOfTxs>
      <CtrlSum>` + fmt.Sprintf("%.2f", batch.TotalAmount) + `</CtrlSum>
      <InitgPty><Nm>Swantara</Nm></InitgPty>
    </GrpHdr>
    <PmtInf>
      <PmtInfId>` + fmt.Sprintf("PMT/%d", batch.ID) + `</PmtInfId>
      <PmtMtd>TRF</PmtMtd>
      <ReqdExctDt><Dt>` + batchDate + `</Dt></ReqdExctDt>
`

	for _, line := range lines {
		xml += `      <CdtTrfTxInf>
        <PmtId><EndToEndId>` + fmt.Sprintf("TX/%d", line.PaymentID) + `</EndToEndId></PmtId>
        <Amt><InstdAmt Ccy="IDR">` + fmt.Sprintf("%.2f", amounts[line.PaymentID]) + `</InstdAmt></Amt>
        <CdtrAgt><FinInstnId><ClrSysMmbId><MmbId>` + fmt.Sprintf("%d", batch.JournalID) + `</MmbId></ClrSysMmbId></FinInstnId></CdtrAgt>
      </CdtTrfTxInf>
`
	}

	xml += `    </PmtInf>
  </CstmrCdtTrfInitn>
</Document>`
	return xml
}

func (s PaymentBatchService) Get(ctx context.Context, batchID, organizationID uint64) (*PaymentBatch, error) {
	batch, err := s.batches.Find(ctx, batchID)
	if err != nil {
		return nil, err
	}
	if batch == nil || batch.OrganizationID != organizationID {
		return nil, ErrEntryNotFound
	}
	return batch, nil
}

func (s PaymentBatchService) ListPayments(ctx context.Context, batchID uint64) ([]*PaymentBatchLine, error) {
	return s.batchLines.ListByBatch(ctx, batchID)
}
