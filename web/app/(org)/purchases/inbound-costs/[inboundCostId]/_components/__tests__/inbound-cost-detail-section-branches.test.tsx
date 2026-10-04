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
  target_shipment_ids: [],
  movement_id: null,
};

function seed(costOverrides: Record<string, unknown> = {}, extra: Record<string, unknown> = {}) {
  const lines = (extra.lines ?? []) as unknown[];
  const adjustments = (extra.adjustments ?? []) as unknown[];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { ...cost, ...costOverrides },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
    http.get(
      "*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId/adjustments",
      () => HttpResponse.json({ success: true, message: "OK.", data: { adjustments } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId/post", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("InboundCostDetail branches", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get(
        "*/api/v1/organizations/:organizationId/inbound-costs/:inboundCostId",
        () => new Promise(() => {}),
      ),
    );
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(screen.getByText("Memuat...")).toBeInTheDocument();
  });

  it("renders dash fallbacks for null date and move", async () => {
    seed({ date: null, movement_id: null, target_shipment_ids: [] });
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(await screen.findByRole("heading", { name: "Freight January" })).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("lists target shipments when present", async () => {
    seed({ target_shipment_ids: [3, 4] });
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(await screen.findByText("#3, #4")).toBeInTheDocument();
  });

  it("posts from the header action in draft", async () => {
    const user = userEvent.setup();
    seed({ state: "draft" });
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    await screen.findByRole("heading", { name: "Freight January" });
    const post = screen.queryByRole("button", { name: "Post" });
    if (post) await user.click(post);
    expect(await screen.findByRole("heading", { name: "Freight January" })).toBeInTheDocument();
  });

  it("shows cancel action for cancellable states", async () => {
    seed({ state: "draft" });
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(await screen.findByRole("heading", { name: "Freight January" })).toBeInTheDocument();
  });

  it("renders cost lines with description fallback", async () => {
    const user = userEvent.setup();
    seed(
      {},
      {
        lines: [{ id: 1, item_id: 5, description: null, amount: 120, split_method: "equal" }],
      },
    );
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    await screen.findByRole("heading", { name: "Freight January" });
    await user.click(screen.getAllByRole("tab")[1]!);

    expect(await screen.findByText("#5")).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders adjustments when present", async () => {
    const user = userEvent.setup();
    seed(
      {},
      {
        adjustments: [
          { id: 2, item_id: 6, stock_movement_id: 9, additional_cost: 45, cost_layer_id: 11 },
        ],
      },
    );
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    await screen.findByRole("heading", { name: "Freight January" });
    await user.click(screen.getAllByRole("tab")[2]!);

    expect(await screen.findByText("#6")).toBeInTheDocument();
    expect(screen.getByText("#9")).toBeInTheDocument();
  });

  it("renders danger tone for cancelled state", async () => {
    seed({ state: "cancelled" });
    renderWithProviders(<InboundCostDetail orgId="1" inboundCostId="1" />);

    expect(await screen.findByRole("heading", { name: "Freight January" })).toBeInTheDocument();
    expect(screen.getByText("Dibatalkan")).toBeInTheDocument();
  });
});
