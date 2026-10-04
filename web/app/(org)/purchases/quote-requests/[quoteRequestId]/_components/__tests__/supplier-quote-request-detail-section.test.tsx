import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestDetail } from "../supplier-quote-request-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const quoteRequest = {
  id: 4,
  organization_id: 1,
  name: "QuoteRequest-0004",
  requester_id: 9,
  supplier_id: 7,
  currency_code: "USD",
  state: "draft",
  order_date: "2026-02-01",
  quote_deadline: "2026-03-01",
  notes: "Urgent restock",
};

const supplier = {
  id: 7,
  organization_id: 1,
  name: "Acme Supplier Co",
  display_name: "Acme Supplier",
  is_organization: true,
  email: null,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const quotes = [
  {
    id: 41,
    quoteRequest_id: 4,
    supplier_id: 7,
    currency_code: "USD",
    state: "submitted",
    quote_date: "2026-02-10",
    valid_until: "2026-03-01",
    notes: null,
    amount_untaxed: 100,
    amount_tax: 10,
    amount_total: 110,
  },
];

const lines = [
  {
    id: 42,
    quoteRequest_id: 4,
    item_id: 5,
    description: "Widget",
    qty: 2,
    unit_id: null,
    needed_by: null,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { quoteRequest } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/quotes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { quotes } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
  );
});

describe("SupplierQuoteRequestDetail", () => {
  it("renders the QuoteRequest header with supplier", async () => {
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    expect((await screen.findAllByText("QuoteRequest-0004")).length).toBeGreaterThan(0);
    expect(screen.getByText("Acme Supplier")).toBeInTheDocument();
  });

  it("switches to the quotes tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("tab", { name: "Penawaran" }));

    expect(await screen.findByText("Penawaran pemasok")).toBeInTheDocument();
  });
});
