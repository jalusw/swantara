import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityAlertDetail } from "../quality-alert-detail-section";

let transitionCalled = false;

function useLocalAlert() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-alerts/:id", () =>
      HttpResponse.json({
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
            state: "open",
            assigned_to: null,
          },
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
  useLocalAlert();
});

describe("QualityAlertDetail", () => {
  it("renders the alert title", async () => {
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    expect(await screen.findByText("Leaking valve")).toBeInTheDocument();
  });

  it("transitions the alert forward", async () => {
    const user = userEvent.setup();
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    await screen.findByText("Leaking valve");
    await user.click(screen.getByRole("button", { name: (name) => name.includes("Dalam proses") }));

    await waitFor(() => expect(transitionCalled).toBe(true));
  });
});
