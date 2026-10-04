import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PayrollRunDetail } from "../payroll-run-detail-section";

function useLocalRun() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/payroll-runs/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          run: {
            id: 3,
            organization_id: 1,
            name: "March 2026",
            period_start: "2026-03-01",
            period_end: "2026-03-31",
            state: "draft",
          },
          payslips: [
            {
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
              ],
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalRun();
});

describe("PayrollRunDetail", () => {
  it("renders the run header with totals", async () => {
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    expect((await screen.findAllByText("March 2026")).length).toBeGreaterThan(0);
    expect(screen.getByText("10.000")).toBeInTheDocument();
  });

  it("switches to the payslips tab on selection", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunDetail orgId="1" runId="3" />);

    await screen.findAllByText("March 2026");
    await user.click(screen.getByRole("tab", { name: "Slip gaji" }));

    expect(await screen.findByText("Basic salary")).toBeInTheDocument();
  });
});
