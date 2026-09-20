import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InvoicesSection } from "../invoices-section";

function useLocalInvoices() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          invoices: [
            {
              id: 1,
              type: "customer_invoice",
              contact_id: 10,
              name: "INV-2026-001",
              state: "draft",
              payment_state: "not_paid",
              amount_untaxed: 1000,
              amount_tax: 110,
              amount_total: 1110,
              amount_residual: 1110,
              invoice_date: "2026-02-01",
            },
            {
              id: 2,
              type: "customer_invoice",
              contact_id: 11,
              name: "INV-2026-002",
              state: "posted",
              payment_state: "paid",
              amount_untaxed: 2000,
              amount_tax: 220,
              amount_total: 2220,
              amount_residual: 0,
              invoice_date: "2026-02-05",
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          journals: [
            {
              id: 1,
              organization_id: 1,
              name: "Sales Journal",
              code: "SAJ",
              type: "sale",
              default_account_id: null,
              bank_account_id: null,
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalInvoices();
});

describe("InvoicesSection", () => {
  it("renders seeded invoices", async () => {
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("INV-2026-001")).toBeInTheDocument();
    expect(screen.getByText("INV-2026-002")).toBeInTheDocument();
  });

  it("filters invoices by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.type(screen.getByPlaceholderText("Search invoices..."), "002");

    expect(await screen.findByText("INV-2026-002")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.click(screen.getByRole("button", { name: "New invoice" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create invoice")).toBeInTheDocument();
  });
});
