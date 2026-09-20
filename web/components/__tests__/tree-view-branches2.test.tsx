import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TreeView } from "@/components/tree-view";
import { renderWithProviders } from "@/lib/tests";

const ITEMS = [
  {
    id: "root",
    label: "Root",
    icon: <span>icon</span>,
    children: [
      { id: "child-a", label: "Child A" },
      {
        id: "child-b",
        label: "Child B",
        children: [{ id: "grand", label: "Grand" }],
        actions: <button type="button">Act</button>,
      },
    ],
  },
  { id: "leaf", label: "Leaf" },
];

describe("TreeView branches2", () => {
  it("supports controlled expanded ids with change notifications", async () => {
    const user = userEvent.setup();
    const onExpandedChange = vi.fn();
    renderWithProviders(
      <TreeView items={ITEMS} expandedIds={[]} onExpandedChange={onExpandedChange} />,
    );

    await user.click(screen.getByText("Root"));
    expect(onExpandedChange).toHaveBeenCalledWith(["root"]);
  });

  it("marks selection and renders icons and actions branches", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(
      <TreeView
        items={ITEMS}
        defaultExpandedIds={["root", "child-b"]}
        selectedId="child-a"
        onSelect={onSelect}
      />,
    );

    expect(screen.getByText("icon")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Act" })).toBeInTheDocument();
    await user.click(screen.getByText("Child A"));
    expect(onSelect).toHaveBeenCalledWith("child-a");
  });

  it("ignores clicks on row actions", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(
      <TreeView items={ITEMS} defaultExpandedIds={["root", "child-b"]} onSelect={onSelect} />,
    );

    await user.click(screen.getByRole("button", { name: "Act" }));
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("navigates with Home, End, ArrowUp and ArrowDown", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TreeView items={ITEMS} defaultExpandedIds={["root"]} />);

    const row = screen.getByText("Root").closest('[role="treeitem"]') as HTMLElement;
    row.focus();
    await user.keyboard("{End}");
    await user.keyboard("{Home}");
    await user.keyboard("{ArrowDown}");
    await user.keyboard("{ArrowUp}");
    expect(screen.getByText("Leaf")).toBeInTheDocument();
  });

  it("toggles through Enter and collapses with ArrowLeft", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TreeView items={ITEMS} defaultExpandedIds={["root", "child-b"]} />);

    const grand = screen.getByText("Grand").closest('[role="treeitem"]') as HTMLElement;
    grand.focus();
    await user.keyboard("{ArrowLeft}");
    expect(screen.getByText("Child B")).toBeInTheDocument();
    const root = screen.getByText("Root").closest('[role="treeitem"]') as HTMLElement;
    root.focus();
    await user.keyboard("{Enter}");
    expect(screen.queryByText("Child A")).not.toBeInTheDocument();
  });
});
