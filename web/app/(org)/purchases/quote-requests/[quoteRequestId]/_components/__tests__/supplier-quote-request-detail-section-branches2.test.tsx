import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestDetail } from "../supplier-quote-request-detail-section";

const baseQuoteRequest = {
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

const baseQuote = {
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

const baseLine = {
  id: 42,
  quoteRequest_id: 4,
  item_id: 5,
  description: "Widget",
  qty: 2,
  unit_id: null,
  needed_by: null,
};

function seedQuoteRequest(overrides = {}) {
  const {
    quoteRequest: quoteRequestOverride = {},
    quotes = [baseQuote],
    lines = [baseLine],
  } = overrides as {
    quoteRequest?: Record<string, unknown>;
    quotes?: unknown[];
    lines?: unknown[];
  };
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { quoteRequest: { ...baseQuoteRequest, ...quoteRequestOverride } },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/quotes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { quotes } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
  );
}

beforeEach(() => {});

describe("SupplierQuoteRequestDetail branches2", () => {
  it("renders the not-found branch when the quoteRequest is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { quoteRequest: null } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/contacts", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/quotes", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { quotes: [] } }),
      ),
      http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/lines", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { lines: [] } }),
      ),
    );
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    expect(await screen.findByText("Permintaan penawaran tidak ditemukan.")).toBeInTheDocument();
  });

  it("hides notes and date fallbacks on both branches", async () => {
    seedQuoteRequest({ quoteRequest: { notes: null, order_date: null, quote_deadline: null } });
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders empty lines and empty quotes branches", async () => {
    seedQuoteRequest({ quotes: [], lines: [] });
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(await screen.findByText("Belum ada baris")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Penawaran" }));
    expect(await screen.findByText("Tidak ada penawaran")).toBeInTheDocument();
  });

  it("shows accepted quote branch when a quote is accepted", async () => {
    seedQuoteRequest({ quotes: [{ ...baseQuote, state: "accepted" }] });
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("tab", { name: "Penawaran" }));
    expect(await screen.findByText("Accepted")).toBeInTheDocument();
  });

  it("hides send actions for sent quoteRequests", async () => {
    seedQuoteRequest({ quoteRequest: { state: "sent" } });
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(screen.queryByRole("button", { name: "Send to supplier" })).not.toBeInTheDocument();
  });
});
