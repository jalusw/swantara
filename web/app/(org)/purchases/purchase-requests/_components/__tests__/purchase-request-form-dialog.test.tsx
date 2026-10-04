import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PurchaseRequestFormDialog } from "../purchase-request-form-dialog";

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

describe("PurchaseRequestFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(screen.getByText("Permintaan pembelian baru")).toBeInTheDocument();
    expect(screen.getByText("Baris permintaan")).toBeInTheDocument();
  });

  it("requires a requester before saving", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <PurchaseRequestFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={onSave} />,
    );

    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Pilih pemohon.")).toBeInTheDocument();
    expect(onSave).not.toHaveBeenCalled();
  });
});
