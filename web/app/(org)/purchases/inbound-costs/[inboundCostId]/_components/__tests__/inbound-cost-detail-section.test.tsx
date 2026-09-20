import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InboundCostDetail } from "../inbound-cost-detail-section";

const cost = {
  id: 1,
  name: "Freight January",
  date: "2026-01-05",
  state: "draft",
  target_shipment_ids: [3],
  movement_id: null,
};

function useInboundCostDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: cost }),
    ),
    http.get("*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines: [] } }),
    ),
    http.get(
      "*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId/adjustments",
      () => HttpResponse.json({ success: true, message: "OK.", data: { adjustments: [] } }),
    ),
  );
}

beforeEach(() => {
  useInboundCostDetailHandlers();
});

describe("InboundCostDetail", () => {
  it("renders cost name with state badge", async () => {
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(await screen.findByRole("heading", { name: "Freight January" })).toBeInTheDocument();
    expect(screen.getByText("Draft")).toBeInTheDocument();
  });

  it("shows empty lines on the cost lines tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    await screen.findByRole("heading", { name: "Freight January" });
    await user.click(screen.getByRole("tab", { name: "Cost lines" }));

    expect(await screen.findByText("No cost lines.")).toBeInTheDocument();
  });
});
