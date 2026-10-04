import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { PaymentDialog } from "@/components/payment-dialog";
import { renderWithProviders } from "@/lib/tests";

const journals = [
  { id: 1, name: "Cash", type: "cash" },
  { id: 2, name: "Bank", type: "bank" },
] as never;

function renderDialog(overrides = {}) {
  const onOpenChange = vi.fn();
  const onConfirm = vi.fn();
  renderWithProviders(
    <PaymentDialog
      title="Record payment"
      description="Choose a journal"
      confirmLabel="Save"
      open
      onOpenChange={onOpenChange}
      onConfirm={onConfirm}
      journals={journals}
      {...overrides}
    />,
  );
  return { onOpenChange, onConfirm };
}

async function pickJournal(offset: number) {
  const user = userEvent.setup();
  await user.click(screen.getByRole("combobox", { name: "Jurnal" }));
  await screen.findByRole("option", { name: "Cash — cash" });
  for (let step = 0; step < offset; step++) {
    await user.keyboard("{ArrowDown}");
  }
  await user.keyboard("{Enter}");
  return user;
}

describe("PaymentDialog", () => {
  it("renders the title and keeps confirm disabled until a journal is picked", async () => {
    renderDialog();
    expect(screen.getByText("Record payment")).toBeInTheDocument();
    expect(screen.getByText("Choose a journal")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    await pickJournal(1);
    expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
  });

  it("confirms with the selected journal and date", async () => {
    const { onConfirm } = renderDialog();
    await pickJournal(2);
    await userEvent.setup().click(screen.getByRole("button", { name: "Save" }));
    expect(onConfirm).toHaveBeenCalledWith(
      expect.objectContaining({ journalId: "2", date: expect.any(String) }),
    );
  });

  it("shows the amount field and requires it when enabled", async () => {
    renderDialog({ showAmount: true });
    const amount = screen.getByRole("spinbutton");
    const user = await pickJournal(1);
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    await user.type(amount, "150000");
    expect(screen.getByRole("button", { name: "Save" })).toBeEnabled();
  });

  it("hides the amount field by default", () => {
    renderDialog();
    expect(screen.queryByRole("spinbutton")).not.toBeInTheDocument();
  });

  it("closes on cancel", async () => {
    const { onOpenChange } = renderDialog();
    await userEvent.setup().click(screen.getByRole("button", { name: "Batal" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
