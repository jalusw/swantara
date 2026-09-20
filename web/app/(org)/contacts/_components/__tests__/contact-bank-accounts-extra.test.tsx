import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ContactBankAccounts } from "../contact-bank-accounts";

const account = {
  id: 5,
  organization_id: 1,
  contact_id: 1,
  account_holder: "Acme Corp",
  bank_name: "DBS",
  iban: null,
  swift_bic: "DBSBSGSG",
  account_number: "123456789",
  routing_number: null,
  currency_code: "USD",
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

function seedAccounts(accounts: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts/:contactId/bank-accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { bank_accounts: accounts } }),
    ),
  );
}

beforeEach(() => {});

describe("ContactBankAccounts extra", () => {
  it("lists existing bank accounts", async () => {
    seedAccounts([account]);
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={() => {}} />);

    expect(await screen.findByText("Acme Corp")).toBeInTheDocument();
    expect(screen.getByText("123456789")).toBeInTheDocument();
    expect(screen.getByText("DBSBSGSG")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    seedAccounts([]);
    const user = userEvent.setup();
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("No bank accounts");
    await user.click(screen.getByRole("button", { name: "Add bank account" }));

    expect(await screen.findByText("New bank account")).toBeInTheDocument();
  });

  it("opens the edit dialog prefilled with the account", async () => {
    seedAccounts([account]);
    const user = userEvent.setup();
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("Acme Corp");
    await user.click(screen.getByRole("button", { name: "Edit" }));

    expect(await screen.findByText("Edit bank account")).toBeInTheDocument();
    expect(screen.getByDisplayValue("Acme Corp")).toBeInTheDocument();
    expect(screen.getByDisplayValue("DBS")).toBeInTheDocument();
  });

  it("requires an account holder before saving", async () => {
    seedAccounts([]);
    const user = userEvent.setup();
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={() => {}} />);

    await screen.findByText("No bank accounts");
    await user.click(screen.getByRole("button", { name: "Add bank account" }));
    await user.click(await screen.findByRole("button", { name: "Save bank account" }));

    expect(await screen.findByText("Enter the account holder.")).toBeInTheDocument();
  });

  it("creates an account and refreshes the list", async () => {
    seedAccounts([]);
    const createCalls: unknown[] = [];
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/contacts/:contactId/bank-accounts",
        async ({ request }) => {
          createCalls.push(await request.json());
          return HttpResponse.json(
            { success: true, message: "Created.", data: { bank_account: account } },
            { status: 201 },
          );
        },
      ),
    );
    const onRefetch = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={onRefetch} />);

    await screen.findByText("No bank accounts");
    await user.click(screen.getByRole("button", { name: "Add bank account" }));
    await user.type(await screen.findByPlaceholderText("Account holder"), "Acme Corp");
    await user.click(screen.getByRole("button", { name: "Save bank account" }));

    await waitFor(() => expect(createCalls).toHaveLength(1));
    await waitFor(() => expect(onRefetch).toHaveBeenCalled());
  });

  it("deletes an account after confirmation", async () => {
    seedAccounts([account]);
    let deleteCalls = 0;
    server.use(
      http.delete(
        "*/api/v1/organizations/:organizationId/contacts/:contactId/bank-accounts/:accountId",
        () => {
          deleteCalls += 1;
          return HttpResponse.json({ success: true, message: "Deleted." });
        },
      ),
    );
    const onRefetch = vi.fn();
    const user = userEvent.setup();
    renderWithProviders(<ContactBankAccounts orgId="1" contactId="1" onRefetch={onRefetch} />);

    await screen.findByText("Acme Corp");
    await user.click(screen.getByRole("button", { name: "Delete" }));
    const confirmDialog = await screen.findByRole("alertdialog");
    await user.click(within(confirmDialog).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
    await waitFor(() => expect(onRefetch).toHaveBeenCalled());
  });
});
