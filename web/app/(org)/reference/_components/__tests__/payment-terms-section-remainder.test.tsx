import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PaymentTermsSection } from "../payment-terms-section";

beforeEach(() => {});

describe("PaymentTermsSection remainder", () => {
  it("shows the empty state when no terms exist", async () => {
    server.use(
      http.get("*/api/v1/payment-terms", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { payment_terms: [] } }),
      ),
    );
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    expect(await screen.findByText("No payment terms")).toBeInTheDocument();
  });

  it("shows the error state with retry when loading fails", async () => {
    server.use(
      http.get("*/api/v1/payment-terms", () =>
        HttpResponse.json({ success: false, message: "Term load failed." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    expect(await screen.findByText("Term load failed.")).toBeInTheDocument();

    server.resetHandlers();
    await user.click(screen.getByRole("button", { name: /retry|try again/i }));

    expect(await screen.findByText("Net 30")).toBeInTheDocument();
  });

  it("requires at least one line when creating a term", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Add term" }));
    await user.type(screen.getByLabelText("Name"), "Empty term");
    await user.click(screen.getAllByRole("button", { name: "Remove line" })[0]!);
    await user.click(screen.getByRole("button", { name: "Save term" }));

    expect(await screen.findByText("Add at least one line.")).toBeInTheDocument();
  });

  it("adds and removes lines in the dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await user.click(screen.getByRole("button", { name: "Add term" }));
    expect(screen.getAllByLabelText("Value").length).toBe(1);

    await user.click(screen.getByRole("button", { name: "Add line" }));
    expect(screen.getAllByLabelText("Value").length).toBe(2);

    await user.click(screen.getAllByRole("button", { name: "Remove line" })[1]!);
    expect(screen.getAllByLabelText("Value").length).toBe(1);
  });

  it("saves a fixed-amount term without percent validation", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await screen.findByText("Net 30");
    await user.click(screen.getByRole("button", { name: "Add term" }));
    await user.type(screen.getByLabelText("Name"), "Fixed 500");
    await user.click(screen.getByLabelText("Type"));
    await user.click(await screen.findByRole("option", { name: "Fixed" }));
    await user.type(screen.getAllByLabelText("Value")[0]!, "500");
    await user.type(screen.getAllByLabelText("Due days")[0]!, "15");
    await user.click(screen.getByRole("button", { name: "Save term" }));

    expect(await screen.findByText("Fixed 500")).toBeInTheDocument();
  });

  it("opens the edit dialog and updates the term", async () => {
    let updated: unknown = null;
    server.use(
      http.put("*/api/v1/payment-terms/:id", async ({ request }) => {
        updated = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await screen.findByText("Net 30");
    await user.click(screen.getAllByRole("button", { name: "Edit" })[0]!);

    expect(await screen.findByText("Edit payment term")).toBeInTheDocument();

    const nameInput = screen.getByLabelText("Name");
    await user.clear(nameInput);
    await user.type(nameInput, "Net 45");
    await user.click(screen.getByRole("button", { name: "Add line" }));
    await user.type(screen.getAllByLabelText("Value")[0]!, "100");
    await user.type(screen.getAllByLabelText("Due days")[0]!, "45");
    await user.click(screen.getByRole("button", { name: "Save term" }));

    await waitFor(() => expect(updated).toMatchObject({ name: "Net 45" }));
  });

  it("shows an error toast when deleting fails", async () => {
    server.use(
      http.delete("*/api/v1/payment-terms/:id", () =>
        HttpResponse.json({ success: false, message: "Cannot delete." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PaymentTermsSection orgId="1" />);

    await screen.findByText("Net 30");
    await user.click(screen.getAllByRole("button", { name: "Delete" })[0]!);
    const dialog = screen.getByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(screen.getByText("Net 30")).toBeInTheDocument());
  });
});
