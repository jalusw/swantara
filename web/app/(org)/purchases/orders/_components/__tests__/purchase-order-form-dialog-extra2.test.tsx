import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseOrderFormDialog } from "../purchase-order-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const supplier = {
  id: 7,
  organization_id: 1,
  name: "Acme Supplier Co",
  display_name: "Acme Supplier",
  is_organization: true,
  parent_id: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  tax_id: null,
  industry: null,
  currency_code: "USD",
  lang: "en",
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const warehouse = {
  id: 1,
  organization_id: 1,
  name: "Main Warehouse",
  code: "WH-01",
  line1: null,
  line2: null,
  city: null,
  state: null,
  postal_code: null,
  country_code: null,
  created_at: STAMP,
  updated_at: STAMP,
};

const item = {
  id: 5,
  organization_id: 1,
  name: "Finished Widget",
  category_id: null,
  type: "stockable",
  unit_id: null,
  purchase_unit_id: null,
  list_price: 100,
  standard_cost: 60,
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
};

function seedForm() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [warehouse] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
}

async function fillRequired(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("combobox", { name: "Pemasok" }));
  await user.click(await screen.findByRole("option", { name: "Acme Supplier" }));
  await user.click(screen.getByRole("combobox", { name: "Item 1" }));
  await user.click(await screen.findByRole("option", { name: "Finished Widget" }));
}

beforeEach(() => {
  seedForm();
});

describe("PurchaseOrderFormDialog extra2", () => {
  it("creates a purchase order and notifies on save", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await fillRequired(user);
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(createCalls).toBe(1));
    await waitFor(() => expect(onSave).toHaveBeenCalled());
  });

  it("blocks saving when every line was removed", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.click(screen.getByRole("combobox", { name: "Pemasok" }));
    await user.click(await screen.findByRole("option", { name: "Acme Supplier" }));
    await user.click(screen.getByRole("button", { name: "Hapus" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Tambahkan minimal satu baris.")).toBeInTheDocument();
    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks saving when a line quantity is not positive", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/purchase-orders", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await fillRequired(user);
    const qtyInput = screen.getAllByRole("spinbutton")[0]!;
    await user.clear(qtyInput);
    await user.type(qtyInput, "0");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("adds and removes order lines", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await user.click(screen.getByRole("button", { name: "Tambah baris" }));
    expect(screen.getByRole("combobox", { name: "Item 2" })).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    expect(screen.queryByRole("combobox", { name: "Item 2" })).not.toBeInTheDocument();
  });

  it("updates unit price, discount and description of a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const spins = screen.getAllByRole("spinbutton");
    await user.clear(spins[1]!);
    await user.type(spins[1]!, "42");
    expect(spins[1]).toHaveValue(42);

    const descriptionInput = screen.getByPlaceholderText("Deskripsi");
    await user.type(descriptionInput, "Fragile");
    expect(descriptionInput).toHaveValue("Fragile");
  });

  it("updates the supplier reference field", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <PurchaseOrderFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    const vendorRefInput = screen.getByLabelText("Referensi vendor");
    await user.type(vendorRefInput, "VEND-99");
    expect(vendorRefInput).toHaveValue("VEND-99");
  });
});
