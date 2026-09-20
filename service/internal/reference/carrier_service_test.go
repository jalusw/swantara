package reference

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestCarrierService_Create_RejectsEmptyName(t *testing.T) {
	ctx := context.Background()
	db, mock := query.NewMockDB(t)

	svc := NewCarrierService(dao.NewBase[Carrier](db), variantExistsStub{exists: true})

	_, err := svc.Create(ctx, &Carrier{Name: "  "})
	if helper.AssertError(t, err, true, ErrCarrierName) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCarrierService_Create_RejectsUnknownDeliveryProduct(t *testing.T) {
	ctx := context.Background()
	db, mock := query.NewMockDB(t)
	deliveryItemID := uint64(7)

	svc := NewCarrierService(dao.NewBase[Carrier](db), variantExistsStub{exists: false})

	_, err := svc.Create(ctx, &Carrier{Name: "DHL", DeliveryItemID: &deliveryItemID})
	if helper.AssertError(t, err, true, ErrCarrierDeliveryProduct) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCarrierService_Create_CreatesCarrier(t *testing.T) {
	ctx := context.Background()
	db, mock := query.NewMockDB(t)
	tracking := "https://track.example.com/{tracking_number}"
	deliveryItemID := uint64(7)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "carriers"`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "DHL", tracking, deliveryItemID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	svc := NewCarrierService(dao.NewBase[Carrier](db), variantExistsStub{exists: true})

	created, err := svc.Create(ctx, &Carrier{Name: "DHL", TrackingURLTpl: &tracking, DeliveryItemID: &deliveryItemID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("carrier id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}
