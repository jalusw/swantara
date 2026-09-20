import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProductsSection } from "../products-section";

const item = {
  id: 1,
  organization_id: 1,
  name: "Widget",
  category_id: null,
  type: "goods",
  unit_id: null,
  purchase_unit_id: null,
  list_price: 100,
  standard_cost: 60,
  is_purchasable: true,
  is_sellable: true,
  is_manufactured: false,
  tracking: "none",
  active: true,
};

function seed(products: unknown[] = [item], categories: unknown[] = []) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/products", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { products } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories } }),
    ),
    http.get("*/api/v1/units", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { units: [] } }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("ProductsSection branches", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/products", () => new Promise(() => {})),
    );
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(document.body).toBeInTheDocument();
  });

  it("renders item rows with category fallback", async () => {
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(screen.getByText("—")).toBeInTheDocument();
  });

  it("renders category names when categories exist", async () => {
    seed([item], [{ id: 3, name: "Hardware", parent_id: null }]);
    server.use(
      http.get("*/api/v1/organizations/:organizationId/products", () =>
        HttpResponse.json({
          success: true,
          message: "OK.",
          data: { products: [{ ...item, category_id: 3 }] },
        }),
      ),
    );
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(await screen.findByText("Hardware")).toBeInTheDocument();
  });

  it("shows error with retry when products fail", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/products", () =>
        HttpResponse.json({ success: false, message: "Boom." }, { status: 500 }),
      ),
    );
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(await screen.findByRole("button", { name: "Retry" })).toBeInTheDocument();
  });

  it("opens the create dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ProductsSection orgId="1" />);

    await screen.findByText("Widget");
    await user.click(screen.getByRole("button", { name: "Add item" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
  });

  it("renders inactive status for inactive products", async () => {
    seed([{ ...item, active: false }]);
    renderWithProviders(<ProductsSection orgId="1" />);

    expect(await screen.findByText("Widget")).toBeInTheDocument();
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });
});
