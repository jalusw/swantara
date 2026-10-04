import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PayrollRunDetail } from "../payroll-run-detail-section";

function runFixture(overrides = {}) {
  return {
    id: 3,
    organization_id: 1,
    name: "March 2026",
    period_start: "2026-03-01",
    period_end: "2026-03-31",
    state: "draft",
    ...overrides,
  };
}

function payslipFixture(overrides = {}) {
  return {
    id: 11,
    run_id: 3,
    employee_id: 1,
    contract_id: 5,
    gross: 10000,
    net: 8500,
    entry_id: null,
    state: "draft",
    lines: [
      {
        id: 101,
        payslip_id: 11,
        rule_id: 2,
        code: "BASIC",
        name: "Basic salary",
        category: "earning",
        amount: 10000,
      },
      {
        id: 102,
        payslip_id: 11,
        rule_id: 3,
        code: "TAX",
        name: "Income tax",
        category: "deduction",
        amount: -1500,
      },
    ],
    ...overrides,
  };
}

function useRunHandlers(run: Record<string, unknown>, payslips: unknown[] = [payslipFixture()]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/payroll-runs/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { run, payslips } }),
    ),
  );
}

beforeEach(() => {});

describe("PayrollRunDetail remainder", () => {
  it("shows the loading state while fetching", () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/payroll-runs/:id", async () => {
        await new Promise((resolve) => setTimeout(resolve, 50));
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { run: runFixture(), payslips: [] },
        });
      }),
    );
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    expect(screen.getByText("Memuat...")).toBeInTheDocument();
  });

  it("shows the not-found state when the run is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/payroll-runs/:id", () =>
        HttpResponse.json({ success: false, message: "Not found." }, { status: 404 }),
      ),
    );
    renderWithProviders(<PayrollRunDetail orgId="1" runId="999" />);

    expect(await screen.findByText("Penggajian tidak ditemukan.")).toBeInTheDocument();
  });

  it("shows the empty payslips message on the payslips tab", async () => {
    useRunHandlers(runFixture(), []);
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("tab", { name: "Slip gaji" }));

    expect(await screen.findByText("Belum ada slip gaji yang dihitung")).toBeInTheDocument();
  });

  it("renders payslip lines with earning and deduction badges", async () => {
    useRunHandlers(runFixture());
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("tab", { name: "Slip gaji" }));

    expect(await screen.findByText("Basic salary")).toBeInTheDocument();
    expect(screen.getByText("earning")).toBeInTheDocument();
    expect(screen.getByText("deduction")).toBeInTheDocument();
  });

  it("confirms a draft run", async () => {
    let confirmed: unknown = null;
    useRunHandlers(runFixture());
    server.use(
      http.post(
        "*/api/v1/organizations/:organizationId/payroll-runs/:id/confirm",
        async ({ request }) => {
          confirmed = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("button", { name: "Konfirmasi" }));

    await waitFor(() => expect(confirmed).toMatchObject({ journal_id: 1 }));
  });

  it("pays a confirmed run", async () => {
    let paid = false;
    useRunHandlers(runFixture({ state: "confirmed" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/payroll-runs/:id/pay", () => {
        paid = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("button", { name: "Bayar" }));

    await waitFor(() => expect(paid).toBe(true));
  });

  it("closes a paid run", async () => {
    let closed = false;
    useRunHandlers(runFixture({ state: "paid" }));
    server.use(
      http.post("*/api/v1/organizations/:organizationId/payroll-runs/:id/close", () => {
        closed = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("button", { name: "Tutup" }));

    await waitFor(() => expect(closed).toBe(true));
  });

  it("hides all actions for a closed run", async () => {
    useRunHandlers(runFixture({ state: "closed" }));
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    expect(screen.queryByRole("button", { name: "Konfirmasi" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Bayar" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Tutup" })).not.toBeInTheDocument();
  });
});
