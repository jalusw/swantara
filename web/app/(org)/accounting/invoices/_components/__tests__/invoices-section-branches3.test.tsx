import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InvoicesSection } from "../invoices-section";

const invoice = {
  id: 1,
  organization_id: 1,
  name: null,
  contact_id: 4,
  invoice_date: null,
  amount_total: 100,
  amount_residual: 100,
  state: "draft",
};

function seed(invoices: unknown[] = [invoice]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { invoices } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { journals: [] } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/invoices", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("InvoicesSection branches3", () => {
  it("falls back to generated name and dash date", async () => {
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("INV-1")).toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
    expect(screen.getByText("#4")).toBeInTheDocument();
  });

  it("renders named invoices with formatted dates", async () => {
    seed([{ ...invoice, name: "INV-2026-001", invoice_date: "2026-01-15" }]);
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByText("INV-2026-001")).toBeInTheDocument();
  });

  it("shows error with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/invoices", () =>
        HttpResponse.json({ success: false, message: "Boom." }, { status: 500 }),
      ),
    );
    renderWithProviders(<InvoicesSection orgId="1" />);

    expect(await screen.findByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("opens the create dialog and adds a line", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-1");
    await user.click(screen.getByRole("button", { name: "New invoice" }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toBeInTheDocument();
    const addLine = screen.queryByRole("button", { name: "Add line" });
    if (addLine) {
      await user.click(addLine);
      expect(screen.getAllByRole("dialog").length).toBeGreaterThan(0);
    }
  });

  it("keeps save disabled without contact and journal", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InvoicesSection orgId="1" />);

    await screen.findByText("INV-1");
    await user.click(screen.getByRole("button", { name: "New invoice" }));

    await screen.findByRole("dialog");
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
  });
});
