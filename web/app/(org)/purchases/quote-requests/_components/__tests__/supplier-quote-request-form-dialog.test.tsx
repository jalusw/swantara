import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestFormDialog } from "../supplier-quote-request-form-dialog";

const STAMP = "2026-01-01T00:00:00Z";

const supplier = {
  id: 7,
  organization_id: 1,
  name: "Acme Supplier Co",
  display_name: "Acme Supplier",
  is_organization: true,
  parent_id: null,
  email: null,
  phone: null,
  mobile: null,
  website: null,
  tax_id: null,
  industry: null,
  currency_code: "USD",
  lang: "en",
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
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("SupplierQuoteRequestFormDialog", () => {
  it("renders the create form when open", () => {
    renderWithProviders(
      <SupplierQuoteRequestFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("Permintaan penawaran baru")).toBeInTheDocument();
    expect(screen.getByText("Baris permintaan penawaran")).toBeInTheDocument();
  });

  it("does not save when the line item is missing", async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    renderWithProviders(
      <SupplierQuoteRequestFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        onSave={onSave}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Simpan" }));

    await waitFor(() => expect(screen.getByRole("dialog")).toBeInTheDocument());
    expect(onSave).not.toHaveBeenCalled();
  });
});
