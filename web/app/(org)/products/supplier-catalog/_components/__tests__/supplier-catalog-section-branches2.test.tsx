import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierCatalogSection } from "../supplier-catalog-section";

const baseItem = {
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
  valid_from: "2026-01-01T00:00:00Z",
  valid_to: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

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

function seedCatalog(items: unknown[] = [baseItem]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { supplier_products: items },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {});

describe("SupplierCatalogSection branches2", () => {
  it("renders supplier sku fallback and null valid-to branches", async () => {
    seedCatalog([{ ...baseItem, vendor_sku: null, valid_to: null, valid_from: null }]);
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("Canvas Tote")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("updates an existing entry through the edit branch", async () => {
    seedCatalog();
    let updateBody: unknown = null;
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/supplier-products/:id",
        async ({ request }) => {
          updateBody = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Ubah produk pemasok" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(updateBody).not.toBeNull());
  });

  it("deletes an entry through the delete branch", async () => {
    seedCatalog();
    let deleteCalls = 0;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/supplier-products/:id", () => {
        deleteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Hapus" }));
    await user.click(await screen.findByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
  });

  it("shows create title branch for the new dialog", async () => {
    seedCatalog([]);
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "Tambah produk pemasok" }));
    expect(await screen.findByText("Produk pemasok baru")).toBeInTheDocument();
  });
});
