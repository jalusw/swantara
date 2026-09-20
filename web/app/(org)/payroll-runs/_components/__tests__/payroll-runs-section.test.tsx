import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PayrollRunsSection } from "../payroll-runs-section";

function useLocalRuns() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/payroll-runs", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          payroll_runs: [
            {
              id: 1,
              organization_id: 1,
              name: "January 2026",
              period_start: "2026-01-01",
              period_end: "2026-01-31",
              state: "draft",
            },
            {
              id: 2,
              organization_id: 1,
              name: "February 2026",
              period_start: "2026-02-01",
              period_end: "2026-02-28",
              state: "confirmed",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalRuns();
});

describe("PayrollRunsSection", () => {
  it("renders seeded payroll runs", async () => {
    renderWithProviders(<PayrollRunsSection orgId="1" />);

    expect(await screen.findByText("January 2026")).toBeInTheDocument();
    expect(screen.getByText("February 2026")).toBeInTheDocument();
  });

  it("filters runs by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PayrollRunsSection orgId="1" />);

    await screen.findByText("January 2026");
    await user.type(screen.getByPlaceholderText("Search runs…"), "February");

    expect((await screen.findAllByText("February 2026")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("January 2026")).not.toBeInTheDocument());
  });
});
