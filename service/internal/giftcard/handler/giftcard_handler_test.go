package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

type giftCardTestCase struct {
	name     string
	cards    giftcard.GiftCardDAOMock
	trans    giftcard.GiftCardTransactionDAOMock
	method   string
	path     string
	body     string
	noTenant bool
	want     int
}

func runGiftCardTests(t *testing.T, tests []giftCardTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := giftcard.NewGiftCardService(tt.cards, tt.trans, giftcard.PosterMock{}, giftcard.TransactionerMock{})
			var app *fiber.App
			if tt.noTenant {
				app = giftCardHandlerTestNoTenant(t, tt.cards, svc)
			} else {
				app = giftCardHandlerTest(t, tt.cards, svc)
			}
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestGiftCardHandler_List(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns gift cards",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[giftcard.GiftCard], error) {
						return &query.Page[giftcard.GiftCard]{Items: []*giftcard.GiftCard{sampleGiftCard()}, Count: 1}, nil
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/",
			body:   "",
			want:   http.StatusOK,
		},
		{
			name:   "rejects invalid query",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodGet,
			path:   "/gift-cards/?filter=bogus:eq:x",
			body:   "",
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			cards:    giftcard.GiftCardDAOMock{},
			method:   http.MethodGet,
			path:     "/gift-cards/",
			body:     "",
			noTenant: true,
			want:     http.StatusUnauthorized,
		},
		{
			name: "returns server error",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[giftcard.GiftCard], error) {
						return nil, errors.New("db down")
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/",
			body:   "",
			want:   http.StatusInternalServerError,
		},
	})
}

func TestGiftCardHandler_Get(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns gift card",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return sampleGiftCard(), nil
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42",
			body:   "",
			want:   http.StatusOK,
		},
		{
			name: "returns not found",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, nil
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42",
			body:   "",
			want:   http.StatusNotFound,
		},
		{
			name: "returns not found for foreign tenant",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return &giftcard.GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(99))}, nil
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42",
			body:   "",
			want:   http.StatusNotFound,
		},
		{
			name: "returns server error",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, errors.New("db down")
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42",
			body:   "",
			want:   http.StatusInternalServerError,
		},
	})
}

func TestGiftCardHandler_Issue(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns created",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.GiftCard, error) {
						return nil, nil
					},
				},
				CreateTxFunc: func(_ context.Context, _ *gorm.DB, card *giftcard.GiftCard) (*giftcard.GiftCard, error) {
					card.ID = 42
					return card, nil
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"code":"GIFT-001","amount":100,"currency_code":"IDR","journal_id":3,"cash_account_id":200,"liability_account_id":300,"date":"2026-02-01"}`,
			want:   http.StatusCreated,
		},
		{
			name:   "rejects invalid body",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"amount":0,"currency_code":"IDR","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			cards:    giftcard.GiftCardDAOMock{},
			method:   http.MethodPost,
			path:     "/gift-cards/",
			body:     `{"amount":100,"currency_code":"IDR","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			noTenant: true,
			want:     http.StatusUnprocessableEntity,
		},
		{
			name:   "rejects invalid expiry date",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"amount":100,"currency_code":"IDR","expiry_date":"bogus","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:   "rejects invalid date",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"amount":100,"currency_code":"IDR","date":"bogus","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name: "rejects duplicate code",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.GiftCard, error) {
						return sampleGiftCard(), nil
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"code":"GIFT-001","amount":100,"currency_code":"IDR","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					SearchFunc: func(_ context.Context, _ string, _ any) (*giftcard.GiftCard, error) {
						return nil, errors.New("db down")
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/",
			body:   `{"code":"GIFT-001","amount":100,"currency_code":"IDR","journal_id":3,"cash_account_id":200,"liability_account_id":300}`,
			want:   http.StatusInternalServerError,
		},
	})
}

func TestGiftCardHandler_Redeem(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns redeemed card",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return &giftcard.GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: giftcard.GiftCardStateActive, Balance: 100}, nil
					},
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *giftcard.GiftCard) (*giftcard.GiftCard, error) {
					return card, nil
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":30,"journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusOK,
		},
		{
			name:   "rejects invalid body",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":0,"journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:   "rejects invalid date",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":30,"date":"bogus","journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, nil
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":30,"journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusNotFound,
		},
		{
			name: "returns insufficient balance",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return &giftcard.GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: giftcard.GiftCardStateActive, Balance: 20}, nil
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":30,"journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name: "returns server error",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, errors.New("db down")
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/redeem",
			body:   `{"amount":30,"journal_id":3,"revenue_account_id":400,"liability_account_id":300}`,
			want:   http.StatusInternalServerError,
		},
	})
}

