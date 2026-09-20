import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ConfirmDialog } from "@/components/confirm-dialog";
import { renderWithProviders } from "@/lib/tests";

describe("ConfirmDialog", () => {
  it("opens on trigger click and fires onConfirm", async () => {
    const user = userEvent.setup();
    const onConfirm = vi.fn();
    renderWithProviders(
      <ConfirmDialog
        title="Delete invoice?"
        confirmLabel="Delete"
        trigger={<button type="button">Delete invoice</button>}
        onConfirm={onConfirm}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Delete invoice" }));
    expect(screen.getByText("Delete invoice?")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Delete" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });
});
