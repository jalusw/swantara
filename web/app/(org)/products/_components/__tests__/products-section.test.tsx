import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProductsSection } from "../products-section";

const products = [
  {
    id: 5,
    organization_id: 1,
    name: "Canvas Tote",
    category_id: 2,
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
  {
    id: 6,
    organization_id: 1,
    name: "Ceramic Mug",
    category_id: 2,
    type: "stockable",
    unit_id: null,
    purchase_unit_id: null,
    list_price: 12,
    standard_cost: 6,
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

const categories = [
  {
    id: 2,
    organization_id: 1,
    name: "Merchandise",
    parent_id: null,
    income_account_id: null,
    expense_account_id: null,
    stock_cost_account_id: null,
    stock_input_account_id: null,
    stock_output_account_id: null,
    cogs_account_id: null,
    cost_method: null,
    valuation: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useProductHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories } }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { units: [] } }),
    ),
  );
}

beforeEach(() => {
  useProductHandlers();
});

describe("ProductsSection", () => {
  it("renders products with category and type", async () => {
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(await screen.findByText("Canvas Tote")).toBeInTheDocument();
    expect(screen.getByText("Ceramic Mug")).toBeInTheDocument();
    expect(screen.getAllByText("Merchandise").length).toBeGreaterThan(0);
    expect(screen.getAllByText("Dapat distok").length).toBeGreaterThan(0);
  });

  it("filters products through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductsSection orgId="1" />);

    await screen.findByText("Canvas Tote");
    await user.type(screen.getByPlaceholderText("Cari produk"), "Mug");

    expect(await screen.findByText("Ceramic Mug")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Canvas Tote")).not.toBeInTheDocument());
  });

  it("opens the item dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductsSection orgId="1" />);

    await screen.findByText("Canvas Tote");
    await user.click(screen.getByRole("button", { name: "Tambah produk" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Produk baru")).toBeInTheDocument();
  });
});
