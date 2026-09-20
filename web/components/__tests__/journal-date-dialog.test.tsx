import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { JournalDateDialog } from "@/components/journal-date-dialog";
import { renderWithProviders } from "@/lib/tests";

const journals = [
  { id: 1, name: "Cash", type: "cash" },
  { id: 2, name: "Bank", type: "bank" },
] as never;

function renderDialog(overrides = {}) {
  const onOpenChange = vi.fn();
  const onConfirm = vi.fn();
  renderWithProviders(
    <JournalDateDialog
      title="Post journal"
      description="Choose a journal and date"
      confirmLabel="Post"
      open
      onOpenChange={onOpenChange}
      onConfirm={onConfirm}
      journals={journals}
      {...overrides}
    />,
  );
  return { onOpenChange, onConfirm };
}

async function pickJournal() {
  const user = userEvent.setup();
  await user.click(screen.getByRole("combobox", { name: "Journal" }));
  await screen.findByRole("option", { name: "Cash — cash" });
  await user.keyboard("{ArrowDown}{Enter}");
  return user;
}

describe("JournalDateDialog", () => {
  it("renders title, description, and keeps confirm disabled until a journal is picked", async () => {
    renderDialog();
    expect(screen.getByText("Post journal")).toBeInTheDocument();
    expect(screen.getByText("Choose a journal and date")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Post" })).toBeDisabled();
    await pickJournal();
    expect(screen.getByRole("button", { name: "Post" })).toBeEnabled();
  });

  it("confirms with the selected journal id and date", async () => {
    const { onConfirm } = renderDialog();
    await pickJournal();
    await userEvent.setup().click(screen.getByRole("button", { name: "Post" }));
    expect(onConfirm).toHaveBeenCalledWith("1", expect.any(String));
  });

  it("updates the date before confirming", async () => {
    const { onConfirm } = renderDialog();
    await pickJournal();
    const dateInput = document.querySelector('input[type="date"]') as HTMLInputElement;
    expect(dateInput).toBeInTheDocument();
    const user = userEvent.setup();
    await user.clear(dateInput);
    await user.type(dateInput, "2026-02-14");
    await user.click(screen.getByRole("button", { name: "Post" }));
    expect(onConfirm).toHaveBeenCalledWith("1", "2026-02-14");
  });

  it("keeps confirm disabled while pending", async () => {
    renderDialog({ isPending: true });
    await pickJournal();
    expect(screen.getByRole("button", { name: "Post" })).toBeDisabled();
  });

  it("closes on cancel with a custom label", async () => {
    const { onOpenChange } = renderDialog({ cancelLabel: "Back" });
    await userEvent.setup().click(screen.getByRole("button", { name: "Back" }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("renders nothing interactive when closed", () => {
    renderDialog({ open: false });
    expect(screen.queryByText("Post journal")).not.toBeInTheDocument();
  });
});
