import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestDetail } from "../supplier-quote-request-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function quoteRequestFixture(overrides = {}) {
  return {
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
    ...overrides,
  };
}

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

const quoteFixture = {
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
};

const lineFixture = {
  id: 42,
  quoteRequest_id: 4,
  item_id: 5,
  description: "Widget",
  qty: 2,
  unit_id: null,
  needed_by: null,
};

function useQuoteRequestHandlers(
  quoteRequest: Record<string, unknown>,
  opts?: { contacts?: unknown[]; quotes?: unknown[]; lines?: unknown[] },
) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { quoteRequest } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { contacts: opts?.contacts ?? [supplier] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/quotes", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { quotes: opts?.quotes ?? [quoteFixture] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/lines", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { lines: opts?.lines ?? [lineFixture] },
      }),
    ),
  );
}

beforeEach(() => {});

describe("SupplierQuoteRequestDetail branches3", () => {
  it("renders cancelled quoteRequests without send actions", async () => {
    useQuoteRequestHandlers(quoteRequestFixture({ state: "cancelled" }));
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    expect((await screen.findAllByText("QuoteRequest-0004")).length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: "Send" })).toBeNull();
  });

  it("falls back to generated names and contact names", async () => {
    useQuoteRequestHandlers(quoteRequestFixture({ name: null, supplier_id: null }), {
      contacts: [{ ...supplier, display_name: null }],
    });
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    expect((await screen.findAllByText("QuoteRequest-4")).length).toBeGreaterThan(0);
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("falls back for sparse lines and quotes", async () => {
    useQuoteRequestHandlers(quoteRequestFixture(), {
      lines: [{ ...lineFixture, item_id: null, description: null, needed_by: "2026-04-01" }],
      quotes: [{ ...quoteFixture, valid_until: null }],
    });
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
    expect(screen.getByText("1 Apr 2026")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Penawaran" }));

    expect(await screen.findByText("Penawaran pemasok")).toBeInTheDocument();
  });
});
