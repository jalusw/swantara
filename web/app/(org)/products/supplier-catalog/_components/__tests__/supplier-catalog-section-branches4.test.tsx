import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierCatalogSection } from "../supplier-catalog-section";

const STAMP = "2026-01-01T00:00:00Z";

const fullItem = {
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
  valid_from: "2026-01-05T00:00:00Z",
  valid_to: "2026-12-31T00:00:00Z",
  created_at: STAMP,
  updated_at: STAMP,
};

const sparseItem = {
  id: 2,
  item_id: 6,
  supplier_id: 10,
  vendor_sku: null,
  vendor_product_name: null,
  min_qty: 1,
  price: null,
  currency_code: null,
  lead_time_days: null,
  priority: 2,
  valid_from: null,
  valid_to: null,
  created_at: STAMP,
  updated_at: STAMP,
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
    created_at: STAMP,
    updated_at: STAMP,
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
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function seed(items: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { supplier_products: items } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

beforeEach(() => {});

describe("SupplierCatalogSection branches4", () => {
  it("edits an item with every optional field set", async () => {
    seed([fullItem]);
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("VEN-TOTE-001");
    await user.click(screen.getByRole("button", { name: "Edit supplier item" }));
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByLabelText("Supplier SKU")).toHaveValue("VEN-TOTE-001");
    expect(within(dialog).getByLabelText("Price")).toHaveValue(32.5);
  });

  it("edits an item with every optional field missing", async () => {
    seed([sparseItem]);
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("Ceramic Mug");
    const edits = screen.getAllByRole("button", { name: "Edit supplier item" });
    const edit = edits[edits.length - 1];
    if (!edit) throw new Error("Expected edit button");
    await user.click(edit);
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByLabelText("Supplier SKU")).toHaveValue("");
    expect(within(dialog).getByLabelText("Currency")).toHaveValue("USD");
  });

  it("creates a supplier item with dates and cleared numerics", async () => {
    seed([]);
    const bodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/supplier-products", async ({ request }) => {
        bodies.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "Add supplier item" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("combobox", { name: "Item" }));
    await user.click(await screen.findByRole("option", { name: "Canvas Tote" }));
    await user.type(within(dialog).getByLabelText("Supplier"), "9");
    await user.type(within(dialog).getByLabelText("Price"), "10");
    await user.type(within(dialog).getByLabelText("Lead time (days)"), "3");
    await user.clear(within(dialog).getByLabelText("Min. quantity"));
    await user.clear(within(dialog).getByLabelText("Currency"));
    await user.clear(within(dialog).getByLabelText("Priority"));

    const dates = dialog.querySelectorAll('input[type="date"]');
    if (dates.length !== 2) throw new Error("Expected date inputs");
    await user.type(dates[0] as HTMLElement, "2026-02-01");
    await user.type(dates[1] as HTMLElement, "2026-11-30");

    await user.click(within(dialog).getByRole("button", { name: "Save" }));

    await waitFor(() => expect(bodies).toHaveLength(1));
    expect(bodies[0]).toMatchObject({
      item_id: 5,
      supplier_id: 9,
      min_qty: 1,
      price: 10,
      currency_code: null,
      lead_time_days: 3,
      priority: 1,
    });
    const dated = bodies[0] as Record<string, unknown>;
    expect(dated.valid_from).toBeTruthy();
    expect(dated.valid_to).toBeTruthy();
  });
});
