import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderWithProviders, server } from "@/lib/tests";
import { SaleOrdersSection } from "../sale-orders-section";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
  useRouter: () => navigationMock,
  usePathname: () => "/",
}));

const orders = [
  {
    id: 1,
    organization_id: 1,
    name: "SO-0001",
    contact_id: 1,
    ship_address_id: null,
    bill_address_id: null,
    price_book_id: 1,
    currency_code: "USD",
    salesperson_id: null,
    sales_group_id: null,
    crm_prospect_id: null,
    warehouse_id: 1,
    state: "draft",
    order_date: "2026-02-01",
    expected_date: null,
    validity_date: null,
    payment_term_id: null,
    incoterm: null,
    customer_po_ref: null,
    amount_untaxed: 100,
    amount_tax: 10,
    amount_total: 110,
    invoice_status: "no",
    delivery_status: "pending",
    note: null,
    created_at: "2026-02-01T00:00:00Z",
    updated_at: "2026-02-01T00:00:00Z",
  },
  {
    id: 2,
    organization_id: 1,
    name: "SO-0002",
    contact_id: 2,
    ship_address_id: null,
    bill_address_id: null,
    price_book_id: 1,
    currency_code: "USD",
    salesperson_id: null,
    sales_group_id: null,
    crm_prospect_id: null,
    warehouse_id: 1,
    state: "confirmed",
    order_date: "2026-02-05",
    expected_date: null,
    validity_date: null,
    payment_term_id: null,
    incoterm: null,
    customer_po_ref: null,
    amount_untaxed: 200,
    amount_tax: 20,
    amount_total: 220,
    invoice_status: "to_invoice",
    delivery_status: "partial",
    note: null,
    created_at: "2026-02-05T00:00:00Z",
    updated_at: "2026-02-05T00:00:00Z",
  },
];

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
  {
    id: 2,
    organization_id: 1,
    name: "PT Nusantara Logistics",
    display_name: "Nusantara Logistics",
    is_organization: true,
    email: "finance@nusantaralog.id",
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useSaleOrdersHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/sale-orders", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { orders } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { price_books: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/crm/opportunities", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { opportunities: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
}

beforeEach(() => {
  useSaleOrdersHandlers();
});

describe("SaleOrdersSection", () => {
  it("renders orders with customer names and status badges", async () => {
    renderWithProviders(<SaleOrdersSection orgId="1" />);

    expect(await screen.findByText("SO-0001")).toBeInTheDocument();
    expect(screen.getByText("SO-0002")).toBeInTheDocument();
    expect(screen.getByText("Bluebird Trading")).toBeInTheDocument();
    expect(screen.getByText("Nusantara Logistics")).toBeInTheDocument();
    expect(screen.getByText("Draf")).toBeInTheDocument();
    expect(screen.getByText("Dikonfirmasi")).toBeInTheDocument();
  });

  it("filters orders through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SaleOrdersSection orgId="1" />);

    await screen.findByText("SO-0001");
    await user.type(screen.getByPlaceholderText("Cari pesanan…"), "SO-0002");

    expect((await screen.findAllByText("SO-0002")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("SO-0001")).not.toBeInTheDocument());
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SaleOrdersSection orgId="1" />);

    await screen.findByText("SO-0001");
    await user.click(screen.getByRole("button", { name: "Penawaran baru" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Baris pesanan")).toBeInTheDocument();
  });
});
