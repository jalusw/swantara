import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityAlertsSection } from "../quality-alerts-section";

let transitionCalled = false;

function useLocalAlerts() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-alerts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          alerts: [
            {
              id: 1,
              item_id: 5,
              batch_id: null,
              check_id: 11,
              title: "Leaking valve",
              description: "Coolant leak detected on line 2",
              severity: "high",
              state: "open",
              assigned_to: null,
            },
            {
              id: 2,
              item_id: 6,
              batch_id: null,
              check_id: null,
              title: "Cracked housing",
              description: "Housing crack found during inspection",
              severity: "medium",
              state: "open",
              assigned_to: null,
            },
          ],
        },
      }),
    ),
    http.put("*/api/v1/organizations/:organizationId/quality-alerts/:id/state", () => {
      transitionCalled = true;
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          alert: {
            id: 1,
            item_id: 5,
            batch_id: null,
            check_id: 11,
            title: "Leaking valve",
            description: "Coolant leak detected on line 2",
            severity: "high",
            state: "in_progress",
            assigned_to: null,
          },
        },
      });
    }),
  );
}

beforeEach(() => {
  transitionCalled = false;
  useLocalAlerts();
});

describe("QualityAlertsSection", () => {
  it("renders seeded alerts", async () => {
    renderWithProviders(<QualityAlertsSection orgId="1" />);

    expect(await screen.findByText("Leaking valve")).toBeInTheDocument();
    expect(screen.getByText("Cracked housing")).toBeInTheDocument();
  });

  it("transitions an alert to in progress", async () => {
    const user = userEvent.setup();
    renderWithProviders(<QualityAlertsSection orgId="1" />);

    await screen.findByText("Leaking valve");
    const transitions = screen.getAllByRole("button", {
      name: (name) => name.endsWith(": Leaking valve"),
    });
    const transitionButton = transitions[0];
    if (transitionButton === undefined) {
      throw new Error("Expected a transition button");
    }
    await user.click(transitionButton);

    await waitFor(() => expect(transitionCalled).toBe(true));
  });
});
