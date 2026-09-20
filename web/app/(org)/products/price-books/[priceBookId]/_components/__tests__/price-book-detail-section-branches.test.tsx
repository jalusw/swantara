import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PriceBookDetail } from "../price-book-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const RULES = [
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
    date_start: "2026-01-01",
    date_end: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    price_book_id: 7,
    applies_to: "category",
    item_id: null,
    category_id: 3,
    min_qty: 1,
    compute_type: "percent",
    fixed_price: null,
    discount_pct: 5,
    date_start: null,
    date_end: "2026-12-31",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 3,
    price_book_id: 7,
    applies_to: "all",
    item_id: 99,
    category_id: 99,
    min_qty: 2,
    compute_type: "formula",
    fixed_price: null,
    discount_pct: null,
    date_start: null,
    date_end: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const PRODUCTS = [
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
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const CATEGORIES = [{ id: 3, organization_id: 1, name: "Bags" }];

function useHandlers(options?: { rules?: unknown[]; failRules?: boolean }) {
  const created: unknown[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/price_books/:priceBookId/rules", () => {
      if (options?.failRules) {
        return HttpResponse.json({ success: false, message: "Rules down." }, { status: 500 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { rules: options?.rules ?? RULES },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: PRODUCTS } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories: CATEGORIES } }),
    ),
    http.post(
      "*/api/v1/organizations/:organizationId/price_books/:priceBookId/rules",
      async ({ request }) => {
        created.push(await request.json());
        return HttpResponse.json(
          { success: true, message: "Created.", data: { rule: { id: 10 } } },
          { status: 201 },
        );
      },
    ),
  );
  return { created };
}

beforeEach(() => {});

describe("PriceBookDetail branches", () => {
  it("renders rules with resolved item and category names", async () => {
    useHandlers();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    expect(await screen.findByText("Canvas Tote")).toBeInTheDocument();
    expect(screen.getByText("Bags")).toBeInTheDocument();
    expect(screen.getAllByText("#99").length).toBe(2);
  });

  it("renders fallback dashes and formatted numbers", async () => {
    useHandlers();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Canvas Tote");
    expect(screen.getByText("45.00")).toBeInTheDocument();
    expect(screen.getByText("5%")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry when rules fail", async () => {
    useHandlers({ failRules: true });
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    expect(await screen.findByText("Rules down.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("opens the create dialog", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Canvas Tote");
    await user.click(screen.getByRole("button", { name: "Add rule" }));
    expect(await screen.findByText("New pricing rule")).toBeInTheDocument();
  });

  it("creates a rule with empty optionals mapped to null", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Canvas Tote");
    await user.click(screen.getByRole("button", { name: "Add rule" }));
    await screen.findByText("New pricing rule");
    await user.click(screen.getByRole("button", { name: "Save rule" }));
    await waitFor(() => expect(created.length).toBe(1));
    const body = created[0] as Record<string, unknown>;
    expect(body.applies_to).toBe("all");
    expect(body.item_id).toBeNull();
    expect(body.category_id).toBeNull();
    expect(body.date_start).toBeNull();
    expect(body.date_end).toBeNull();
  });

  it("cancels the create dialog", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Canvas Tote");
    await user.click(screen.getByRole("button", { name: "Add rule" }));
    await screen.findByText("New pricing rule");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => expect(screen.queryByText("New pricing rule")).not.toBeInTheDocument());
  });
});
