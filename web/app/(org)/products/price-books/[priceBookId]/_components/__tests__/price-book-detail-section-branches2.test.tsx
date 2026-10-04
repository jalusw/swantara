import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PriceBookDetail } from "../price-book-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const RULES = [
  {
    id: 1,
    price_book_id: 7,
    applies_to: "all",
    item_id: null,
    category_id: null,
    min_qty: 1,
    compute_type: "fixed",
    fixed_price: null,
    discount_pct: null,
    date_start: null,
    date_end: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const PRODUCTS = [
  {
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
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const CATEGORIES = [{ id: 3, organization_id: 1, name: "Bags" }];

function useHandlers(options?: {
  failProducts?: boolean;
  failCategories?: boolean;
  created?: unknown[];
}) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/price_books/:priceBookId/rules", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rules: RULES } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/products", () => {
      if (options?.failProducts) {
        return HttpResponse.json({ success: false, message: "Products down." }, { status: 500 });
      }
      return HttpResponse.json({ success: true, message: "OK.", data: { products: PRODUCTS } });
    }),
    http.get("*/api/v1/organizations/:organizationId/item-categories", () => {
      if (options?.failCategories) {
        return HttpResponse.json({ success: false, message: "Categories down." }, { status: 500 });
      }
      return HttpResponse.json({
        success: true,
        message: "OK.",
        data: { categories: CATEGORIES },
      });
    }),
    http.post(
      "*/api/v1/organizations/:organizationId/price_books/:priceBookId/rules",
      async ({ request }) => {
        options?.created?.push(await request.json());
        return HttpResponse.json(
          { success: true, message: "Dibuat.", data: { rule: { id: 10 } } },
          { status: 201 },
        );
      },
    ),
  );
}

beforeEach(() => {});

async function openDialog(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole("button", { name: "Tambah aturan" }));
  return await screen.findByRole("dialog");
}

describe("PriceBookDetail branches2", () => {
  it("shows the products error state", async () => {
    useHandlers({ failProducts: true });
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);

    expect(await screen.findByText("Products down.")).toBeInTheDocument();
  });

  it("shows the categories error state", async () => {
    useHandlers({ failCategories: true });
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);

    expect(await screen.findByText("Categories down.")).toBeInTheDocument();
  });

  it("creates a fixed rule scoped to a item", async () => {
    const created: unknown[] = [];
    useHandlers({ created });
    const user = userEvent.setup();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Aturan harga");

    const dialog = await openDialog(user);
    await user.click(within(dialog).getByRole("combobox", { name: "Cakupan" }));
    await user.click(await screen.findByRole("option", { name: "Produk" }));
    await user.click(within(dialog).getByRole("combobox", { name: "Produk" }));
    await user.click(await screen.findByRole("option", { name: "Canvas Tote" }));

    const minQty = within(dialog).getByLabelText("Jumlah minimum");
    await user.clear(minQty);
    await user.type(minQty, "0");
    await user.type(within(dialog).getByLabelText("Harga"), "45");

    const dates = dialog.querySelectorAll('input[type="date"]');
    if (dates.length !== 2) throw new Error("Expected date inputs");
    await user.type(dates[0] as HTMLElement, "2026-01-01");
    await user.type(dates[1] as HTMLElement, "2026-12-31");

    await user.click(within(dialog).getByRole("button", { name: "Simpan aturan" }));

    await waitFor(() => expect(created).toHaveLength(1));
    expect(created[0]).toMatchObject({
      applies_to: "item",
      item_id: 5,
      min_qty: 1,
      compute_type: "fixed",
      fixed_price: 45,
    });
    const dated = created[0] as Record<string, unknown>;
    expect(dated.date_start).toBeTruthy();
    expect(dated.date_end).toBeTruthy();
  });

  it("creates a percent rule scoped to a category", async () => {
    const created: unknown[] = [];
    useHandlers({ created });
    const user = userEvent.setup();
    renderWithProviders(<PriceBookDetail orgId="1" priceBookId="7" />);
    await screen.findByText("Aturan harga");

    const dialog = await openDialog(user);
    await user.click(within(dialog).getByRole("combobox", { name: "Cakupan" }));
    await user.click(await screen.findByRole("option", { name: "Kategori" }));
    await user.click(within(dialog).getByRole("combobox", { name: "Kategori" }));
    await user.click(await screen.findByRole("option", { name: "Bags" }));
    await user.click(within(dialog).getByRole("combobox", { name: "Jenis perhitungan" }));
    await user.click(await screen.findByRole("option", { name: "Persen" }));
    await user.type(within(dialog).getByLabelText("Diskon %"), "10");

    await user.click(within(dialog).getByRole("button", { name: "Simpan aturan" }));

    await waitFor(() => expect(created).toHaveLength(1));
    expect(created[0]).toMatchObject({
      applies_to: "category",
      category_id: 3,
      compute_type: "percent",
      discount_pct: 10,
      item_id: null,
    });
  });
});
