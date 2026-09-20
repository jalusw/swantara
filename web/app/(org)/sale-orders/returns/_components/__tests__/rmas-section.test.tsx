import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { RmasSection } from "../rmas-section";

const STAMP = "2026-01-01T00:00:00Z";

const rmas = [
  {
    id: 9,
    organization_id: 1,
    name: "RMA-0001",
    type: "customer_return",
    contact_id: 3,
    origin_order_type: "sale_order",
    origin_order_id: 7,
    reason: "Damaged",
    state: "draft",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/rmas", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rmas } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [] } }),
    ),
  );
});

describe("RmasSection", () => {
  it("renders seeded RMAs", async () => {
    renderWithProviders(<RmasSection orgId="1" />);

    expect(await screen.findByText("RMA-0001")).toBeInTheDocument();
    expect(screen.getByText("SALE_ORDER-7")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RmasSection orgId="1" />);

    await screen.findByText("RMA-0001");
    await user.click(screen.getByRole("button", { name: "New RMA" }));

    expect(await screen.findByRole("heading", { name: "Create RMA" })).toBeInTheDocument();
  });
});