func TestGiftCardHandler_Refund(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns refunded card",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return &giftcard.GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), State: giftcard.GiftCardStateUsed, Balance: 0}, nil
					},
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *giftcard.GiftCard) (*giftcard.GiftCard, error) {
					return card, nil
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/refund",
			body:   `{"amount":10,"journal_id":3,"refund_account_id":500,"liability_account_id":300}`,
			want:   http.StatusOK,
		},
		{
			name:   "rejects invalid body",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/42/refund",
			body:   `{"amount":0,"journal_id":3,"refund_account_id":500,"liability_account_id":300}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name: "returns not found",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, nil
					},
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/42/refund",
			body:   `{"amount":10,"journal_id":3,"refund_account_id":500,"liability_account_id":300}`,
			want:   http.StatusNotFound,
		},
	})
}

func TestGiftCardHandler_ForfeitExpired(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns forfeited cards",
			cards: giftcard.GiftCardDAOMock{
				ListDueForExpiryFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*giftcard.GiftCard, error) {
					return []*giftcard.GiftCard{sampleGiftCard()}, nil
				},
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, card *giftcard.GiftCard) (*giftcard.GiftCard, error) {
					return card, nil
				},
			},
			method: http.MethodPost,
			path:   "/gift-cards/forfeit-expired",
			body:   `{"journal_id":3,"liability_account_id":300,"income_account_id":600}`,
			want:   http.StatusOK,
		},
		{
			name:   "rejects invalid body",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/forfeit-expired",
			body:   `{"journal_id":0,"liability_account_id":300,"income_account_id":600}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:     "rejects missing tenant",
			cards:    giftcard.GiftCardDAOMock{},
			method:   http.MethodPost,
			path:     "/gift-cards/forfeit-expired",
			body:     `{"journal_id":3,"liability_account_id":300,"income_account_id":600}`,
			noTenant: true,
			want:     http.StatusUnprocessableEntity,
		},
		{
			name:   "rejects invalid as_of",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/forfeit-expired",
			body:   `{"as_of":"bogus","journal_id":3,"liability_account_id":300,"income_account_id":600}`,
			want:   http.StatusUnprocessableEntity,
		},
		{
			name:   "rejects invalid date",
			cards:  giftcard.GiftCardDAOMock{},
			method: http.MethodPost,
			path:   "/gift-cards/forfeit-expired",
			body:   `{"date":"bogus","journal_id":3,"liability_account_id":300,"income_account_id":600}`,
			want:   http.StatusUnprocessableEntity,
		},
	})
}

func TestGiftCardHandler_ListTransactions(t *testing.T) {
	runGiftCardTests(t, []giftCardTestCase{
		{
			name: "returns transactions",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return sampleGiftCard(), nil
					},
				},
			},
			trans: giftcard.GiftCardTransactionDAOMock{
				ListByGiftCardFunc: func(_ context.Context, _ uint64) ([]*giftcard.GiftCardTransaction, error) {
					return []*giftcard.GiftCardTransaction{{Base: model.Base{ID: 1}, GiftCardID: 42, Type: giftcard.TransactionIssue, Amount: 100}}, nil
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42/transactions",
			body:   "",
			want:   http.StatusOK,
		},
		{
			name: "returns not found",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, nil
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42/transactions",
			body:   "",
			want:   http.StatusNotFound,
		},
		{
			name: "returns server error",
			cards: giftcard.GiftCardDAOMock{
				CRUDMock: dao.CRUDMock[giftcard.GiftCard]{
					FindFunc: func(_ context.Context, _ uint64) (*giftcard.GiftCard, error) {
						return nil, errors.New("db down")
					},
				},
			},
			method: http.MethodGet,
			path:   "/gift-cards/42/transactions",
			body:   "",
			want:   http.StatusInternalServerError,
		},
	})
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

func giftCardHandlerTest(
	t *testing.T,
	cards giftcard.GiftCardDAO,
	svc giftcard.GiftCardService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewGiftCardHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func giftCardHandlerTestNoTenant(
	t *testing.T,
	cards giftcard.GiftCardDAO,
	svc giftcard.GiftCardService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewGiftCardHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleGiftCard() *giftcard.GiftCard {
	return &giftcard.GiftCard{
		Base:              model.Base{ID: 42},
		OrganizationID:    helper.Ptr(uint64(10)),
		Code:              "GIFT-001",
		InitialAmount:     100,
		Balance:           100,
		CurrencyCode:      "IDR",
		State:             giftcard.GiftCardStateActive,
		IssuedFromOrderID: helper.Ptr(uint64(7)),
	}
}

func writeErrorStatus(app *fiber.App, path string, fn func(c fiber.Ctx) error) int {
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func TestGiftCardHandler_WriteGiftCardError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: giftcard.ErrGiftCardNotFound, want: http.StatusNotFound},
		{name: "code required", err: giftcard.ErrGiftCardCodeRequired, want: http.StatusUnprocessableEntity},
		{name: "amount invalid", err: giftcard.ErrGiftCardAmountInvalid, want: http.StatusUnprocessableEntity},
		{name: "inactive", err: giftcard.ErrGiftCardInactive, want: http.StatusUnprocessableEntity},
		{name: "expired", err: giftcard.ErrGiftCardExpired, want: http.StatusUnprocessableEntity},
		{name: "insufficient", err: giftcard.ErrGiftCardInsufficient, want: http.StatusUnprocessableEntity},
		{name: "accounts", err: giftcard.ErrGiftCardAccounts, want: http.StatusUnprocessableEntity},
		{name: "liability only", err: giftcard.ErrGiftCardLiabilityOnly, want: http.StatusUnprocessableEntity},
		{name: "revenue", err: giftcard.ErrGiftCardRevenue, want: http.StatusUnprocessableEntity},
		{name: "refund account", err: giftcard.ErrGiftCardRefundAccount, want: http.StatusUnprocessableEntity},
		{name: "forfeit account", err: giftcard.ErrGiftCardForfeitAccount, want: http.StatusUnprocessableEntity},
		{name: "already exists", err: giftcard.ErrGiftCardAlreadyExists, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			got := writeErrorStatus(app, "/write-error", func(c fiber.Ctx) error {
				return writeGiftCardError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
