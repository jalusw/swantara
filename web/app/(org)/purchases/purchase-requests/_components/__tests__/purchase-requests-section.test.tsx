import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestsSection } from "../purchase-requests-section";

const STAMP = "2026-01-01T00:00:00Z";

const requester = {
  id: 7,
  organization_id: 1,
  name: "Aria Chen",
  display_name: "Aria Chen",
  is_organization: false,
  parent_id: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  tax_id: null,
  industry: null,
  currency_code: null,
  lang: "en",
  active: true,
  created_at: STAMP,
  updated_at: STAMP,
};

const department = {
  id: 3,
  organization_id: 1,
  name: "Engineering",
  description: null,
  parent_id: null,
  manager_id: null,
  dimension_id: null,
};

const request = {
  id: 1,
  organization_id: 1,
  name: "PR-0001",
  requester_id: 7,
  department_id: 3,
  state: "draft",
  needed_by: "2026-03-01",
  lines: [
    {
      id: 1,
      request_id: 1,
      item_id: 5,
      description: null,
      qty: 4,
      unit_id: null,
      needed_by: null,
    },
  ],
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
  is_manufactured: false,
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

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/purchase-requests", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { requisitions: [request] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [requester] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/departments", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { departments: [department] },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("PurchaseRequestsSection", () => {
  it("renders seeded requisitions with requester and department", async () => {
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    expect(await screen.findByText("PR-0001")).toBeInTheDocument();
    expect(screen.getByText("Aria Chen")).toBeInTheDocument();
    expect(screen.getByText("Engineering")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PurchaseRequestsSection orgId="1" />);

    await screen.findByText("PR-0001");
    await user.click(screen.getByRole("button", { name: "Permintaan baru" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Permintaan pembelian baru")).toBeInTheDocument();
  });
});
