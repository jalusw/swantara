import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { RmaFormDialog } from "../rma-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const contacts = [
  {
    id: 3,
    organization_id: 1,
    name: "Bluebird Trading Pte. Ltd.",
    display_name: "Bluebird Trading",
    is_organization: true,
    email: null,
    active: true,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const products = [
  {
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
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
  );
});

describe("RmaFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <RmaFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(screen.getByRole("heading", { name: "Retur baru" })).toBeInTheDocument();
    expect(screen.getByText("Baris retur")).toBeInTheDocument();
  });

  it("appends another line through the add line button", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <RmaFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    await user.click(screen.getByRole("button", { name: "Tambah baris" }));

    expect((await screen.findAllByRole("combobox", { name: "Item" })).length).toBe(2);
  });
});
