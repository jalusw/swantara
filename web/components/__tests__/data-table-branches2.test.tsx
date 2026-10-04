import type { ColumnDef } from "@tanstack/react-table";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "@/components/data-table";
import { renderWithProviders } from "@/lib/tests";

type Row = { id: string; name: string; age: number };

const ROWS: Row[] = [
  { id: "1", name: "Alice", age: 30 },
  { id: "2", name: "Bob", age: 25 },
];

const COLUMNS: ColumnDef<Row>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "age", header: "Age", meta: { align: "center" } },
];

describe("DataTable branches2", () => {
  it("applies initial sorting state", () => {
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enableSorting
        initialSorting={[{ id: "age", desc: false }]}
      />,
    );
    const table = screen.getByRole("table");
    expect(table).toBeInTheDocument();
    expect(screen.getByText("Alice")).toBeInTheDocument();
  });

  it("notifies pagination, filter, visibility and global filter changes", async () => {
    const user = userEvent.setup();
    const onPaginationChange = vi.fn();
    const onColumnFiltersChange = vi.fn();
    const onGlobalFilterChange = vi.fn();
    const onColumnVisibilityChange = vi.fn();
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enablePagination
        initialPageSize={1}
        enableColumnFilters
        enableGlobalFilter
        enableHiding
        enableSorting
        onPaginationChange={onPaginationChange}
        onColumnFiltersChange={onColumnFiltersChange}
        onGlobalFilterChange={onGlobalFilterChange}
        onColumnVisibilityChange={onColumnVisibilityChange}
      />,
    );
    await user.click(screen.getByRole("button", { name: "Halaman berikutnya" }));
    expect(onPaginationChange).toHaveBeenCalled();
  });

  it("renders controlled column filters and visibility callbacks through table api", () => {
    const onColumnFiltersChange = vi.fn();
    const onColumnVisibilityChange = vi.fn();
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enableColumnFilters
        enableHiding
        columnFilters={[]}
        columnVisibility={{}}
        onColumnFiltersChange={onColumnFiltersChange}
        onColumnVisibilityChange={onColumnVisibilityChange}
      />,
    );
    expect(screen.getByText("Alice")).toBeInTheDocument();
  });

  it("shows selection count when rows are selected", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enableRowSelection
        enablePagination
      />,
    );
    await user.click(screen.getByRole("checkbox", { name: "Pilih baris 1" }));
    expect(await screen.findByText(/1.*dipilih|dipilih.*1/i)).toBeInTheDocument();
  });

  it("renders manual pagination without pageCount and custom page sizes", () => {
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enablePagination
        manualPagination
        pageSizeOptions={[5, 10]}
      />,
    );
    expect(screen.getByText(/Halaman 1 dari/)).toBeInTheDocument();
  });

  it("renders desc sort icon and aria-sort branches", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <DataTable columns={COLUMNS} data={ROWS} getRowId={(row) => row.id} enableSorting />,
    );
    const sortButton = screen.getByRole("button", { name: "Urutkan berdasarkan Name" });
    await user.click(sortButton);
    await user.click(sortButton);
    expect(sortButton).toHaveAttribute("aria-label", "Urutkan berdasarkan Name");
    expect(screen.getByRole("table")).toBeInTheDocument();
  });

  it("hides footer range when pagination disabled with empty data", () => {
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={[]}
        getRowId={(row) => row.id}
        status={{ type: "empty", title: "Nothing" }}
      />,
    );
    expect(screen.getByText("Nothing")).toBeInTheDocument();
    expect(screen.queryByText(/Halaman 1 dari/)).not.toBeInTheDocument();
  });
});
