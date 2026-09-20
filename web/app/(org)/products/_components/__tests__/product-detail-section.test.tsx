import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProductDetail } from "../product-detail-section";

const item = {
  id: 5,
  organization_id: 1,
  name: "Canvas Tote",
  category_id: null,
  type: "stockable",
  unit_id: null,
  purchase_unit_id: null,
  list_price: 50,
  standard_cost: 30,
  is_purchasable: true,
  is_sellable: true,
  is_manufactured: false,
  tracking: "none",
  weight: 0,
  volume: 0,
  hs_code: null,
  description_sale: null,
  description_purchase: null,
  active: true,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

const variants = [
  {
    id: 21,
    item_id: 5,
    sku: "TOTE-RED",
    barcode: "899000000021",
    attribute_json: { Color: "Red" },
    extra_cost: 0,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

const recipes = [
  {
    id: 31,
    organization_id: 1,
    item_id: 5,
    code: "BOM-TOTE-001",
    qty: 1,
    unit_id: null,
    type: "manufacture",
    version: 1,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

function useDetailHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products/:itemId", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { item } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products/:itemId/variants", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { variants } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/recipes", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { recipes } }),
    ),
  );
}

beforeEach(() => {
  useDetailHandlers();
});

describe("ProductDetail", () => {
  it("renders the item overview", async () => {
    renderWithProviders(<ProductDetail orgId="1" itemId="5" />);

    expect((await screen.findAllByText("Canvas Tote")).length).toBeGreaterThan(0);
    expect(screen.getByText("General")).toBeInTheDocument();
    expect(screen.getByText("Pricing")).toBeInTheDocument();
  });

  it("shows variants on the variants tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductDetail orgId="1" itemId="5" />);

    await screen.findAllByText("Canvas Tote");
    await user.click(screen.getByRole("tab", { name: "Variants" }));

    expect(await screen.findByText("TOTE-RED")).toBeInTheDocument();
  });

  it("shows bills of materials on the recipes tab", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductDetail orgId="1" itemId="5" />);

    await screen.findAllByText("Canvas Tote");
    await user.click(screen.getByRole("tab", { name: "Bills of materials" }));

    expect(await screen.findByText("BOM-TOTE-001")).toBeInTheDocument();
  });
});
