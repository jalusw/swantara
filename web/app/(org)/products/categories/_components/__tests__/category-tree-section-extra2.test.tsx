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

function seedCategories(list = categories, delayMs = 0) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/item-categories", async () => {
      if (delayMs > 0) await new Promise((resolve) => setTimeout(resolve, delayMs));
      return HttpResponse.json({ success: true, message: "OK.", data: { categories: list } });
    }),
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

describe("CategoryTreeSection extra2", () => {
  it("shows a loading state while categories load", async () => {
    seedCategories(categories, 50);
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    expect(await screen.findByText("Loading...")).toBeInTheDocument();
    expect(await screen.findByText("Raw Materials")).toBeInTheDocument();
  });

  it("shows the empty state when no categories exist", async () => {
    seedCategories([]);
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    expect(await screen.findByText("No categories yet.")).toBeInTheDocument();
  });

  it("opens the edit dialog prefilled with the category", async () => {
    seedCategories();
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Metals");
    const editButtons = screen.getAllByRole("button", { name: "Edit" });
    await user.click(editButtons[1]!);
    const dialog = await screen.findByRole("dialog");

    expect(within(dialog).getByText("Edit category")).toBeInTheDocument();
    expect(within(dialog).getByLabelText("Name")).toHaveValue("Metals");
  });

  it("blocks deleting a parent that still has children", async () => {
    seedCategories();
    let deleteCalls = 0;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/item-categories/:id", () => {
        deleteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Raw Materials");
    const deleteButtons = screen.getAllByRole("button", { name: "Delete" });
    await user.click(deleteButtons[0]!);
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Delete" }));

    expect(deleteCalls).toBe(0);
    expect(screen.getByText("Raw Materials")).toBeInTheDocument();
  });

  it("deletes a leaf category after confirmation", async () => {
    seedCategories();
    let deleteCalls = 0;
    server.use(
      http.delete("*/api/v1/organizations/:organizationId/item-categories/:id", () => {
        deleteCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Metals");
    const deleteButtons = screen.getAllByRole("button", { name: "Delete" });
    await user.click(deleteButtons[1]!);
    const dialog = await screen.findByRole("alertdialog");
    await user.click(within(dialog).getByRole("button", { name: "Delete" }));

    await waitFor(() => expect(deleteCalls).toBe(1));
  });

  it("creates a category from the dialog", async () => {
    seedCategories();
    const createBodies: unknown[] = [];
    server.use(
      http.post("*/api/v1/organizations/:organizationId/item-categories", async ({ request }) => {
        createBodies.push(await request.json());
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    const dialog = await screen.findByRole("dialog");
    await user.type(within(dialog).getByLabelText("Name"), "Plastics");
    await user.click(within(dialog).getByRole("button", { name: "Save category" }));

    await waitFor(() => expect(createBodies).toHaveLength(1));
    expect(createBodies[0]).toMatchObject({ name: "Plastics" });
  });

  it("updates a category from the edit dialog", async () => {
    seedCategories();
    let updateCalls = 0;
    server.use(
      http.put("*/api/v1/organizations/:organizationId/item-categories/:id", () => {
        updateCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Metals");
    const editButtons = screen.getAllByRole("button", { name: "Edit" });
    await user.click(editButtons[1]!);
    const dialog = await screen.findByRole("dialog");
    const nameInput = within(dialog).getByLabelText("Name");
    await user.clear(nameInput);
    await user.type(nameInput, "Alloys");
    await user.click(within(dialog).getByRole("button", { name: "Save category" }));

    await waitFor(() => expect(updateCalls).toBe(1));
  });

  it("requires a name before saving", async () => {
    seedCategories();
    const user = userEvent.setup();
    renderWithProviders(<CategoryTreeSection orgId="1" />);

    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    const dialog = await screen.findByRole("dialog");
    await user.click(within(dialog).getByRole("button", { name: "Save category" }));

    expect(await within(dialog).findByText("Enter a category name.")).toBeInTheDocument();
  });
});
