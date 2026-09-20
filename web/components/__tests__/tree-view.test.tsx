import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { TreeView } from "@/components/tree-view";
import { renderWithProviders } from "@/lib/tests";

const items = [
  {
    id: "accounts",
    label: "Chart of accounts",
    children: [
      { id: "assets", label: "Assets" },
      {
        id: "liabilities",
        label: "Liabilities",
        children: [{ id: "payable", label: "Accounts payable" }],
      },
    ],
  },
  { id: "products", label: "Item categories", children: [] },
];

describe("TreeView", () => {
  it("renders root nodes and reveals children on expand", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TreeView items={items} aria-label="Ledger tree" />);

    expect(screen.getByText("Chart of accounts")).toBeInTheDocument();
    expect(screen.queryByText("Assets")).not.toBeInTheDocument();

    await user.click(screen.getByText("Chart of accounts"));

    expect(screen.getByText("Assets")).toBeInTheDocument();
    expect(screen.getByText("Liabilities")).toBeInTheDocument();
  });

  it("selects a row on click", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(
      <TreeView items={items} onSelect={onSelect} defaultExpandedIds={["accounts"]} />,
    );

    await user.click(screen.getByText("Assets"));

    expect(onSelect).toHaveBeenCalledWith("assets");
  });

  it("navigates with the keyboard and expands via ArrowRight", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <TreeView items={items} aria-label="Keyboard tree" selectedId="accounts" />,
    );

    const row = screen.getByText("Chart of accounts").closest('[role="treeitem"]') as HTMLElement;
    expect(row.getAttribute("role")).toBe("treeitem");
    row.focus();

    await user.keyboard("{ArrowRight}");
    expect(screen.getByText("Assets")).toBeInTheDocument();

    await user.keyboard("{ArrowDown}");
    expect(screen.getByText("Assets").closest('[role="treeitem"]')).toHaveFocus();

    await user.keyboard("{ArrowLeft}");
    expect(screen.getByText("Chart of accounts").closest('[role="treeitem"]')).toHaveFocus();
  });
});
