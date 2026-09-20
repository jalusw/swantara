import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { BomsSection } from "../recipes-section";

const STAMP = "2026-01-01T00:00:00Z";

const recipe = {
  id: 1,
  organization_id: 1,
  item_id: 5,
  code: "BOM-001",
  qty: 1,
  unit_id: null,
  type: "manufacture",
  version: 1,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const item = {
  id: 5,
  organization_id: 1,
  name: "Finished Widget",
  category_id: null,
  type: "stockable",
  unit_id: null,
  purchase_unit_id: null,
  list_price: 100,
  standard_cost: 60,
  is_purchasable: true,
  is_sellable: true,
  is_manufactured: true,
  tracking: "none",
  weight: 0,
  volume: 0,
  hs_code: null,
  description_sale: null,
  description_purchase: null,
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const unit = {
  id: 1,
  category_id: 1,
  name: "Units",
  factor: 1,
  unit_type: "reference",
  rounding: 0.01,
  created_at: STAMP,
  updated_at: STAMP,
};

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/recipes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { recipes: [recipe] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { units: [unit] } }),
    ),
  );
});

describe("BomsSection", () => {
  it("renders seeded recipes with item names", async () => {
    renderWithProviders(<BomsSection orgId="1" />);

    expect(await screen.findByText("BOM-001")).toBeInTheDocument();
    expect(screen.getByText("Finished Widget")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<BomsSection orgId="1" />);

    await screen.findByText("BOM-001");
    await user.click(screen.getByRole("button", { name: "Add BoM" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("New bill of materials")).toBeInTheDocument();
  });
});
