import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ReorderRuleFormDialog } from "../reorder-rule-form-dialog";

beforeEach(() => {});

describe("ReorderRuleFormDialog", () => {
  it("renders quantity fields when open", async () => {
    renderWithProviders(
      <ReorderRuleFormDialog open onOpenChange={() => {}} orgId="1" onSave={() => {}} />,
    );

    expect(await screen.findByText("New reorder rule")).toBeInTheDocument();
    expect(screen.getByLabelText("Item")).toBeInTheDocument();
    expect(screen.getByLabelText("Min qty")).toBeInTheDocument();
  });

  it("closes without saving from the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <ReorderRuleFormDialog open onOpenChange={onOpenChange} orgId="1" onSave={() => {}} />,
    );

    await user.click(await screen.findByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
