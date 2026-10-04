import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TaxesSection } from "../taxes-section";

const STAMP = "2026-01-01T00:00:00Z";

function tax(id: number, patch: Record<string, unknown> = {}) {
  return {
    id,
    organization_id: 1,
    name: `Tax ${id}`,
    amount: 11,
    type: "percent",
    scope: "sale",
    price_include: false,
    tax_account_id: null,
    refund_tax_account_id: null,
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
    ...patch,
  };
}

const accounts = [
  { id: 10, organization_id: 1, code: "2100", name: "Tax Payable", type: "liability" },
];

function seedTaxes(taxes: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/taxes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { taxes } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
  );
}

beforeEach(() => {
  seedTaxes([
    tax(1, { name: "PPN 11%" }),
    tax(2, { name: "Fixed Fee", type: "fixed", scope: "purchase", price_include: true }),
  ]);
});

describe("TaxesSection extra2", () => {
  it("renders amount, scope badges and price-include marks", async () => {
    renderWithProviders(<TaxesSection orgId="1" />);

    expect(await screen.findByText("PPN 11%")).toBeInTheDocument();
    expect(screen.getByText("Tetap")).toBeInTheDocument();
    expect(screen.getByText("Pembelian")).toBeInTheDocument();
    expect(screen.getByText("✓")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("shows the error state with retry", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/taxes", () =>
        HttpResponse.json({ success: false, message: "tax boom" }, { status: 500 }),
      ),
    );
    renderWithProviders(<TaxesSection orgId="1" />);

    expect(await screen.findByText("tax boom")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Coba lagi" })).toBeInTheDocument();
  });

  it("deletes a tax and refetches", async () => {
    let deleteCalls = 0;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/taxes/:id", () => {
        deleteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.click(screen.getAllByRole("button", { name: "Hapus" })[0]!);
    await user.click(await screen.findByRole("button", { name: "Hapus" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
  });

  it("creates a tax from the dialog", async () => {
    let createCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/taxes", () => {
        createCalls += 1;
        return HttpResponse.json({ success: true, message: "Dibuat.", data: {} }, { status: 201 });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.click(screen.getByRole("button", { name: "Tambah pajak" }));
    const dialog = await screen.findByRole("dialog");

    const nameInput = within(dialog).getAllByRole("textbox")[0]!;
    await user.type(nameInput, "New Tax");
    await user.click(within(dialog).getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(createCalls).toBe(1));
  });

  it("keeps save disabled without a name", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.click(screen.getByRole("button", { name: "Tambah pajak" }));
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByRole("button", { name: "Simpan" })).toBeDisabled();
  });

  it("updates an existing tax from the edit dialog", async () => {
    let updateCalls = 0;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/taxes/:id", () => {
        updateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("PPN 11%");
    await user.click(screen.getAllByRole("button", { name: "Ubah" })[0]!);

    expect(await screen.findByText("Ubah pajak")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(updateCalls).toBe(1));
  });

  it("disables amount for group taxes", async () => {
    seedTaxes([tax(3, { name: "Group Tax", type: "group", amount: null })]);
    const user = userEvent.setup();
    renderWithProviders(<TaxesSection orgId="1" />);

    await screen.findByText("Group Tax");
    await user.click(screen.getAllByRole("button", { name: "Ubah" })[0]!);
    const dialog = await screen.findByRole("dialog");

    expect(screen.getAllByText("Kelompok").length).toBeGreaterThan(1);
    const amountInput = within(dialog).getAllByRole("spinbutton")[0]!;
    expect(amountInput).toBeDisabled();
  });
});
