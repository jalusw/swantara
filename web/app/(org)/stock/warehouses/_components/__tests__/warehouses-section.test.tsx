import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { WarehousesSection } from "../warehouses-section";

const STAMP = "2026-01-01T00:00:00Z";

const warehouses = [
  {
    id: 1,
    organization_id: 1,
    name: "Main Warehouse",
    code: "WH-01",
    line1: null,
    line2: null,
    city: null,
    state: null,
    postal_code: null,
    country_code: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/warehouses", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { warehouses } }),
    ),
  );
});

describe("WarehousesSection", () => {
  it("renders seeded warehouses", async () => {
    renderWithProviders(<WarehousesSection orgId="1" />);

    expect(await screen.findByText("Main Warehouse")).toBeInTheDocument();
    expect(screen.getByText("WH-01")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WarehousesSection orgId="1" />);

    await screen.findByText("Main Warehouse");
    await user.click(screen.getByRole("button", { name: "Add warehouse" }));

    expect(await screen.findByRole("heading", { name: "New warehouse" })).toBeInTheDocument();
  });
});
