import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestFormDialog } from "../supplier-quote-request-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

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

const item = {
  id: 5,
  organization_id: 1,
  name: "Finished Widget",
  type: "stockable",
  list_price: 100,
  standard_cost: 60,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

function renderDialog(onSave = vi.fn()) {
  return renderWithProviders(
    <SupplierQuoteRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
  );
}

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("SupplierQuoteRequestFormDialog remainder", () => {
  it("adds and removes lines", async () => {
    const user = userEvent.setup();
    renderDialog();

    expect(screen.getAllByText("Item").length).toBe(1);
    await user.click(screen.getByRole("button", { name: "Add line" }));
    expect(screen.getAllByText("Item").length).toBe(2);

    await user.click(screen.getAllByRole("button", { name: "Remove" })[0]!);
    expect(screen.getAllByText("Item").length).toBe(1);
  });

  it("shows the empty-lines message after removing all lines", async () => {
    const user = userEvent.setup();
    renderDialog();

    await user.click(screen.getByRole("button", { name: "Remove" }));

    expect(screen.getByText("Add at least one line.")).toBeInTheDocument();
  });

  it("blocks saving when every line was removed", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDialog(onSave);

    await user.click(screen.getByRole("button", { name: "Remove" }));
    await user.click(screen.getByRole("button", { name: "Save" }));

    expect(onSave).not.toHaveBeenCalled();
  });

  it("creates an quoteRequest with supplier, dates and line details", async () => {
    let created: unknown = null;
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/supplier-quote-requests",
        async ({ request }) => {
          created = await request.json();
          return HttpResponse.json(
            { success: true, message: "Created.", data: {} },
            { status: 201 },
          );
        },
      ),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDialog(onSave);

    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByLabelText("Supplier"));
    await user.click(await screen.findByRole("option", { name: "Acme Supplier" }));
    await user.type(within(dialog).getByPlaceholderText("USD"), "USD");

    await user.click(within(dialog).getByLabelText("Item 1"));
    await user.click(await screen.findByRole("option", { name: "Finished Widget" }));

    const qtyInput = within(dialog).getAllByRole("spinbutton")[0]!;
    await user.clear(qtyInput);
    await user.type(qtyInput, "3");

    await user.click(within(dialog).getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).toHaveBeenCalled());
    expect(created).toMatchObject({
      supplier_id: 7,
      currency_code: "USD",
      lines: [expect.objectContaining({ item_id: 5, qty: 3 })],
    });
  });

  it("shows an error toast when creation fails", async () => {
    server.use(
      http.post("*/api/v1/organizations/:organizationId/supplier-quote-requests", () =>
        HttpResponse.json({ success: false, message: "Cannot create." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderDialog(onSave);

    const dialog = screen.getByRole("dialog");
    await user.click(within(dialog).getByLabelText("Item 1"));
    await user.click(await screen.findByRole("option", { name: "Finished Widget" }));
    await user.click(within(dialog).getByRole("button", { name: "Save" }));

    await waitFor(() => expect(onSave).not.toHaveBeenCalled());
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("updates line description and needed-by date", async () => {
    const user = userEvent.setup();
    renderDialog();

    const dialog = screen.getByRole("dialog");
    const description = within(dialog).getByPlaceholderText("Description");
    await user.type(description, "Urgent restock");
    expect(description).toHaveValue("Urgent restock");
  });
});
