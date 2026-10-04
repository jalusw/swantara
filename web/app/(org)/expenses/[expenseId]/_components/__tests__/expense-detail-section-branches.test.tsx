import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ExpenseDetail } from "../expense-detail-section";

const BASE_REPORT = {
  id: 7,
  name: "Trip to Jakarta",
  employee_id: 1,
  state: "draft",
  payment_mode: "own_account",
  total_amount: 1500,
  entry_id: null,
  submitted_at: null,
  approved_by: null,
  lines: [],
};

const LINES = [
  {
    id: 31,
    category_id: 1,
    item_id: null,
    description: "Hotel stay",
    expense_date: "2026-03-01",
    quantity: 2,
    unit_price: 750,
    amount: 1500,
    tax_ids: [],
    dimension_id: null,
    project_id: null,
    reimbursable: true,
    receipt_attachment_id: null,
  },
  {
    id: 32,
    category_id: 1,
    item_id: null,
    description: null,
    expense_date: null,
    quantity: 1,
    unit_price: 100,
    amount: 100,
    tax_ids: [],
    dimension_id: null,
    project_id: null,
    reimbursable: false,
    receipt_attachment_id: null,
  },
];

function useHandlers(options?: { report?: unknown }) {
  const actions: string[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/expense-reports/:id", () => {
      if (options?.report === null) {
        return HttpResponse.json({ success: false, message: "Missing." }, { status: 404 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { report: options?.report ?? BASE_REPORT },
      });
    }),
    http.post(
      "*/api/v1/organizations/:organizationId/expense-reports/:id/:action",
      ({ params }) => {
        actions.push(String(params.action));
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      },
    ),
  );
  return { actions };
}

beforeEach(() => {});

describe("ExpenseDetail branches", () => {
  it("renders the draft report with submit action", async () => {
    useHandlers();
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    expect((await screen.findAllByText("Trip to Jakarta")).length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "Ajukan" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Setujui" })).not.toBeInTheDocument();
  });

  it("renders the not-found state when the report is missing", async () => {
    useHandlers({ report: null });
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="9" />);
    expect(await screen.findByText("Klaim biaya tidak ditemukan.")).toBeInTheDocument();
  });

  it("submits a draft report", async () => {
    const user = userEvent.setup();
    const { actions } = useHandlers();
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    await user.click(screen.getByRole("button", { name: "Ajukan" }));
    await waitFor(() => expect(actions).toContain("submit"));
  });

  it("renders approve and refuse for submitted reports", async () => {
    useHandlers({ report: { ...BASE_REPORT, state: "submitted" } });
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    expect(screen.getByRole("button", { name: "Setujui" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Tolak" })).toBeInTheDocument();
  });

  it("renders post for approved reports", async () => {
    const user = userEvent.setup();
    const { actions } = useHandlers({ report: { ...BASE_REPORT, state: "approved" } });
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    await user.click(screen.getByRole("button", { name: "Posting" }));
    await waitFor(() => expect(actions).toContain("post"));
  });

  it("renders reimburse and bill for posted reports", async () => {
    const user = userEvent.setup();
    const { actions } = useHandlers({ report: { ...BASE_REPORT, state: "posted" } });
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    await user.click(screen.getByRole("button", { name: "Ganti biaya" }));
    await user.click(screen.getByRole("button", { name: "Tagihkan" }));
    await waitFor(() => expect(actions).toContain("reimburse"));
    await waitFor(() => expect(actions).toContain("bill"));
  });

  it("renders lines with fallbacks", async () => {
    const user = userEvent.setup();
    useHandlers({
      report: {
        ...BASE_REPORT,
        submitted_at: "2026-03-02T00:00:00Z",
        approved_by: 9,
        entry_id: 11,
        lines: LINES,
      },
    });
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    expect(screen.getByText("#9")).toBeInTheDocument();
    expect(screen.getByText("#11")).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: "Baris" }));
    expect(await screen.findByText("Hotel stay")).toBeInTheDocument();
    expect(screen.getByText("✓")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders the no-lines state", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<ExpenseDetail orgId="1" expenseId="7" />);
    await screen.findAllByText("Trip to Jakarta");
    await user.click(screen.getByRole("tab", { name: "Baris" }));
    expect(await screen.findByText("Belum ada baris biaya")).toBeInTheDocument();
  });
});
