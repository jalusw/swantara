import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierCatalogSection } from "../supplier-catalog-section";

const supplierProducts = [
  {
    id: 1,
    item_id: 5,
    supplier_id: 9,
    vendor_sku: "VEN-TOTE-001",
    vendor_product_name: null,
    min_qty: 10,
    price: 32.5,
    currency_code: "USD",
    lead_time_days: 7,
    priority: 1,
    valid_from: null,
    valid_to: null,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    item_id: 6,
    supplier_id: 10,
    vendor_sku: "VEN-MUG-002",
    vendor_product_name: null,
    min_qty: 24,
    price: 8.75,
    currency_code: "USD",
    lead_time_days: 14,
    priority: 2,
    valid_from: null,
    valid_to: null,
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
  {
    id: 6,
    organization_id: 1,
    name: "Ceramic Mug",
    category_id: null,
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

function useCatalogHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { supplier_products: supplierProducts },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {
  useCatalogHandlers();
});

describe("SupplierCatalogSection", () => {
  it("renders supplier products with item names and supplier SKUs", async () => {
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("Canvas Tote")).toBeInTheDocument();
    expect(screen.getByText("Ceramic Mug")).toBeInTheDocument();
    expect(screen.getByText("VEN-TOTE-001")).toBeInTheDocument();
    expect(screen.getByText("VEN-MUG-002")).toBeInTheDocument();
    expect(screen.getByText("Pemasok #9")).toBeInTheDocument();
  });

  it("filters entries through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.type(screen.getByPlaceholderText("Cari produk pemasok"), "MUG-002");

    expect(await screen.findByText("VEN-MUG-002")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("VEN-TOTE-001")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Tambah produk pemasok" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Produk pemasok baru")).toBeInTheDocument();
  });
});
