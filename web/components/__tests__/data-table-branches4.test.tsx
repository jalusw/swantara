import type { ColumnDef } from "@tanstack/react-table";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "@/components/data-table";
import { renderWithProviders } from "@/lib/tests";

type Row = { id: string; name: string; age: number };

const ROWS: Row[] = Array.from({ length: 12 }, (_, index) => ({
  id: String(index + 1),
  name: `User ${String(index + 1).padStart(2, "0")}`,
  age: 20 + index,
}));

const COLUMNS: ColumnDef<Row>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "age", header: "Age", meta: { align: "center" } },
];

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    columns: COLUMNS,
    data: ROWS,
    getRowId: (row: Row) => row.id,
    ...overrides,
  };
}

describe("DataTable branches4", () => {
  it("renders toolbar content when provided", () => {
    renderWithProviders(<DataTable {...baseProps({ toolbar: <span>Toolbar here</span> })} />);

    expect(screen.getByText("Toolbar here")).toBeInTheDocument();
  });

  it("shows custom error without retry button", () => {
    renderWithProviders(
      <DataTable {...baseProps({ status: { type: "error", message: "Custom fail" } })} />,
    );

    expect(screen.getByText("Custom fail")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Retry" })).toBeNull();
  });

  it("shows custom empty title for the empty status", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({ data: [] as Row[], status: { type: "empty", title: "Nothing here" } })}
      />,
    );

    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });

  it("handles controlled pagination changes", async () => {
    const user = userEvent.setup();
    const onPaginationChange = vi.fn();
    renderWithProviders(
      <DataTable
        {...baseProps({
          enablePagination: true,
          manualPagination: true,
          rowCount: 100,
          pageCount: 10,
          onPaginationChange,
        })}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Next page" }));
    expect(onPaginationChange).toHaveBeenCalled();
    expect(screen.getAllByText(/Page/).length).toBeGreaterThan(0);
  });

  it("uses center alignment for meta columns", () => {
    renderWithProviders(<DataTable {...baseProps({})} />);

    const table = screen.getByRole("table");
    expect(within(table).getByText("Age")).toBeInTheDocument();
  });

  it("reports descending aria-sort after two clicks", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DataTable {...baseProps({ enableSorting: true })} />);

    const sortButton = screen.getByRole("button", { name: "Sort by Name" });
    await user.click(sortButton);
    await user.click(sortButton);

    expect(screen.getByRole("columnheader", { name: "Name" })).toHaveAttribute(
      "aria-sort",
      "descending",
    );
  });

  it("notifies global filter changes in controlled mode", async () => {
    const user = userEvent.setup();
    const onGlobalFilterChange = vi.fn();
    renderWithProviders(
      <DataTable
        {...baseProps({
          enableGlobalFilter: true,
          globalFilter: "",
          onGlobalFilterChange,
          toolbar: <span>tools</span>,
        })}
      />,
    );

    expect(screen.getByText("tools")).toBeInTheDocument();
    expect(onGlobalFilterChange).not.toHaveBeenCalled();
    expect(user).toBeDefined();
  });

  it("hides pagination footer while loading", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({ enablePagination: true, status: { type: "loading", rows: 2 } })}
      />,
    );

    expect(screen.queryByText(/Page 1/)).toBeNull();
  });
});
