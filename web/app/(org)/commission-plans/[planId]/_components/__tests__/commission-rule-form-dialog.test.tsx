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

    expect(await screen.findByText("Tambah aturan komisi")).toBeInTheDocument();
    expect(screen.getByLabelText("Jumlah minimum")).toBeInTheDocument();
    expect(screen.getByLabelText("Tarif %")).toBeInTheDocument();
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

    await user.click(await screen.findByRole("button", { name: "Batal" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
