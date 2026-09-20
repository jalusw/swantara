import { screen } from "@testing-library/react";
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

function seedForm(opportunities: unknown[] = []) {
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
  );
}

beforeEach(() => {
  seedForm();
});

describe("SaleOrderFormDialog extra", () => {
  it("shows the empty-lines hint after removing the only line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Remove" }));

    expect(await screen.findByText("Add at least one line.")).toBeInTheDocument();
  });

  it("requires a customer before saving", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(await screen.findByText("Select a customer.")).toBeInTheDocument();
  });

  it("explains when no won opportunities exist", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    const opportunityTrigger = screen.getByRole("combobox", { name: "Opportunity (won)" });
    await user.click(opportunityTrigger);

    expect(await screen.findByText("No won opportunities")).toBeInTheDocument();
  });

  it("updates the quantity of a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
    );

    await screen.findByRole("dialog");
    const qtyInput = screen.getByDisplayValue("1");
    await user.clear(qtyInput);
    await user.type(qtyInput, "4");

    expect(qtyInput).toHaveValue(4);
  });
});
