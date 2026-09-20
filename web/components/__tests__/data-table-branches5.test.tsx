import type { ColumnDef } from "@tanstack/react-table";
import { screen } from "@testing-library/react";
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
  { accessorKey: "age", header: "Age" },
];

describe("DataTable branches5", () => {
  it("notifies pagination changes in controlled mode", async () => {
    const user = userEvent.setup();
    const onPaginationChange = vi.fn();
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enablePagination
        pagination={{ pageIndex: 0, pageSize: 5 }}
        onPaginationChange={onPaginationChange}
      />,
    );

    await user.click(screen.getByRole("button", { name: "Next page" }));

    expect(onPaginationChange).toHaveBeenCalledWith(expect.objectContaining({ pageIndex: 1 }));
  });

  it("falls back to data length for manual pagination without a row count", () => {
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enablePagination
        manualPagination
      />,
    );

    expect(screen.getByText("1–10 of 12")).toBeInTheDocument();
  });

  it("shows empty state when column filters remove every row", () => {
    renderWithProviders(
      <DataTable
        columns={COLUMNS}
        data={ROWS}
        getRowId={(row) => row.id}
        enableColumnFilters
        initialColumnFilters={[{ id: "name", value: "zzz-no-match" }]}
      />,
    );

    expect(screen.getByText("No records found")).toBeInTheDocument();
    expect(screen.queryByText("User 01")).toBeNull();
  });

  it("renders grouped columns with placeholder headers", () => {
    const grouped: ColumnDef<Row>[] = [
      {
        header: "People",
        columns: [
          { accessorKey: "name", header: "Name" },
          { accessorKey: "age", header: "Age" },
        ],
      },
    ];
    renderWithProviders(<DataTable columns={grouped} data={ROWS} getRowId={(row) => row.id} />);

    expect(screen.getByText("People")).toBeInTheDocument();
    expect(screen.getByText("User 01")).toBeInTheDocument();
  });
});
