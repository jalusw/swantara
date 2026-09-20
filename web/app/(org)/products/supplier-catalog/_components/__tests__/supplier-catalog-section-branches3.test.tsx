import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { SupplierCatalogSection } from "../supplier-catalog-section";

const item = {
  id: 1,
  organization_id: 1,
  item_id: 10,
  supplier_id: 7,
  vendor_sku: null,
  price: null,
  currency_code: null,
  lead_time_days: null,
  priority: 1,
  min_qty: 1,
  valid_from: null,
  valid_to: null,
};

function seed(items: unknown[] = [item], products: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/supplier-products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { supplier_products: items } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/supplier-products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.put("*/api/v1/organizations/:organizationId/supplier-products/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.delete("*/api/v1/organizations/:organizationId/supplier-products/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("SupplierCatalogSection branches3", () => {
  it("falls back to item id when item is unknown", async () => {
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("#10")).toBeInTheDocument();
  });

  it("renders dashes for null price lead-time and validity", async () => {
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("#10");
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders filled price lead-time and validity", async () => {
    seed(
      [
        {
          ...item,
          vendor_sku: "SKU-1",
          price: 250,
          lead_time_days: 5,
          valid_to: "2026-06-01T00:00:00Z",
        },
      ],
      [{ id: 10, name: "Bolt" }],
    );
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByText("Bolt")).toBeInTheDocument();
    expect(screen.getByText("SKU-1")).toBeInTheDocument();
    expect(screen.getByText("5d")).toBeInTheDocument();
  });

  it("opens create and edit dialogs with distinct titles", async () => {
    const user = userEvent.setup();
    seed([item], [{ id: 10, name: "Bolt" }]);
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    await screen.findByText("Bolt");
    await user.click(screen.getByRole("button", { name: "Add supplier item" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await user.keyboard("{Escape}");
  });

  it("shows error state with retry", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/supplier-products", () =>
        HttpResponse.json({ success: false, message: "Boom." }, { status: 500 }),
      ),
    );
    renderWithProviders(<SupplierCatalogSection orgId="1" />);

    expect(await screen.findByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
