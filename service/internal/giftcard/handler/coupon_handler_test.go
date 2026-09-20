package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"

	"gorm.io/gorm"
)

func TestCouponHandler_List(t *testing.T) {
	tests := []struct {
		name     string
		coupons  giftcard.CouponDAOMock
		noTenant bool
		path     string
		want     int
	}{
		{
			name: "returns coupons",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[giftcard.Coupon], error) {
						return &query.Page[giftcard.Coupon]{Items: []*giftcard.Coupon{sampleCoupon()}, Count: 1}, nil
					},
				},
			},
			path: "/coupons/",
			want: http.StatusOK,
		},
		{
			name:    "rejects invalid query",
			coupons: giftcard.CouponDAOMock{},
			path:    "/coupons/?filter=bogus:eq:x",
			want:    http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			coupons:  giftcard.CouponDAOMock{},
			noTenant: true,
			path:     "/coupons/",
			want:     http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[giftcard.Coupon], error) {
						return nil, errors.New("db down")
					},
				},
			},
			path: "/coupons/",
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			var app *fiber.App
			if tt.noTenant {
				app = couponHandlerTestNoTenant(t, tt.coupons, svc)
			} else {
				app = couponHandlerTest(t, tt.coupons, svc)
			}
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_Get(t *testing.T) {
	tests := []struct {
		name    string
		coupons giftcard.CouponDAOMock
		path    string
		want    int
	}{
		{
			name: "returns coupon",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusOK,
		},
		{
			name: "returns not found",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return nil, nil
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusNotFound,
		},
		{
			name: "returns not found for foreign tenant",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return &giftcard.Coupon{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusNotFound,
		},
		{
			name:    "rejects invalid id",
			coupons: giftcard.CouponDAOMock{},
			path:    "/coupons/abc",
			want:    http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return nil, errors.New("db down")
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			app := couponHandlerTest(t, tt.coupons, svc)
			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_Create(t *testing.T) {
	tests := []struct {
		name     string
		coupons  giftcard.CouponDAOMock
		noTenant bool
		body     string
		want     int
	}{
		{
			name: "returns created",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, nil
					},
					CreateFunc: func(_ context.Context, coupon *giftcard.Coupon) (*giftcard.Coupon, error) {
						coupon.ID = 1
						return coupon, nil
					},
				},
			},
			body: `{"code":"SAVE10","discount_type":"percent","discount_value":10}`,
			want: http.StatusCreated,
		},
		{
			name:    "rejects invalid body",
			coupons: giftcard.CouponDAOMock{},
			body:    `{"discount_type":"percent","discount_value":10}`,
			want:    http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			coupons:  giftcard.CouponDAOMock{},
			noTenant: true,
			body:     `{"code":"SAVE10","discount_type":"percent","discount_value":10}`,
			want:     http.StatusUnprocessableEntity,
		},
		{
			name:    "rejects invalid expiry date",
			coupons: giftcard.CouponDAOMock{},
			body:    `{"code":"SAVE10","discount_type":"percent","discount_value":10,"expiry_date":"bogus"}`,
			want:    http.StatusUnprocessableEntity,
		},
		{
			name: "rejects invalid discount type",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, nil
					},
				},
			},
			body: `{"code":"SAVE10","discount_type":"bogus","discount_value":10}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "rejects duplicate code",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
				},
			},
			body: `{"code":"SAVE10","discount_type":"percent","discount_value":10}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, errors.New("db down")
					},
				},
			},
			body: `{"code":"SAVE10","discount_type":"percent","discount_value":10}`,
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			var app *fiber.App
			if tt.noTenant {
				app = couponHandlerTestNoTenant(t, tt.coupons, svc)
			} else {
				app = couponHandlerTest(t, tt.coupons, svc)
			}
			resp, err := doRequest(app, http.MethodPost, "/coupons/", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_Update(t *testing.T) {
	tests := []struct {
		name    string
		coupons giftcard.CouponDAOMock
		path    string
		body    string
		want    int
	}{
		{
			name: "returns updated coupon",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, nil
					},
					UpdateFunc: func(_ context.Context, coupon *giftcard.Coupon) (*giftcard.Coupon, error) {
						return coupon, nil
					},
				},
			},
			path: "/coupons/1",
			body: `{"code":"SAVE20","discount_type":"fixed","discount_value":50}`,
			want: http.StatusOK,
		},
		{
			name:    "rejects invalid id",
			coupons: giftcard.CouponDAOMock{},
			path:    "/coupons/abc",
			body:    `{"code":"SAVE20","discount_type":"fixed","discount_value":50}`,
			want:    http.StatusUnprocessableEntity,
		},
		{
			name: "rejects invalid body",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
				},
			},
			path: "/coupons/1",
			body: `{"discount_type":"fixed","discount_value":50}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return nil, nil
					},
				},
			},
			path: "/coupons/1",
			body: `{"code":"SAVE20","discount_type":"fixed","discount_value":50}`,
			want: http.StatusNotFound,
		},
		{
			name: "rejects invalid expiry date",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
				},
			},
			path: "/coupons/1",
			body: `{"code":"SAVE20","discount_type":"fixed","discount_value":50,"expiry_date":"bogus"}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "rejects duplicate code",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return &giftcard.Coupon{Base: model.Base{ID: 2}, OrganizationID: helper.Ptr(uint64(10)), Code: "SAVE20"}, nil
					},
				},
			},
			path: "/coupons/1",
			body: `{"code":"SAVE20","discount_type":"fixed","discount_value":50}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return nil, errors.New("db down")
					},
				},
			},
			path: "/coupons/1",
			body: `{"code":"SAVE20","discount_type":"fixed","discount_value":50}`,
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			app := couponHandlerTest(t, tt.coupons, svc)
			resp, err := doRequest(app, http.MethodPut, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_Delete(t *testing.T) {
	tests := []struct {
		name    string
		coupons giftcard.CouponDAOMock
		path    string
		want    int
	}{
		{
			name: "returns no content",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
					DeleteFunc: func(_ context.Context, _ uint64) error {
						return nil
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusNoContent,
		},
		{
			name:    "rejects invalid id",
			coupons: giftcard.CouponDAOMock{},
			path:    "/coupons/abc",
			want:    http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return nil, nil
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusNotFound,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
					DeleteFunc: func(_ context.Context, _ uint64) error {
						return errors.New("db down")
					},
				},
			},
			path: "/coupons/1",
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			app := couponHandlerTest(t, tt.coupons, svc)
			resp, err := doRequest(app, http.MethodDelete, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_Redeem(t *testing.T) {
	tests := []struct {
		name     string
		coupons  giftcard.CouponDAOMock
		noTenant bool
		body     string
		want     int
	}{
		{
			name: "returns discount",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return sampleCoupon(), nil
					},
				},
				RedeemTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*giftcard.Coupon, error) {
					coupon := sampleCoupon()
					coupon.UsedCount = 1
					return coupon, nil
				},
			},
			body: `{"code":"SAVE10","subtotal":200}`,
			want: http.StatusOK,
		},
		{
			name:    "rejects invalid body",
			coupons: giftcard.CouponDAOMock{},
			body:    `{"subtotal":200}`,
			want:    http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			coupons:  giftcard.CouponDAOMock{},
			noTenant: true,
			body:     `{"code":"SAVE10","subtotal":200}`,
			want:     http.StatusUnprocessableEntity,
		},
		{
			name:    "rejects invalid date",
			coupons: giftcard.CouponDAOMock{},
			body:    `{"code":"SAVE10","subtotal":200,"date":"bogus"}`,
			want:    http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, nil
					},
				},
			},
			body: `{"code":"SAVE10","subtotal":200}`,
			want: http.StatusNotFound,
		},
		{
			name: "returns expired",
			coupons: func() giftcard.CouponDAOMock {
				expired := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return giftcard.CouponDAOMock{
					CRUDMock: dao.CRUDMock[giftcard.Coupon]{
						SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
							return &giftcard.Coupon{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), Code: "SAVE10", DiscountType: giftcard.CouponDiscountPercent, DiscountValue: 10, ExpiryDate: &expired}, nil
						},
					},
				}
			}(),
			body: `{"code":"SAVE10","subtotal":200,"date":"2026-03-01"}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "returns over limit",
			coupons: func() giftcard.CouponDAOMock {
				limit := 5
				return giftcard.CouponDAOMock{
					CRUDMock: dao.CRUDMock[giftcard.Coupon]{
						SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
							return &giftcard.Coupon{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), Code: "SAVE10", DiscountType: giftcard.CouponDiscountPercent, DiscountValue: 10, UsageLimit: &limit, UsedCount: limit}, nil
						},
					},
				}
			}(),
			body: `{"code":"SAVE10","subtotal":200}`,
			want: http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			coupons: giftcard.CouponDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.Coupon]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.Coupon, error) {
						return nil, errors.New("db down")
					},
				},
			},
			body: `{"code":"SAVE10","subtotal":200}`,
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewCouponService(tt.coupons, giftcard.TransactionerMock{})
			var app *fiber.App
			if tt.noTenant {
				app = couponHandlerTestNoTenant(t, tt.coupons, svc)
			} else {
				app = couponHandlerTest(t, tt.coupons, svc)
			}
			resp, err := doRequest(app, http.MethodPost, "/coupons/redeem", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestCouponHandler_WriteCouponError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: giftcard.ErrCouponNotFound, want: http.StatusNotFound},
		{name: "code required", err: giftcard.ErrCouponCodeRequired, want: http.StatusUnprocessableEntity},
		{name: "organization", err: giftcard.ErrCouponOrganization, want: http.StatusUnprocessableEntity},
		{name: "discount invalid", err: giftcard.ErrCouponDiscountInvalid, want: http.StatusUnprocessableEntity},
		{name: "already exists", err: giftcard.ErrCouponAlreadyExists, want: http.StatusUnprocessableEntity},
		{name: "expired", err: giftcard.ErrCouponExpired, want: http.StatusUnprocessableEntity},
		{name: "over limit", err: giftcard.ErrCouponOverLimit, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			got := writeErrorStatus(app, "/write-error", func(c fiber.Ctx) error {
				return writeCouponError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func couponHandlerTest(
	t *testing.T,
	coupons giftcard.CouponDAO,
	svc giftcard.CouponService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewCouponHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func couponHandlerTestNoTenant(
	t *testing.T,
	coupons giftcard.CouponDAO,
	svc giftcard.CouponService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewCouponHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleCoupon() *giftcard.Coupon {
	return &giftcard.Coupon{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Code:           "SAVE10",
		DiscountType:   giftcard.CouponDiscountPercent,
		DiscountValue:  10,
		UsedCount:      0,
	}
}
