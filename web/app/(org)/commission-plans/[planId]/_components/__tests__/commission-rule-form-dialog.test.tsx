import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { CommissionRuleFormDialog } from "../commission-rule-form-dialog";

describe("CommissionRuleFormDialog", () => {
  it("renders amount and rate fields when open", async () => {
    renderWithProviders(
      <CommissionRuleFormDialog
        open
        onOpenChange={() => {}}
        orgId="1"
        planId="1"
        onSave={() => {}}
      />,
    );

    expect(await screen.findByText("Add commission rule")).toBeInTheDocument();
    expect(screen.getByLabelText("Min amount")).toBeInTheDocument();
    expect(screen.getByLabelText("Rate %")).toBeInTheDocument();
  });

  it("closes without saving from the cancel button", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    renderWithProviders(
      <CommissionRuleFormDialog
        open
        onOpenChange={onOpenChange}
        orgId="1"
        planId="1"
        onSave={() => {}}
      />,
    );

    await user.click(await screen.findByRole("button", { name: "Cancel" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
