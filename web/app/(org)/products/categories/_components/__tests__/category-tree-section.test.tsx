import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CategoryTreeSection } from "../category-tree-section";

const STAMP = "2026-01-01T00:00:00Z";

const categories = [
  {
    id: 1,
    name: "Raw Materials",
    parent_id: null,
    income_account_id: null,
    expense_account_id: null,
    stock_cost_account_id: null,
    stock_input_account_id: null,
    stock_output_account_id: null,
    cogs_account_id: null,
    cost_method: null,
    valuation: null,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    name: "Metals",
    parent_id: 1,
    income_account_id: null,
    expense_account_id: null,
    stock_cost_account_id: null,
    stock_input_account_id: null,
    stock_output_account_id: null,
    cogs_account_id: null,
    cost_method: "average",
    valuation: "automated",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { accounts: [{ id: 10, name: "COGS Expense" }] },
      }),
    ),
  );
});

describe("CategoryTreeSection", () => {
  it("renders seeded categories as a tree", async () => {
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    expect(await screen.findByText("Raw Materials")).toBeInTheDocument();
    expect(screen.getByText("Metals")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Tambah kategori" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Kategori baru")).toBeInTheDocument();
  });
});
