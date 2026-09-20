import type { ColumnDef } from "@tanstack/react-table";
import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { DataTable } from "@/components/data-table";
import { renderWithProviders } from "@/lib/tests";

type Row = { id: string; name: string; age: number };

const ROWS: Row[] = Array.from({ length: 8 }, (_, index) => ({
  id: String(index + 1),
  name: `User ${String(index + 1).padStart(2, "0")}`,
  age: 20 + (index % 10),
}));

const COLUMNS: ColumnDef<Row>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "age", header: "Age", meta: { align: "right" } },
];

const STATIC_COLUMNS: ColumnDef<Row>[] = [
  { accessorKey: "name", header: "Name", enableSorting: false },
  { accessorKey: "age", header: "Age" },
];

function baseProps(overrides: Record<string, unknown> = {}) {
  return {
    columns: COLUMNS,
    data: ROWS,
    getRowId: (row: Row) => row.id,
    ...overrides,
  };
}

describe("DataTable branches3", () => {
  it("applies initial column filters to narrow rows", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          enableColumnFilters: true,
          initialColumnFilters: [{ id: "name", value: "User 01" }],
        })}
      />,
    );
    const table = screen.getByRole("table");
    expect(within(table).getAllByRole("row")).toHaveLength(2);
  });

  it("applies an initial global filter to narrow rows", () => {
    renderWithProviders(
      <DataTable {...baseProps({ enableGlobalFilter: true, initialGlobalFilter: "User 02" })} />,
    );
    expect(screen.getByRole("table").textContent).toContain("User 02");
    expect(screen.getByRole("table").textContent).not.toContain("User 03");
  });

  it("hides columns from initial visibility state", () => {
    renderWithProviders(
      <DataTable {...baseProps({ enableHiding: true, initialColumnVisibility: { age: false } })} />,
    );
    expect(screen.queryByRole("columnheader", { name: "Age" })).toBeNull();
    expect(screen.getByRole("columnheader", { name: "Name" })).toBeInTheDocument();
  });

  it("selects rows without a selection callback and shows the count", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DataTable
        {...baseProps({ data: ROWS.slice(0, 3), enableRowSelection: true, enablePagination: true })}
      />,
    );
    await user.click(screen.getByRole("checkbox", { name: "Select row 2" }));
    expect(screen.getByText("1 selected")).toBeInTheDocument();
  });

  it("hides the pagination footer while loading", () => {
    renderWithProviders(
      <DataTable {...baseProps({ enablePagination: true, status: { type: "loading" } })} />,
    );
    expect(screen.queryByText("Page 1 of 1")).toBeNull();
  });

  it("renders default empty copy when the status carries no title", () => {
    renderWithProviders(<DataTable {...baseProps({ data: [], status: { type: "empty" } })} />);
    expect(screen.getByText("No records found")).toBeInTheDocument();
  });

  it("renders plain headers for columns that opt out of sorting", () => {
    renderWithProviders(
      <DataTable {...baseProps({ columns: STATIC_COLUMNS, enableSorting: true })} />,
    );
    expect(screen.queryByRole("button", { name: "Sort by Name" })).toBeNull();
    expect(screen.getByRole("button", { name: "Sort by Age" })).toBeInTheDocument();
  });

  it("merges partial label overrides over translated defaults", () => {
    renderWithProviders(
      <DataTable {...baseProps({ data: [], labels: { emptyTitle: "Nothing here" } })} />,
    );
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });
});
