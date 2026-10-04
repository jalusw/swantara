import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SaleOrderFormDialog } from "../sale-order-form-dialog";

const contacts = [
  {
    id: 1,
    organization_id: 1,
    name: "Bluebird Trading Pte. Ltd.",
    display_name: "Bluebird Trading",
    is_organization: true,
    email: "billing@bluebird.sg",
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const price_books = [
  {
    id: 1,
    name: "Retail",
    currency_code: "USD",
    organization_id: 1,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const warehouses = [
  {
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

const wonOpportunities = [
  { id: 3, organization_id: 1, name: "Big Deal", is_won: true },
  { id: 4, organization_id: 1, name: "Lost Deal", is_won: false },
];

const variants = [{ id: 9, item_id: 5, name: "Red / L" }];

function seedForm(opportunities: unknown[] = wonOpportunities) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { price_books } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products/:itemId/variants", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { variants } }),
    ),
  );
}

async function fillRequired(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
  await user.click(await screen.findByRole("option", { name: "Bluebird Trading" }));
  await user.click(screen.getByRole("combobox", { name: "Peluang dimenangkan" }));
  await user.click(await screen.findByRole("option", { name: "Big Deal" }));
  await user.click(screen.getByRole("combobox", { name: "Item 1" }));
  await user.click(await screen.findByRole("option", { name: /Canvas Tote/ }));
}

beforeEach(() => {
  seedForm();
});

describe("SaleOrderFormDialog extra2", () => {
  it("creates a quotation and reports the new id", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders", () => {
        createCalls += 1;
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { order: { id: 12 } },
        });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
    );

    await screen.findByRole("dialog");
    await fillRequired(user);
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(createCalls).toBe(1));
    await waitFor(() => expect(onSave).toHaveBeenCalledWith("12"));
  });

  it("blocks saving when every line was removed", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders", () => {
        createCalls += 1;
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { order: { id: 12 } },
        });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("combobox", { name: "Pelanggan" }));
    await user.click(await screen.findByRole("option", { name: "Bluebird Trading" }));
    await user.click(screen.getByRole("combobox", { name: "Peluang dimenangkan" }));
    await user.click(await screen.findByRole("option", { name: "Big Deal" }));
    await user.click(screen.getByRole("button", { name: "Hapus baris" }));
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() =>
      expect(screen.queryByText("Pelanggan wajib dipilih.")).not.toBeInTheDocument(),
    );
    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("blocks saving when a line quantity is not positive", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/sale-orders", () => {
        createCalls += 1;
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { order: { id: 12 } },
        });
      }),
    );
    const onSave = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={onSave} />,
    );

    await screen.findByRole("dialog");
    await fillRequired(user);
    const qtyInput = screen.getAllByRole("spinbutton")[0]!;
    await user.clear(qtyInput);
    await user.type(qtyInput, "0");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(createCalls).toBe(0);
    expect(onSave).not.toHaveBeenCalled();
  });

  it("loads variants after a item is picked", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("combobox", { name: "Item 1" }));
    await user.click(await screen.findByRole("option", { name: /Canvas Tote/ }));

    expect(await screen.findByText("1 varian — menggunakan harga templat")).toBeInTheDocument();
  });

  it("shows only won opportunities in the opportunity picker", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("combobox", { name: "Peluang dimenangkan" }));

    expect(await screen.findByRole("option", { name: "Big Deal" })).toBeInTheDocument();
    expect(screen.queryByRole("option", { name: "Lost Deal" })).not.toBeInTheDocument();
  });

  it("preselects the initial opportunity when provided", async () => {
    renderWithProviders(
      <SaleOrderFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        initialOpportunityId="3"
        onSave={vi.fn()}
      />,
    );

    await screen.findByRole("dialog");
    expect(await screen.findByText("Big Deal")).toBeInTheDocument();
  });

  it("updates discount and description of a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    const discountInput = screen.getByDisplayValue("0");
    await user.clear(discountInput);
    await user.type(discountInput, "10");
    expect(discountInput).toHaveValue(10);

    const descriptionInput = screen.getByPlaceholderText("Deskripsi");
    await user.type(descriptionInput, "Gift wrap");
    expect(descriptionInput).toHaveValue("Gift wrap");
  });
});
