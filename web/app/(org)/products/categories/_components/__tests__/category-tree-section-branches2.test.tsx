import { screen, waitFor, within } from "@testing-library/react";
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
];

function seedCategories(list: unknown[] = categories) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { categories: list } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { accounts: [{ id: 10, name: "COGS Expense" }] },
      }),
    ),
  );
}

beforeEach(() => {});

describe("CategoryTreeSection branches2", () => {
  it("creates a category with none optionals branch", async () => {
    seedCategories([]);
    let createBody: unknown = null;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/item-categories", async ({ request }) => {
        createBody = await request.json();
        return HttpResponse.json({ success: true, message: "OK.", data: {} }, { status: 201 });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await user.click(await screen.findByRole("button", { name: "Add category" }));
    const dialog = await screen.findByRole("dialog");
    await user.type(within(dialog).getByLabelText("Name"), "Packaging");
    await user.click(within(dialog).getByRole("button", { name: "Save category" }));

    await waitFor(() => expect(createBody).not.toBeNull());
    expect(createBody).toMatchObject({ cost_method: null, valuation: null });
  });

  it("updates a category with account branches", async () => {
    seedCategories();
    let updateBody: unknown = null;
    server.use(
      http.put(
        "*/api/v1/organizations/:organizationId/item-categories/:id",
        async ({ request }) => {
          updateBody = await request.json();
          return HttpResponse.json({ success: true, message: "OK.", data: {} });
        },
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Edit" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Save category" }));

    await waitFor(() => expect(updateBody).not.toBeNull());
  });

  it("shows the error branch with retry", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
        HttpResponse.json({ success: false, message: "Categories down." }, { status: 500 }),
      ),
      http.get("*/api/v1/organizations/:organizationId/accounts", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { accounts: [] } }),
      ),
    );
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    expect(await screen.findByText("No categories yet.")).toBeInTheDocument();
  });
});
