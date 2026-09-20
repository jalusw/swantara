import { screen } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InvoicesSection } from "../invoices-section";

const invoices = [
  {
    id: 1,
    type: "customer_invoice",
    contact_id: 10,
    name: null,
    state: "draft",
    payment_state: "not_paid",
    amount_untaxed: 1000,
    amount_tax: 110,
    amount_total: 1110,
    amount_residual: 1110,
    invoice_date: null,
  },
  {
    id: 2,
    type: "vendor_bill",
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
];

const journals = [
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
];

function seedInvoices(list: unknown[] = invoices) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { invoices: list } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
  );
}

beforeEach(() => {});

describe("InvoicesSection branches2", () => {
  it("renders type and state branches for mixed invoices", async () => {
    seedInvoices();
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("INV-1")).toBeInTheDocument();
    expect(screen.getByText("INV-2026-002")).toBeInTheDocument();
  });

  it("renders the empty branch when no invoices exist", async () => {
    seedInvoices([]);
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("No invoices found.")).toBeInTheDocument();
  });

  it("posts a draft invoice through the post branch", async () => {
    seedInvoices();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-1");
    expect(screen.getByText("INV-2026-002")).toBeInTheDocument();
    expect(screen.getByText("#10")).toBeInTheDocument();
    expect(screen.getByText("#11")).toBeInTheDocument();
  });

  it("shows the error branch with retry", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/invoices", () =>
        HttpResponse.json({ success: false, message: "Invoices down." }, { status: 500 }),
      ),
      http.get("*/api/v1/organizations/:organizationId/journals", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
      ),
    );
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("Invoices down.")).toBeInTheDocument();
  });
});
