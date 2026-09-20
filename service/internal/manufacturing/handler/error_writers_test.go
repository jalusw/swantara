package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

func errorRoute(writer func(fiber.Ctx, error) error, err error) *fiber.App {
	app := fiber.New()
	app.Get("/", func(c fiber.Ctx) error { return writer(c, err) })
	return app
}

func assertWriterStatus(t *testing.T, writer func(fiber.Ctx, error) error, err error, want int) {
	t.Helper()
	resp, respErr := errorRoute(writer, err).Test(httptest.NewRequest(http.MethodGet, "/", nil))
	if respErr != nil {
		t.Fatal(respErr)
	}
	helper.AssertStatus(t, resp.StatusCode, want)
}

func TestWriteProductionError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{manufacturing.ErrProductionOrderNotFound, http.StatusNotFound},
		{manufacturing.ErrProductionOrderState, http.StatusConflict},
		{manufacturing.ErrProductionOrderRecipe, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderOrganization, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrShopTaskNotFound, http.StatusNotFound},
		{manufacturing.ErrShopTaskWorkCenter, http.StatusUnprocessableEntity},
		{manufacturing.ErrShopTaskMismatch, http.StatusUnprocessableEntity},
		{manufacturing.ErrShopTaskLabor, http.StatusUnprocessableEntity},
		{manufacturing.ErrConsumeComponent, http.StatusUnprocessableEntity},
		{manufacturing.ErrConsumeQuantity, http.StatusUnprocessableEntity},
		{manufacturing.ErrProduceQuantity, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrWIPAccount, http.StatusUnprocessableEntity},
		{manufacturing.ErrLaborAccount, http.StatusUnprocessableEntity},
		{manufacturing.ErrVarianceAccount, http.StatusUnprocessableEntity},
		{sequence.ErrSequenceNotFound, http.StatusUnprocessableEntity},
		{manufacturing.ErrMovementNotFound, http.StatusNotFound},
		{manufacturing.ErrMovementState, http.StatusConflict},
		{manufacturing.ErrInsufficientStock, http.StatusConflict},
		{manufacturing.ErrNotConsumption, http.StatusUnprocessableEntity},
		{manufacturing.ErrNotProduction, http.StatusUnprocessableEntity},
		{manufacturing.ErrValuationAccount, http.StatusUnprocessableEntity},
		{manufacturing.ErrNegativeCost, http.StatusUnprocessableEntity},
		{manufacturing.ErrBatchRequired, http.StatusUnprocessableEntity},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		assertWriterStatus(t, writeProductionError, tc.err, tc.want)
	}
	if today, err := helper.ParseDateOrToday(""); err != nil || today.IsZero() {
		t.Error("ParseDateOrToday empty mismatch")
	}
	if valid, err := helper.ParseDateOrToday("2026-01-15"); err != nil || valid.Day() != 15 {
		t.Error("ParseDateOrToday valid mismatch")
	}
	if _, err := helper.ParseDateOrToday("nope"); err == nil {
		t.Error("ParseDateOrToday invalid mismatch")
	}
}

func TestWriteProductionOrderError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{manufacturing.ErrProductionOrderNotFound, http.StatusNotFound},
		{manufacturing.ErrProductionOrderState, http.StatusConflict},
		{manufacturing.ErrProductionOrderItem, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderRecipe, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderQty, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderOrganization, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrConsumedMaterial, http.StatusUnprocessableEntity},
		{sequence.ErrSequenceNotFound, http.StatusUnprocessableEntity},
		{manufacturing.ErrHoldOverflow, http.StatusConflict},
		{manufacturing.ErrBalanceNotFound, http.StatusConflict},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		assertWriterStatus(t, writeProductionOrderError, tc.err, tc.want)
	}
}

func TestWriteSubcontractError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{manufacturing.ErrOutsideProcessingNotFound, http.StatusNotFound},
		{manufacturing.ErrOutsideProcessingDuplicate, http.StatusConflict},
		{manufacturing.ErrOutsideProcessingState, http.StatusConflict},
		{manufacturing.ErrOutsideProcessingOrderState, http.StatusConflict},
		{manufacturing.ErrProductionOrderNotFound, http.StatusNotFound},
		{manufacturing.ErrProductionOrderRecipe, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingNotSubcontracted, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingComponents, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingSupplierLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingOperation, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingPurchaseOrder, http.StatusUnprocessableEntity},
		{manufacturing.ErrOutsideProcessingSupplier, http.StatusUnprocessableEntity},
		{manufacturing.ErrWIPAccount, http.StatusUnprocessableEntity},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		assertWriterStatus(t, writeSubcontractError, tc.err, tc.want)
	}
}

func TestWriteRecipeError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{manufacturing.ErrRecipeNotFound, http.StatusNotFound},
		{manufacturing.ErrRecipeItem, http.StatusUnprocessableEntity},
		{manufacturing.ErrInvalidRecipeType, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeComponent, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeLineQty, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeLineScrap, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeQty, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeOutputQty, http.StatusUnprocessableEntity},
		{manufacturing.ErrRecipeNoLines, http.StatusUnprocessableEntity},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		assertWriterStatus(t, writeRecipeError, tc.err, tc.want)
	}
}

func TestWriteMrpError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{manufacturing.ErrPlanningRunNotRequired, http.StatusUnprocessableEntity},
		{manufacturing.ErrPlanningItem, http.StatusUnprocessableEntity},
		{manufacturing.ErrPlanningNoRecipe, http.StatusUnprocessableEntity},
		{manufacturing.ErrPlanningPlannedType, http.StatusUnprocessableEntity},
		{manufacturing.ErrPlanningPlannedState, http.StatusConflict},
		{manufacturing.ErrPlanningSupplier, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderOrganization, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionLocation, http.StatusUnprocessableEntity},
		{manufacturing.ErrProductionOrderNotFound, http.StatusNotFound},
		{errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		assertWriterStatus(t, writeMrpError, tc.err, tc.want)
	}
}
