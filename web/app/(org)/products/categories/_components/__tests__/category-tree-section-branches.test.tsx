import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CategoryTreeSection } from "../category-tree-section";

const STAMP = "2026-01-01T00:00:00Z";

const CATEGORIES = [
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
    income_account_id: 10,
    expense_account_id: 11,
    stock_cost_account_id: 12,
    stock_input_account_id: 13,
    stock_output_account_id: 14,
    cogs_account_id: 15,
    cost_method: "average",
    valuation: "automated",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

const ACCOUNTS = [
  { id: 10, name: "Sales Income" },
  { id: 11, name: "COGS Expense" },
];

function useHandlers(options?: { categories?: unknown[]; failCreate?: boolean }) {
  const posted: unknown[] = [];
  const put: unknown[] = [];
  const deleted: string[] = [];
  server.use(
    http.get("*/api/v1/organizations/:organizationId/item-categories", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { categories: options?.categories ?? CATEGORIES },
      }),
    ),
    http.get("*/api/v1/organizations/:organizationId/accounts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { accounts: ACCOUNTS } }),
    ),
    http.post("*/api/v1/organizations/:organizationId/item-categories", async ({ request }) => {
      posted.push(await request.json());
      if (options?.failCreate) {
        return HttpResponse.json({ success: false, message: "Nope." }, { status: 422 });
      }
      return HttpResponse.json(
        { success: true, message: "Created.", data: { category: { id: 9 } } },
        { status: 201 },
      );
    }),
    http.put("*/api/v1/organizations/:organizationId/item-categories/:id", async ({ request }) => {
      put.push(await request.json());
      return HttpResponse.json({ success: true, message: "OK.", data: {} });
    }),
    http.delete("*/api/v1/organizations/:organizationId/item-categories/:id", ({ params }) => {
      deleted.push(String(params.id));
      return HttpResponse.json({ success: true, message: "Deleted." });
    }),
  );
  return { posted, put, deleted };
}

beforeEach(() => {});

describe("CategoryTreeSection branches", () => {
  it("renders the empty state without categories", async () => {
    useHandlers({ categories: [] });
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    expect(await screen.findByText("No categories yet.")).toBeInTheDocument();
  });

  it("closes the create dialog through cancel", async () => {
    const user = userEvent.setup();
    useHandlers();
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    await screen.findByRole("dialog");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => expect(screen.queryByRole("dialog")).not.toBeInTheDocument());
  });

  it("creates a category with null defaults", async () => {
    const user = userEvent.setup();
    const { posted } = useHandlers();
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    await user.type(await screen.findByLabelText("Name"), "Packaging");
    await user.click(screen.getByRole("button", { name: "Save category" }));
    await waitFor(() => expect(posted.length).toBe(1));
    const body = posted[0] as Record<string, unknown>;
    expect(body.name).toBe("Packaging");
    expect(body.parent_id).toBeNull();
    expect(body.cost_method).toBeNull();
    expect(body.income_account_id).toBeNull();
  });

  it("keeps the dialog open when creation fails", async () => {
    const user = userEvent.setup();
    useHandlers({ failCreate: true });
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Raw Materials");
    await user.click(screen.getByRole("button", { name: "Add category" }));
    await user.type(await screen.findByLabelText("Name"), "Broken");
    await user.click(screen.getByRole("button", { name: "Save category" }));
    await waitFor(() => expect(screen.getByRole("dialog")).toBeInTheDocument());
    expect(screen.getByLabelText("Name")).toHaveValue("Broken");
  });

  it("edits a category with prefilled values and updates it", async () => {
    const user = userEvent.setup();
    const { put } = useHandlers();
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Metals");
    const edits = screen.getAllByRole("button", { name: "Edit" });
    const lastEdit = edits[edits.length - 1];
    if (lastEdit === undefined) throw new Error("expected an edit button");
    await user.click(lastEdit);
    expect(await screen.findByText("Edit category")).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("Metals");
    await user.click(screen.getByRole("button", { name: "Save category" }));
    await waitFor(() => expect(put.length).toBe(1));
  });

  it("blocks deletion of a parent category", async () => {
    const user = userEvent.setup();
    const { deleted } = useHandlers();
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Raw Materials");
    const deletes = screen.getAllByRole("button", { name: "Delete" });
    const firstDelete = deletes[0];
    if (firstDelete === undefined) throw new Error("expected a delete button");
    await user.click(firstDelete);
    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Delete" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).not.toBeInTheDocument());
    expect(deleted.length).toBe(0);
    expect(screen.getByText("Raw Materials")).toBeInTheDocument();
  });

  it("deletes a leaf category", async () => {
    const user = userEvent.setup();
    const { deleted } = useHandlers();
    renderWithProviders(<CategoryTreeSection orgId="1" />);
    await screen.findByText("Metals");
    const deletes = screen.getAllByRole("button", { name: "Delete" });
    const lastDelete = deletes[deletes.length - 1];
    if (lastDelete === undefined) throw new Error("expected a delete button");
    await user.click(lastDelete);
    const confirm = await screen.findByRole("alertdialog");
    await user.click(within(confirm).getByRole("button", { name: "Delete" }));
    await waitFor(() => expect(deleted).toEqual(["2"]));
  });
});
