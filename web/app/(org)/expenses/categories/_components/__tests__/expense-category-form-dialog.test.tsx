import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ExpenseCategory } from "@/lib/services/swantara";
import { renderWithProviders } from "@/lib/tests";
import { ExpenseCategoryFormDialog } from "../expense-category-form-dialog";

const category = {
  id: 1,
  organizationId: 1,
  name: "Travel",
  expenseAccountId: null,
  defaultTaxIds: [],
} as ExpenseCategory;

beforeEach(() => {});

describe("ExpenseCategoryFormDialog", () => {
  it("renders the create title with name field", async () => {
    renderWithProviders(
      <ExpenseCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        category={null}
        onSave={vi.fn()}
      />,
    );

    expect(await screen.findByText("Kategori baru")).toBeInTheDocument();
    expect(screen.getByLabelText("Nama")).toBeInTheDocument();
  });

  it("seeds the name when editing and accepts changes", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <ExpenseCategoryFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        category={category}
        onSave={vi.fn()}
      />,
    );

    const nameInput = await screen.findByDisplayValue("Travel");
    await user.type(nameInput, " Plus");

    expect(nameInput).toHaveValue("Travel Plus");
  });
});
