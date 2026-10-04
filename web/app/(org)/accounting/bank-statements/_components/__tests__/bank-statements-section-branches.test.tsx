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
  {
    id: 2,
    organization_id: 1,
    name: "BCA January",
    journal_id: null,
    date: null,
    balance_start: 0,
    balance_end: 0,
    state: "reconciled",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

function useHandlers(options?: { failStatements?: boolean }) {
  const created: unknown[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/bank-statements", () => {
      if (options?.failStatements) {
        return HttpResponse.json({ success: false, message: "Statements down." }, { status: 500 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { bank_statements: STATEMENTS },
      });
    }),
    http.get("*/api/v1/organizations/:organizationId/journals", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { journals: [{ id: 4, organization_id: 1, name: "BCA Journal" }] },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/bank-statements", async ({ request }) => {
      created.push(await request.json());
      return HttpResponse.json(
        { success: true, message: "Dibuat.", data: { bank_statement: { id: 3 } } },
        { status: 201 },
      );
    }),
  );
  return { created };
}

beforeEach(() => {});

describe("BankStatementsSection branches", () => {
  it("renders fallback names, journals, and dates", async () => {
    useHandlers();
    renderWithProviders(<BankStatementsSection orgId="1" />);
    expect(await screen.findByText("BS-1")).toBeInTheDocument();
    expect(screen.getByText("BCA January")).toBeInTheDocument();
    expect(screen.getByText("BCA Journal")).toBeInTheDocument();
    expect(screen.getByText("#null")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
    const badges = document.querySelectorAll('[data-slot="badge"]');
    expect(badges.length).toBeGreaterThan(0);
    const firstBadge = badges[0];
    if (firstBadge === undefined) throw new Error("expected a badge");
    expect(firstBadge.textContent).toContain("Draf");
  });

  it("links to the statement detail page", async () => {
    useHandlers();
    renderWithProviders(<BankStatementsSection orgId="1" />);
    await screen.findByText("BS-1");
    expect(screen.getByRole("link", { name: "BS-1" })).toHaveAttribute(
      "href",
      "/accounting/bank-statements/1",
    );
  });

  it("shows the error state with retry when statements fail", async () => {
    useHandlers({ failStatements: true });
    renderWithProviders(<BankStatementsSection orgId="1" />);
    expect(await screen.findByText("Statements down.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Coba lagi" })).toBeInTheDocument();
  });

  it("keeps save disabled without a journal and cancels cleanly", async () => {
    const user = userEvent.setup();
    const { created } = useHandlers();
    renderWithProviders(<BankStatementsSection orgId="1" />);
    await screen.findByText("BCA January");
    await user.click(screen.getByRole("button", { name: "Mutasi baru" }));
    await screen.findByText("Buat mutasi bank");
    expect(screen.getByRole("button", { name: "Buat" })).toBeDisabled();
    expect(created.length).toBe(0);
    await user.click(screen.getByRole("button", { name: "Batal" }));
    await waitFor(() => expect(screen.queryByText("Buat mutasi bank")).not.toBeInTheDocument());
  });

  it("edits statement fields in the create dialog", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<BankStatementsSection orgId="1" />);
    await screen.findByText("BCA January");
    await user.click(screen.getByRole("button", { name: "Mutasi baru" }));
    await screen.findByText("Buat mutasi bank");
    const dialog = screen.getByRole("dialog");
    const visible = (value: string) =>
      Array.from(dialog.querySelectorAll("input")).find(
        (input) =>
          input.getAttribute("aria-hidden") !== "true" && input.getAttribute("value") === value,
      ) as HTMLInputElement;
    const nameInput = visible("");
    await user.type(nameInput, "BCA Feb");
    expect(nameInput).toHaveValue("BCA Feb");
    const balanceInput = visible("0");
    await user.clear(balanceInput);
    await user.type(balanceInput, "250");
    expect(balanceInput).toHaveValue(250);
  });
});
