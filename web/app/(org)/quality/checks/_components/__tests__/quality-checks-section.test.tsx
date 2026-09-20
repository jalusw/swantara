import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityChecksSection } from "../quality-checks-section";

function useLocalChecks() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-checks", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          checks: [
            {
              id: 11,
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
            {
              id: 12,
              point_id: 1,
              item_id: 5,
              batch_id: null,
              shipment_id: 8,
              production_order_id: null,
              measured_value: 15,
              result: "pass",
              checked_by: null,
              checked_at: null,
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalChecks();
});

describe("QualityChecksSection", () => {
  it("renders seeded quality checks", async () => {
    renderWithProviders(<QualityChecksSection orgId="1" />);

    expect(await screen.findByText("#11")).toBeInTheDocument();
    expect(screen.getByText("#12")).toBeInTheDocument();
  });

  it("opens the record result dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<QualityChecksSection orgId="1" />);

    await screen.findByText("#11");
    await user.click(screen.getByRole("button", { name: "Record result" }));

    expect(await screen.findByText("Record Result")).toBeInTheDocument();
  });
});
