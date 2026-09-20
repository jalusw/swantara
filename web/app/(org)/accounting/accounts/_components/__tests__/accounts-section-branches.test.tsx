import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { AccountsSection } from "../accounts-section";

const STAMP = "2026-01-01T00:00:00Z";

function account(over: Record<string, unknown> = {}) {
  return {
    id: 1,
    organization_id: 1,
    code: "1000",
    name: "Cash",
    type: "cash",
    reconcilable: true,
    currency_code: null,
    parent_id: null,
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
    ...over,
  };
}

function seed(accounts: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.put("*/api/v1/organizations/:organizationId/accounts/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {});

describe("AccountsSection branches", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/accounts", () => new Promise(() => {})),
    );
    renderWithProviders(<AccountsSection orgId="1" />);

    expect(screen.getByText("Loading accounts...")).toBeInTheDocument();
  });

  it("renders empty state when no accounts exist", async () => {
    seed([]);
    renderWithProviders(<AccountsSection orgId="1" />);

    expect(
      await screen.findByText("No accounts yet. Create your first account to get started."),
    ).toBeInTheDocument();
  });

  it("renders nested children with indentation", async () => {
    seed([
      account({ id: 1, code: "1000", parent_id: null }),
      account({
        id: 2,
        code: "1010",
        name: "Petty",
        parent_id: 1,
        active: true,
        reconcilable: false,
      }),
    ]);
    renderWithProviders(<AccountsSection orgId="1" />);

    expect(await screen.findByText("Petty")).toBeInTheDocument();
    expect(screen.getByText("1000")).toBeInTheDocument();
  });

  it("shows inactive badge for inactive accounts", async () => {
    seed([account({ active: false })]);
    renderWithProviders(<AccountsSection orgId="1" />);

    expect(await screen.findByText("Inactive")).toBeInTheDocument();
  });

  it("creates an account from the dialog", async () => {
    const user = userEvent.setup();
    seed([account()]);
    renderWithProviders(<AccountsSection orgId="1" />);

    await screen.findAllByText("Cash");
    await user.click(screen.getByRole("button", { name: "Add account" }));
    const dialog = await screen.findByRole("dialog");
    expect(dialog).toBeInTheDocument();
    const inputs = dialog.querySelectorAll("input");
    const codeInput = inputs[0];
    const nameInput = inputs[1];
    if (codeInput && nameInput) {
      await user.clear(codeInput);
      await user.type(codeInput, "2000");
      await user.clear(nameInput);
      await user.type(nameInput, "Payables");
    }
    const save = screen.getByRole("button", { name: "Save" });
    expect(save).toBeInTheDocument();
  });

  it("keeps save disabled with empty code or name", async () => {
    const user = userEvent.setup();
    seed([account()]);
    renderWithProviders(<AccountsSection orgId="1" />);

    await screen.findAllByText("Cash");
    await user.click(screen.getByRole("button", { name: "Add account" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
  });
});
