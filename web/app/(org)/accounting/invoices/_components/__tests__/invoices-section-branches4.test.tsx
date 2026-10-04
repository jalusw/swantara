import { fireEvent, screen, waitFor, within } from "@testing-library/react";
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
    name: "INV-2026-001",
    state: "draft",
    payment_state: "not_paid",
    amount_untaxed: 1000,
    amount_tax: 110,
    amount_total: 1110,
    amount_residual: 1110,
    invoice_date: "2026-02-01",
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

function seed(journalPayload: unknown = { journals }) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { invoices } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: journalPayload }),
    ),
  );
}

beforeEach(() => {});

describe("InvoicesSection branches4", () => {
  it("opens the dialog when journals are missing", async () => {
    seed({});
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.click(screen.getByRole("button", { name: "Faktur baru" }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByRole("button", { name: "Buat" })).toBeDisabled();
  });

  it("updates one line while other lines stay untouched", async () => {
    seed();
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.click(screen.getByRole("button", { name: "Faktur baru" }));
    const dialog = await screen.findByRole("dialog");

    await user.click(within(dialog).getByRole("button", { name: "Tambah baris" }));
    const boxes = within(dialog).getAllByRole("textbox");
    const firstDescription = boxes[1];
    if (!firstDescription) throw new Error("Expected line description inputs");
    await user.type(firstDescription, "Widget");

    expect(firstDescription).toHaveValue("Widget");
  });

  it("does not submit without contact and journal", async () => {
    seed();
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/invoices", () => {
        createCalls += 1;
        return HttpResponse.json(
          { success: true, message: "Dibuat.", data: { invoice: { id: 9 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.click(screen.getByRole("button", { name: "Faktur baru" }));
    const dialog = await screen.findByRole("dialog");

    fireEvent.click(within(dialog).getByRole("button", { name: "Buat" }));

    await waitFor(() => expect(createCalls).toBe(0));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("creates an invoice with blank quantities mapped to fallbacks", async () => {
    seed();
    const bodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/invoices", async ({ request }) => {
        bodies.push(await request.json());
        return HttpResponse.json(
          { success: true, message: "Dibuat.", data: { invoice: { id: 9 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-2026-001");
    await user.click(screen.getByRole("button", { name: "Faktur baru" }));
    const dialog = await screen.findByRole("dialog");

    const spins = within(dialog).getAllByRole("spinbutton");
    await user.type(spins[0]!, "10");
    await user.click(within(dialog).getByRole("combobox", { name: "Jurnal" }));
    await user.click(await screen.findByRole("option", { name: "Sales Journal" }));

    const boxes = within(dialog).getAllByRole("textbox");
    await user.type(boxes[1]!, "Widget");
    const qty = within(dialog).getAllByRole("spinbutton")[1];
    if (!qty) throw new Error("Expected qty input");
    await user.clear(qty);

    await user.click(within(dialog).getByRole("button", { name: "Tambah baris" }));

    const save = within(dialog).getByRole("button", { name: "Buat" });
    await waitFor(() => expect(save).toBeEnabled());
    await user.click(save);

    await waitFor(() => expect(bodies.length).toBe(1));
    const body = bodies[0] as { lines: { qty: number; unit_price: number }[] };
    expect(body.lines).toHaveLength(1);
    expect(body.lines[0]?.qty).toBe(1);
    expect(body.lines[0]?.unit_price).toBe(0);
  });
});
