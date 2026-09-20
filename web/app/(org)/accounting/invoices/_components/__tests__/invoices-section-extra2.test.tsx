import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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
    type: "customer_invoice",
    contact_id: 11,
    name: "INV-2026-002",
    state: "cancelled",
    payment_state: "not_paid",
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

function seedInvoices() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { invoices } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals } }),
    ),
  );
}

beforeEach(() => {
  seedInvoices();
});

describe("InvoicesSection extra2", () => {
  it("falls back to generated name and dash for missing date", async () => {
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("INV-1")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry and refetches", async () => {
    let calls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/invoices", () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json({ success: false, message: "boom" }, { status: 500 });
        }
        return HttpResponse.json({ success: true, message: "OK.", data: { invoices } });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("boom")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("INV-2026-002")).toBeInTheDocument();
    expect(calls).toBe(2);
  });

  it("adds and removes invoice lines in the create dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-002");
    await user.click(screen.getByRole("button", { name: "New invoice" }));
    const dialog = await screen.findByRole("dialog");

    await user.click(within(dialog).getByRole("button", { name: "Add line" }));
    expect(within(dialog).getAllByRole("button", { name: "Remove" })).toHaveLength(2);

    await user.click(within(dialog).getAllByRole("button", { name: "Remove" })[0]!);
    expect(within(dialog).getAllByRole("button", { name: "Remove" })).toHaveLength(1);
  });

  it("keeps save disabled until contact and journal are set", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-002");
    await user.click(screen.getByRole("button", { name: "New invoice" }));
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByRole("button", { name: "Create" })).toBeDisabled();
  });

  it("creates an invoice and refreshes the list", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/invoices", () => {
        createCalls += 1;
        return HttpResponse.json(
          { success: true, message: "Created.", data: { invoice: { id: 9 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-002");
    await user.click(screen.getByRole("button", { name: "New invoice" }));
    const dialog = await screen.findByRole("dialog");

    const spinners = within(dialog).getAllByRole("spinbutton");
    await user.type(spinners[0]!, "10");
    await user.click(within(dialog).getByRole("combobox", { name: "Journal" }));
    await user.click(await screen.findByRole("option", { name: "Sales Journal" }));

    const save = within(dialog).getByRole("button", { name: "Create" });
    await waitFor(() => expect(save).toBeEnabled());
    await user.click(save);

    await waitFor(() => expect(createCalls).toBe(1));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("closes the dialog from the cancel button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-002");
    await user.click(screen.getByRole("button", { name: "New invoice" }));
    const dialog = await screen.findByRole("dialog");

    await user.click(within(dialog).getByRole("button", { name: "Cancel" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });
});
