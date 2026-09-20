import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { BankStatementsSection } from "../bank-statements-section";

const STAMP = "2026-01-01T00:00:00Z";

const STATEMENTS = [
  {
    id: 1,
    organization_id: 1,
    name: null,
    journal_id: 4,
    date: "2026-02-01",
    balance_start: 1000,
    balance_end: 1500,
    state: "draft",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function seedStatements(statements: unknown[] = STATEMENTS) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/bank-statements", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { bank_statements: statements },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          journals: [{ id: 4, organization_id: 1, name: "BCA Journal", code: "BCA", type: "bank" }],
        },
      }),
    ),
  );
}

beforeEach(() => {});

describe("BankStatementsSection branches2", () => {
  it("renders the empty branch when no statements exist", async () => {
    seedStatements([]);
    renderWithProviders(<BankStatementsSection orgId="1" />);

    expect(await screen.findByText("No bank statements found.")).toBeInTheDocument();
  });

  it("creates a statement through the dialog success branch", async () => {
    seedStatements([]);
    let createBody: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/bank-statements", async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json(
          { success: true, message: "Created.", data: { bank_statement: { id: 9 } } },
          { status: 201 },
        );
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<BankStatementsSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "New statement" }));
    const dialog = await screen.findByRole("dialog");
    const { within } = await import("@testing-library/react");
    await user.click(within(dialog).getByRole("combobox", { name: "Bank journal" }));
    await user.click(await screen.findByRole("option", { name: "BCA Journal" }));
    await user.click(within(dialog).getByRole("button", { name: "Create" }));

    await waitFor(() => expect(createBody).not.toBeNull());
  });

  it("keeps the dialog open when journal is missing", async () => {
    seedStatements([]);
    const user = userEvent.setup();
    renderWithProviders(<BankStatementsSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "New statement" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
  });
});
