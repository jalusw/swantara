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

function useFormHandlers() {
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
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
}

function renderDialog() {
  return renderWithProviders(
    <SaleOrderFormDialog open onOpenChange={() => {}} orgId="1" onSave={vi.fn()} />,
  );
}

beforeEach(() => {
  useFormHandlers();
});

describe("SaleOrderFormDialog", () => {
  it("renders the quotation form with order lines", async () => {
    renderDialog();

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Order lines")).toBeInTheDocument();
    expect(screen.getAllByText("Select item").length).toBeGreaterThan(0);
  });

  it("adds another line when requested", async () => {
    const user = userEvent.setup();
    renderDialog();

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Add line" }));

    expect(screen.getAllByText("Select item")).toHaveLength(2);
  });

  it("removes a line when requested", async () => {
    const user = userEvent.setup();
    renderDialog();

    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Add line" }));
    expect(screen.getAllByText("Select item")).toHaveLength(2);
    await user.click(screen.getAllByRole("button", { name: "Remove" })[0]!);

    expect(screen.getAllByText("Select item")).toHaveLength(1);
  });
});
