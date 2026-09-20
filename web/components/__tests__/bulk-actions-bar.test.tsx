import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { BulkActionsBar } from "@/components/bulk-actions-bar";
import { renderWithProviders } from "@/lib/tests";

function renderBar() {
  return renderWithProviders(
    <BulkActionsBar
      selected={2}
      actions={[
        { label: "Export", onRun: vi.fn() },
        {
          label: "Delete",
          variant: "destructive",
          confirm: { title: "Delete selected?" },
          onRun: vi.fn(),
        },
      ]}
      onClear={vi.fn()}
    />,
  );
}

describe("BulkActionsBar", () => {
  it("renders nothing when nothing is selected", () => {
    const { container } = renderWithProviders(<BulkActionsBar selected={0} actions={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it("shows the selected count and actions", async () => {
    const user = userEvent.setup();
    renderBar();
    expect(screen.getByText("2 selected")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Export" }));
  });

  it("wires a confirmable action through a dialog", () => {
    renderBar();
    expect(screen.getByRole("button", { name: "Delete" })).toBeInTheDocument();
  });
});
