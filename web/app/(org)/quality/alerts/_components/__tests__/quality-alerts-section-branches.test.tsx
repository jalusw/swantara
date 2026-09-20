import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityAlertsSection } from "../quality-alerts-section";

const STAMP = "2026-01-01T00:00:00Z";

function alertFixture(overrides = {}) {
  return {
    id: 1,
    item_id: 5,
    batch_id: null,
    check_id: 11,
    title: "Leaking valve",
    description: "Coolant leak detected on line 2",
    severity: "high",
    state: "open",
    assigned_to: null,
    created_at: STAMP,
    updated_at: STAMP,
    ...overrides,
  };
}

function useAlerts(alerts: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-alerts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { alerts } }),
    ),
    http.put("*/api/v1/organizations/:organizationId/quality-alerts/:id/state", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { alert: {} } }),
    ),
  );
}

beforeEach(() => {});

describe("QualityAlertsSection branches", () => {
  it("shows the error state with retry", async () => {
    let calls = 0;
    server.use(
      http.get("*/api/v1/organizations/:organizationId/quality-alerts", () => {
        calls += 1;
        if (calls === 1) {
          return HttpResponse.json({ success: false, message: "Alerts down." }, { status: 500 });
        }
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { alerts: [alertFixture()] },
        });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<QualityAlertsSection orgId="1" />);

    expect(await screen.findByText("Alerts down.")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Retry" }));

    expect(await screen.findByText("Leaking valve")).toBeInTheDocument();
  });

  it("renders fallbacks for sparse and historical alerts", async () => {
    useAlerts([
      alertFixture(),
      alertFixture({
        id: 3,
        title: "",
        description: null,
        severity: null,
      }),
      alertFixture({
        id: 4,
        title: "Long running leak",
        description: "x".repeat(80),
        severity: null,
        state: "in_progress",
      }),
      alertFixture({ id: 5, title: "Old alert", state: "cancelled" }),
    ]);
    renderWithProviders(<QualityAlertsSection orgId="1" />);

    expect(await screen.findByText("Leaking valve")).toBeInTheDocument();
    expect(screen.getByText("Alert #3")).toBeInTheDocument();
    expect(screen.getByText(`${"x".repeat(60)}…`)).toBeInTheDocument();
    expect(screen.queryByText("Old alert")).toBeNull();
  });

  it("transitions an in-progress alert to solved", async () => {
    let transitioned: unknown = null;
    useAlerts([alertFixture({ id: 4, title: "Stuck valve", state: "in_progress" })]);
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/quality-alerts/:id/state",
        async ({ request }) => {
          transitioned = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: { alert: {} } });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<QualityAlertsSection orgId="1" />);

    await screen.findByText("Stuck valve");
    const transitions = screen.getAllByRole("button", {
      name: (name) => name.endsWith(": Stuck valve"),
    });
    const solve = transitions[0];
    if (!solve) throw new Error("Expected a transition button");
    await user.click(solve);

    await waitFor(() => expect(transitioned).toMatchObject({ state: "solved" }));
  });
});
