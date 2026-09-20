import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierQuoteRequestsSection } from "../supplier-quote-requests-section";

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

const quoteRequest = {
  id: 1,
  organization_id: 1,
  name: "QuoteRequest-0001",
  requester_id: 1,
  supplier_id: 7,
  currency_code: "USD",
  state: "draft",
  order_date: "2026-02-01",
  quote_deadline: "2026-02-10",
  notes: null,
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
    http.get("*/api/v1/organizations/:organizationId/supplier-quote-requests", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { quoteRequests: [quoteRequest] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/contacts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { contacts: [supplier] } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products: [item] } }),
    ),
  );
});

describe("SupplierQuoteRequestsSection", () => {
  it("renders seeded quoteRequests with supplier names", async () => {
    renderWithProviders(<SupplierQuoteRequestsSection orgId="1" />);

    expect(await screen.findByText("QuoteRequest-0001")).toBeInTheDocument();
    expect(screen.getByText("Acme Supplier")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<SupplierQuoteRequestsSection orgId="1" />);

    await screen.findByText("QuoteRequest-0001");
    await user.click(screen.getByRole("button", { name: "New QuoteRequest" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "New QuoteRequest" })).toBeInTheDocument();
  });
});
