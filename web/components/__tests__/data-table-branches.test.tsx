import type { ColumnDef } from "@tanstack/react-table";
import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "@/components/data-table";
import { renderWithProviders } from "@/lib/tests";

type Row = { id: string; name: string; age: number };

const ROWS: Row[] = Array.from({ length: 25 }, (_, index) => ({
  id: String(index + 1),
  name: `User ${String(index + 1).padStart(2, "0")}`,
  age: 20 + (index % 10),
}));

const COLUMNS: ColumnDef<Row>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "age", header: "Age", meta: { align: "right" } },
];

function baseProps(overrides = {}) {
  return {
    columns: COLUMNS,
    data: ROWS,
    getRowId: (row: Row) => row.id,
    ...overrides,
  };
}

describe("DataTable branches", () => {
  it("sorts ascending then descending through the header button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DataTable {...baseProps({ enableSorting: true })} />);
    const sortButton = screen.getByRole("button", { name: "Sort by Name" });
    await user.click(sortButton);
    const table = screen.getByRole("table");
    expect(within(table).getAllByRole("row")[1]).toHaveTextContent("User 01");
    await user.click(sortButton);
    expect(within(table).getAllByRole("row")[1]).toHaveTextContent("User 25");
  });

  it("notifies controlled sorting changes without internal state", async () => {
    const user = userEvent.setup();
    const onSortingChange = vi.fn();
    renderWithProviders(
      <DataTable {...baseProps({ enableSorting: true, sorting: [], onSortingChange })} />,
    );
    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    expect(onSortingChange).toHaveBeenCalled();
  });

  it("selects rows and reports selection", async () => {
    const user = userEvent.setup();
    const onSelectionChange = vi.fn();
    renderWithProviders(
      <DataTable
        {...baseProps({
          data: ROWS.slice(0, 3),
          enableRowSelection: true,
          onSelectionChange,
        })}
      />,
    );
    await user.click(screen.getByRole("checkbox", { name: "Select row 1" }));
    expect(onSelectionChange).toHaveBeenCalledWith(["1"]);
    await user.click(screen.getByRole("checkbox", { name: "Select all rows" }));
    expect(onSelectionChange).toHaveBeenCalled();
  });

  it("paginates forward and backward with range text", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DataTable
        {...baseProps({ enablePagination: true, initialPageSize: 10 })}
        labels={{ pagePrevious: "Prev", pageNext: "Next" }}
      />,
    );
    expect(screen.getByText("Page 1 of 3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Next" }));
    expect(await screen.findByText("Page 2 of 3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Prev" }));
    expect(await screen.findByText("Page 1 of 3")).toBeInTheDocument();
  });

  it("renders manual pagination page counts", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          enablePagination: true,
          manualPagination: true,
          pageCount: 7,
          rowCount: 70,
        })}
      />,
    );
    expect(screen.getByText("Page 1 of 7")).toBeInTheDocument();
  });

  it("renders controlled pagination state", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          enablePagination: true,
          pagination: { pageIndex: 1, pageSize: 10 },
        })}
      />,
    );
    expect(screen.getByText("Page 2 of 3")).toBeInTheDocument();
  });

  it("shows the loading skeleton with default rows", () => {
    const { container } = renderWithProviders(
      <DataTable {...baseProps({ status: { type: "loading" } })} />,
    );
    expect(container.querySelector('[data-slot="data-table"]')).toBeInTheDocument();
    expect(screen.getByRole("table")).toBeInTheDocument();
  });

  it("shows the loading skeleton with custom rows", () => {
    renderWithProviders(<DataTable {...baseProps({ status: { type: "loading", rows: 1 } })} />);
    expect(screen.getByRole("table")).toBeInTheDocument();
  });

  it("shows error with retry and default message", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    renderWithProviders(<DataTable {...baseProps({ status: { type: "error", onRetry } })} />);
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalled();
  });

  it("shows error with a custom message and no retry", () => {
    renderWithProviders(
      <DataTable {...baseProps({ status: { type: "error", message: "Broken" } })} />,
    );
    expect(screen.getByText("Broken")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Retry" })).not.toBeInTheDocument();
  });

  it("shows empty state with custom title and description", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          data: [],
          status: { type: "empty", title: "Nothing here", description: "Add one" },
        })}
      />,
    );
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
    expect(screen.getByText("Add one")).toBeInTheDocument();
  });

  it("shows empty state for zero rows without status", () => {
    renderWithProviders(<DataTable {...baseProps({ data: [] })} />);
    expect(screen.getByRole("table")).toBeInTheDocument();
  });

  it("renders toolbar content", () => {
    renderWithProviders(
      <DataTable {...baseProps({ data: ROWS.slice(0, 2), toolbar: <span>Toolbar item</span> })} />,
    );
    expect(screen.getByText("Toolbar item")).toBeInTheDocument();
  });

  it("renders rows with selection state", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DataTable {...baseProps({ data: ROWS.slice(0, 2), enableRowSelection: true })} />,
    );
    await user.click(screen.getByRole("checkbox", { name: "Select row 1" }));
    await waitFor(() =>
      expect(screen.getByRole("checkbox", { name: "Select row 1" })).toBeChecked(),
    );
  });

  it("filters rows through controlled column filters", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          enableColumnFilters: true,
          columnFilters: [{ id: "name", value: "User 01" }],
        })}
      />,
    );
    expect(screen.getByText("User 01")).toBeInTheDocument();
    expect(screen.queryByText("User 02")).not.toBeInTheDocument();
  });

  it("filters rows through the global filter", () => {
    renderWithProviders(
      <DataTable {...baseProps({ enableGlobalFilter: true, globalFilter: "User 02" })} />,
    );
    expect(screen.getByText("User 02")).toBeInTheDocument();
    expect(screen.queryByText("User 03")).not.toBeInTheDocument();
  });

  it("hides columns through visibility state", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          data: ROWS.slice(0, 1),
          enableHiding: true,
          columnVisibility: { age: false },
        })}
      />,
    );
    expect(screen.getByText("User 01")).toBeInTheDocument();
    expect(screen.queryByText("Age")).not.toBeInTheDocument();
  });

  it("uses column id when header is not a string", () => {
    const columns: ColumnDef<Row>[] = [{ accessorKey: "name", header: () => "Fancy" }];
    renderWithProviders(
      <DataTable
        columns={columns}
        data={ROWS.slice(0, 1)}
        getRowId={(row) => row.id}
        enableSorting
      />,
    );
    expect(screen.getByRole("button", { name: "Sort by name" })).toBeInTheDocument();
  });

  it("centers aligned columns", () => {
    const columns: ColumnDef<Row>[] = [
      { accessorKey: "name", header: "Name", meta: { align: "center", className: "extra" } },
    ];
    renderWithProviders(
      <DataTable columns={columns} data={ROWS.slice(0, 1)} getRowId={(row) => row.id} />,
    );
    expect(screen.getByText("Name")).toBeInTheDocument();
  });

  it("renders placeholders without headers", () => {
    const columns: ColumnDef<Row>[] = [
      { id: "group", header: "Group", columns: [{ accessorKey: "name", header: "Name" }] },
    ];
    renderWithProviders(
      <DataTable columns={columns} data={ROWS.slice(0, 1)} getRowId={(row) => row.id} />,
    );
    expect(screen.getByText("Name")).toBeInTheDocument();
  });

  it("supports manual sorting and filtering flags", () => {
    renderWithProviders(
      <DataTable
        {...baseProps({
          data: ROWS.slice(0, 2),
          enableSorting: true,
          manualSorting: true,
          manualFiltering: true,
          enableColumnFilters: true,
        })}
      />,
    );
    expect(screen.getByText("User 01")).toBeInTheDocument();
  });
});
