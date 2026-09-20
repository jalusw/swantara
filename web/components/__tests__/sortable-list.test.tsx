import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { SortableList } from "@/components/sortable-list";
import { renderWithProviders } from "@/lib/tests";

type Item = { id: string; label: string };

describe("SortableList", () => {
  it("reorders via the move buttons", async () => {
    const user = userEvent.setup();
    const items: Item[] = [
      { id: "a", label: "A" },
      { id: "b", label: "B" },
    ];
    const onReorder = vi.fn();

    renderWithProviders(
      <SortableList
        items={items}
        onReorder={onReorder}
        render={(item) => <span>{item.label}</span>}
      />,
    );

    const firstRow = screen.getByText("A").closest("li") as HTMLElement;
    await user.click(within(firstRow).getByRole("button", { name: "Move down" }));
    expect(onReorder).toHaveBeenCalledWith([
      { id: "b", label: "B" },
      { id: "a", label: "A" },
    ]);
  });

  it("disables moving the first item up", () => {
    renderWithProviders(
      <SortableList
        items={[{ id: "a", label: "A" }]}
        onReorder={vi.fn()}
        render={(item) => <span>{item.label}</span>}
      />,
    );
    const row = screen.getByText("A").closest("li") as HTMLElement;
    expect(within(row).getByRole("button", { name: "Move up" })).toBeDisabled();
  });
});
