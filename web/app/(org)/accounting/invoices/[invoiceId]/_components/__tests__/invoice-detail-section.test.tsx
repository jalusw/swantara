import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InvoiceDetailSection } from "../invoice-detail-section";

function useLocalInvoice() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          invoice: {
            id: 7,
            organization_id: 1,
            entry_id: null,
            type: "customer_invoice",
            contact_id: 10,
            name: "INV-2026-007",
            reference: "PO-99",
            invoice_date: "2026-02-01",
            due_date: "2026-03-01",
            currency_code: null,
            journal_id: 1,
            payment_term_id: null,
            state: "posted",
            payment_state: "partial",
            amount_untaxed: 1000,
            amount_tax: 110,
            amount_total: 1110,
            amount_residual: 400,
            created_at: "2026-02-01T00:00:00Z",
            updated_at: "2026-02-01T00:00:00Z",
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalInvoice();
});

describe("InvoiceDetailSection", () => {
  it("renders invoice header with totals", async () => {
    renderWithProviders(<InvoiceDetailSection orgId="1" invoiceId="7" />);

    expect(await screen.findByText("INV-2026-007")).toBeInTheDocument();
    expect(screen.getByText("1,110.00")).toBeInTheDocument();
  });

  it("renders residual without duplicate back navigation", async () => {
    renderWithProviders(<InvoiceDetailSection orgId="1" invoiceId="7" />);

    await screen.findByText("INV-2026-007");
    expect(screen.queryByText("Back to invoices")).not.toBeInTheDocument();
    expect(screen.getByText("400.00")).toBeInTheDocument();
  });
});
