import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PriceBookDetail } from "../price-book-detail-section";

const rules = [
  {
    id: 1,
    price_book_id: 7,
    applies_to: "item",
    item_id: 5,
    category_id: null,
    min_qty: 10,
    compute_type: "fixed",
    fixed_price: 45,
    discount_pct: null,
    date_start: null,
    date_end: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    price_book_id: 7,
    applies_to: "all",
    item_id: null,
    category_id: null,
    min_qty: 1,
    compute_type: "percent",
    fixed_price: null,
    discount_pct: 5,
    date_start: null,
    date_end: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const products = [
  {
    id: 5,
    organization_id: 1,
    name: "Canvas Tote",
    category_id: null,
    type: "stockable",
    unit_id: null,
    purchase_unit_id: null,
    list_price: 50,
    standard_cost: 30,
    is_purchasable: true,
    is_sellable: true,
    is_manufactured: false,
    tracking: "none",
    weight: 0,
    volume: 0,
    hs_code: null,
    description_sale: null,
    description_purchase: null,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useRuleHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/price_books/:priceBookId/rules", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rules } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories: [] } }),
    ),
  );
}

beforeEach(() => {
  useRuleHandlers();
});

describe("PriceBookDetail", () => {
  it("renders rules with resolved item names", async () => {
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);

    expect(await screen.findByText("Canvas Tote")).toBeInTheDocument();
    expect(screen.getByText("45.00")).toBeInTheDocument();
    expect(screen.getByText("5%")).toBeInTheDocument();
    expect(screen.getByText("Fixed price")).toBeInTheDocument();
  });

  it("opens the rule dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);

    await screen.findByText("Canvas Tote");
    await user.click(screen.getByRole("button", { name: "Add rule" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("New pricing rule")).toBeInTheDocument();
  });
});
