import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityAlertDetail } from "../quality-alert-detail-section";

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
    ...overrides,
  };
}

function useAlert(alert: unknown, status = 200, message = "OK.") {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-alerts/:id", () =>
      status === 200
        ? HttpResponse.json({ success: true, message, data: { alert } })
        : HttpResponse.json({ success: false, message }, { status }),
    ),
    http.put("*/api/v1/organizations/:organizationId/quality-alerts/:id/state", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { alert: {} } }),
    ),
  );
}

beforeEach(() => {});

describe("QualityAlertDetail branches", () => {
  it("shows the service error message", async () => {
    useAlert(null, 500, "Alert down.");
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    expect(await screen.findByText("Alert down.")).toBeInTheDocument();
  });

  it("shows the not-found state when the alert payload is empty", async () => {
    useAlert(null);
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    expect(await screen.findByText("Peringatan tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders fallbacks for a sparse in-progress alert and solves it", async () => {
    let transitioned: unknown = null;
    useAlert(
      alertFixture({
        title: "",
        description: "",
        severity: null,
        state: "in_progress",
        assigned_to: 7,
      }),
    );
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
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    expect(await screen.findByText("Peringatan #1")).toBeInTheDocument();
    expect(screen.getByText("Pengguna #7")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: (name) => name.includes("Selesai") }));

    await waitFor(() => expect(transitioned).toMatchObject({ state: "solved" }));
  });

  it("shows no transitions for solved alerts", async () => {
    useAlert(alertFixture({ state: "solved" }));
    renderWithProviders(<QualityAlertDetail orgId="1" alertId="1" />);

    expect(await screen.findByText("Leaking valve")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: (name) => name.includes("Selesai") })).toBeNull();
  });
});
