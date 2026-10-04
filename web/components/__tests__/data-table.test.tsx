import type { ColumnDef } from "@tanstack/react-table";
import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "@/components/data-table";
import { renderWithProviders } from "@/lib/tests";

vi.mock("@tanstack/react-table", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-table")>();
  return {
    ...actual,
    useReactTable: vi.fn(() => ({
      getHeaderGroups: () => [{ id: "hg-1", headers: [] }],
      getRowModel: () => ({ rows: [] }),
      getFilteredRowModel: () => ({ rows: [] }),
      getVisibleLeafColumns: () => [],
      getState: () => ({ pagination: { pageIndex: 0, pageSize: 10 } }),
      getIsSomeRowsSelected: () => false,
      getIsAllPageRowsSelected: () => false,
      getPageCount: () => 1,
      getCanPreviousPage: () => false,
      getCanNextPage: () => false,
    })),
  };
});

type Row = { id: string; name: string };

const columns: ColumnDef<Row, string>[] = [{ accessorKey: "name", header: "Name" }];

describe("DataTable", () => {
  it("renders the table with an aria-label", () => {
    renderWithProviders(
      <DataTable
        columns={columns}
        data={[{ id: "1", name: "Alice" }]}
        getRowId={(row) => row.id}
      />,
    );

    expect(screen.getByRole("table", { name: "Tabel data" })).toBeInTheDocument();
  });

  it("renders with a custom aria-label", () => {
    renderWithProviders(
      <DataTable columns={columns} data={[]} getRowId={(row) => row.id} ariaLabel="Users" />,
    );

    expect(screen.getByRole("table", { name: "Users" })).toBeInTheDocument();
  });

  it("has data-slot attribute", () => {
    const { container } = renderWithProviders(
      <DataTable columns={columns} data={[]} getRowId={(row) => row.id} />,
    );

    expect(container.firstElementChild).toHaveAttribute("data-slot", "data-table");
  });
});
