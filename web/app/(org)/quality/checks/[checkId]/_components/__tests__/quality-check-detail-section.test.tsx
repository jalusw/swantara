import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityCheckDetail } from "../quality-check-detail-section";

function useLocalCheck() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-checks/:id", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          check: {
            id: 3,
            point_id: 1,
            item_id: 5,
            batch_id: null,
            shipment_id: 7,
            production_order_id: null,
            measured_value: null,
            result: "pending",
            checked_by: null,
            checked_at: null,
          },
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalCheck();
});

describe("QualityCheckDetail", () => {
  it("renders the check title", async () => {
    renderWithProviders(<QualityCheckDetail orgId="1" checkId="3" />);

    expect(await screen.findByText("Pemeriksaan mutu #3")).toBeInTheDocument();
  });

  it("opens the record result dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<QualityCheckDetail orgId="1" checkId="3" />);

    await screen.findByText("Pemeriksaan mutu #3");
    await user.click(screen.getByRole("button", { name: "Catat hasil" }));

    expect(await screen.findByRole("heading", { name: "Catat hasil" })).toBeInTheDocument();
  });
});
