import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { PriceBooksSection } from "../price-books-section";

type PriceBookRow = {
  id: number;
  name: string;
  currency_code: string | null;
  organization_id: number;
  active: boolean;
  created_at: string;
  updated_at: string;
};

const seedRows: PriceBookRow[] = [
  {
    id: 1,
    name: "Retail",
    currency_code: "USD",
    organization_id: 1,
    active: true,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
  {
    id: 2,
    name: "Wholesale IDR",
    currency_code: "IDR",
    organization_id: 1,
    active: false,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
  },
];

let rows: PriceBookRow[];

function usePriceBookHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/price_books", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { price_books: rows } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/price_books", async ({ request }) => {
      const body = (await request.json()) as { name: string; currency_code?: string | null };
      const created: PriceBookRow = {
        id: rows.length + 1,
        name: body.name,
        currency_code: body.currency_code ?? null,
        organization_id: 1,
        active: true,
        created_at: "2026-01-01T00:00:00Z",
        updated_at: "2026-01-01T00:00:00Z",
      };
      rows = [...rows, created];
      return HttpResponse.json(
        { success: true, message: "Dibuat.", data: { price_book: created } },
        { status: 201 },
      );
    }),
  );
}

beforeEach(() => {
  rows = [...seedRows];
  usePriceBookHandlers();
});

describe("PriceBooksSection", () => {
  it("renders price_books with currency and status", async () => {
    renderWithProviders(<PriceBooksSection orgId="1" />);

    expect(await screen.findByText("Retail")).toBeInTheDocument();
    expect(screen.getByText("Wholesale IDR")).toBeInTheDocument();
    expect(screen.getByText("USD")).toBeInTheDocument();
    expect(screen.getByText("Aktif")).toBeInTheDocument();
    expect(screen.getByText("Tidak aktif")).toBeInTheDocument();
  });

  it("filters price_books through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PriceBooksSection orgId="1" />);

    await screen.findByText("Retail");
    await user.type(screen.getByPlaceholderText("Cari daftar harga"), "Wholesale");

    expect(await screen.findByText("Wholesale IDR")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Retail")).not.toBeInTheDocument());
  });

  it("creates a price_book from the dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<PriceBooksSection orgId="1" />);

    await screen.findByText("Retail");
    await user.click(screen.getByRole("button", { name: "Tambah daftar harga" }));
    await user.type(await screen.findByLabelText("Nama"), "Distributor");
    await user.click(screen.getByRole("button", { name: "Simpan" }));

    expect(await screen.findByText("Distributor")).toBeInTheDocument();
  });
});
