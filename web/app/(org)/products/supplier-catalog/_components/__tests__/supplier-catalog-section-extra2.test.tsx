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
  valid_from: null,
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

function seedCatalog(items: unknown[] = [baseItem], fail = false) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-products", () => {
      if (fail) return HttpResponse.json({ success: false, message: "Boom." }, { status: 500 });
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { supplier_products: items },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {});

describe("SupplierCatalogSection extra2", () => {
  it("shows placeholders when price and lead time are missing", async () => {
    seedCatalog([{ ...baseItem, price: null, lead_time_days: null }]);
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the valid-to date when present", async () => {
    seedCatalog([{ ...baseItem, valid_to: "2026-12-31T00:00:00Z" }]);
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    expect(screen.getByText("31 Des 2026")).toBeInTheDocument();
  });

  it("falls back to the item id when the item is unknown", async () => {
    seedCatalog([{ ...baseItem, item_id: 999 }]);
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("#999")).toBeInTheDocument();
  });

  it("creates a supplier item from the dialog", async () => {
    seedCatalog([]);
    const createBodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/supplier-products", async ({ request }) => {
        createBodies.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "Tambah produk pemasok" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("combobox", { name: "Produk" }));
    await user.click(await screen.findByRole("option", { name: "Canvas Tote" }));
    await user.type(within(dialog).getByLabelText("Pemasok"), "9");
    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(createBodies).toHaveLength(1));
    expect(createBodies[0]).toMatchObject({ item_id: 5, supplier_id: 9 });
  });

  it("opens the edit dialog with the item locked", async () => {
    seedCatalog();
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Ubah produk pemasok" }));
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByText("Ubah produk pemasok")).toBeInTheDocument();
    expect(within(dialog).getByLabelText("Produk")).toHaveValue("Canvas Tote");
  });

  it("updates a supplier item from the edit dialog", async () => {
    seedCatalog();
    let updateCalls = 0;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/supplier-products/:id", () => {
        updateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Ubah produk pemasok" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(updateCalls).toBe(1));
  });

  it("deletes a supplier item after confirmation", async () => {
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
    const confirmDialog = await screen.findByRole("alertdialog");
    await user.click(within(confirmDialog).getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
  });

  it("retries after a load error", async () => {
    seedCatalog([], true);
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("Boom.")).toBeInTheDocument();
    seedCatalog();
    await user.click(screen.getByRole("button", { name: "Coba lagi" }));

    expect(await screen.findByText("VEN-TOTE-001")).toBeInTheDocument();
  });
});
