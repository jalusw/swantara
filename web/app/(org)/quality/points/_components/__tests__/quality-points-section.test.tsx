import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { QualityPointsSection } from "../quality-points-section";

function useLocalPoints() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/quality-points", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          points: [
            {
              id: 1,
              organization_id: 1,
              item_id: null,
              operation: "Final inspection",
              test_type: "pass_fail",
              norm_min: null,
              norm_max: null,
              unit_id: null,
            },
            {
              id: 2,
              organization_id: 1,
              item_id: null,
              operation: "Incoming QC",
              test_type: "measure",
              norm_min: 10,
              norm_max: 20,
              unit_id: null,
            },
          ],
        },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { products: [] },
      }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { units: [] },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalPoints();
});

describe("QualityPointsSection", () => {
  it("renders seeded quality points", async () => {
    renderWithProviders(<QualityPointsSection orgId="1" />);

    expect(await screen.findByText("Final inspection")).toBeInTheDocument();
    expect(screen.getByText("Incoming QC")).toBeInTheDocument();
  });

  it("filters points by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<QualityPointsSection orgId="1" />);

    await screen.findByText("Final inspection");
    await user.type(screen.getByPlaceholderText("Search quality points…"), "Incoming");

    expect(await screen.findByText("Incoming QC")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Final inspection")).not.toBeInTheDocument());
  });
});
