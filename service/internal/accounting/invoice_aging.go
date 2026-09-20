package accounting

import (
	"context"
	"sort"
	"time"
)

type AgingBucket string

const (
	AgingCurrent  AgingBucket = "current"
	AgingBucket30 AgingBucket = "days_1_30"
	AgingBucket60 AgingBucket = "days_31_60"
	AgingBucket90 AgingBucket = "days_61_90"
	AgingOver90   AgingBucket = "over_90_days"
)

func BucketFor(dueDate *time.Time, asOf time.Time) AgingBucket {
	if dueDate == nil || !dueDate.Before(asOf) {
		return AgingCurrent
	}
	days := int(asOf.Sub(*dueDate).Hours() / 24)
	switch {
	case days <= 30:
		return AgingBucket30
	case days <= 60:
		return AgingBucket60
	case days <= 90:
		return AgingBucket90
	default:
		return AgingOver90
	}
}

type AgingReportRow struct {
	ContactID uint64
	Current   float64
	Days1_30  float64
	Days31_60 float64
	Days61_90 float64
	Over90    float64
	Total     float64
}

func (s InvoiceService) AgingReport(ctx context.Context, organizationID uint64, asOf *time.Time) ([]AgingReportRow, error) {
	moment := s.now().UTC()
	if asOf != nil {
		moment = *asOf
	}
	invoices, err := s.invoices.ListOpenByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	byContact := map[uint64]*AgingReportRow{}
	for _, invoice := range invoices {
		row, ok := byContact[invoice.ContactID]
		if !ok {
			row = &AgingReportRow{ContactID: invoice.ContactID}
			byContact[invoice.ContactID] = row
		}
		residual := invoice.AmountResidual.Float64()
		switch BucketFor(invoice.DueDate, moment) {
		case AgingBucket30:
			row.Days1_30 += residual
		case AgingBucket60:
			row.Days31_60 += residual
		case AgingBucket90:
			row.Days61_90 += residual
		case AgingOver90:
			row.Over90 += residual
		default:
			row.Current += residual
		}
		row.Total += residual
	}
	rows := make([]AgingReportRow, 0, len(byContact))
	for _, row := range byContact {
		rows = append(rows, *row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ContactID < rows[j].ContactID })
	return rows, nil
}
