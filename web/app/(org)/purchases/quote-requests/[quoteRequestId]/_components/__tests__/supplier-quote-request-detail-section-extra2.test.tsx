import { screen, waitFor } from "@testing-library/react";
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

describe("SupplierQuoteRequestDetail extra2", () => {
  it("sends a draft QuoteRequest to the supplier", async () => {
    seedQuoteRequest();
    let sendCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/send", () => {
        sendCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("button", { name: "Send to supplier" }));

    await waitFor(() => expect(sendCalls).toBe(1));
  });

  it("cancels a draft QuoteRequest", async () => {
    seedQuoteRequest();
    let cancelCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/cancel", () => {
        cancelCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    await waitFor(() => expect(cancelCalls).toBe(1));
  });

  it("accepts a submitted quote from the quotes tab", async () => {
    seedQuoteRequest();
    let acceptCalls = 0;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/quotes/:quoteId/accept",
        () => {
          acceptCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("tab", { name: "Quotes" }));
    await user.click(await screen.findByRole("button", { name: "Accept" }));

    await waitFor(() => expect(acceptCalls).toBe(1));
  });

  it("creates a purchase order through the confirm dialog", async () => {
    seedQuoteRequest({ quoteRequest: { state: "done" } });
    let poCalls = 0;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/supplier-quote-requests/:id/purchase-order",
        () => {
          poCalls += 1;
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("button", { name: "Create purchase order" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Create PO" }));

    await waitFor(() => expect(poCalls).toBe(1));
  });

  it("shows the empty-lines state when the QuoteRequest has no lines", async () => {
    seedQuoteRequest({ lines: [] });
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(await screen.findByText("No lines.")).toBeInTheDocument();
  });

  it("shows the empty-quotes state on the quotes tab", async () => {
    seedQuoteRequest({ quotes: [] });
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("tab", { name: "Quotes" }));

    expect(await screen.findByText("No quotes received yet.")).toBeInTheDocument();
  });

  it("shows the accepted-quote hint when a quote was accepted", async () => {
    seedQuoteRequest({ quotes: [{ ...baseQuote, state: "accepted" }] });
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    await user.click(screen.getByRole("tab", { name: "Quotes" }));

    expect(await screen.findByText(/Accepted quote total/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Accept" })).not.toBeInTheDocument();
  });

  it("shows the not-found state when the QuoteRequest is missing", async () => {
    seedQuoteRequest();
    server.use(
      http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    expect(await screen.findByText("QuoteRequest not found.")).toBeInTheDocument();
  });

  it("falls back to the supplier id when the supplier name is unknown", async () => {
    seedQuoteRequest();
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(screen.getByText("#7")).toBeInTheDocument();
  });

  it("omits the notes row when the QuoteRequest has no notes", async () => {
    seedQuoteRequest({ quoteRequest: { notes: null } });
    renderWithProviders(<SupplierQuoteRequestDetail orgId="1" quoteRequestId="4" />);

    await screen.findAllByText("QuoteRequest-0004");
    expect(screen.queryByText("Urgent restock")).not.toBeInTheDocument();
  });
});
