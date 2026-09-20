import type { ColumnDef } from "@tanstack/react-table";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DataTable } from "../tanstack-table";

type Row = { id: string; name: string; amount: number };

const columns: ColumnDef<Row>[] = [
  {
    accessorKey: "name",
    header: "Name",
    meta: { align: "left" },
  },
  {
    accessorKey: "amount",
    header: "Amount",
    meta: { align: "right" },
  },
];

const rows: Row[] = [
  { id: "b", name: "Zeta", amount: 20 },
  { id: "a", name: "Alpha", amount: 10 },
];

function renderTable(extra?: Partial<React.ComponentProps<typeof DataTable<Row, unknown>>>) {
  return renderWithProviders(
    <DataTable
      data={rows}
      columns={columns}
      getRowId={(row) => row.id}
      ariaLabel="Customers"
      {...extra}
    />,
  );
}

describe("DataTable (TanStack)", () => {
  it("renders headers and rows", () => {
    renderTable();
    expect(screen.getByRole("table", { name: "Customers" })).toBeInTheDocument();
    expect(screen.getByText("Name")).toBeInTheDocument();
    expect(screen.getByText("Alpha")).toBeInTheDocument();
    expect(screen.getByText("Zeta")).toBeInTheDocument();
  });

  it("sorts rows when a sortable header is clicked", async () => {
    const user = userEvent.setup();
    renderTable({ enableSorting: true });
    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    const cells = screen.getAllByRole("cell");
    expect(cells[0]).toHaveTextContent("Alpha");
    expect(cells[2]).toHaveTextContent("Zeta");
  });

  it("reports row selection changes", async () => {
    const user = userEvent.setup();
    const onSelectionChange = vi.fn();
    renderTable({ enableRowSelection: true, onSelectionChange });
    await user.click(screen.getByRole("checkbox", { name: "Select row a" }));
    expect(onSelectionChange).toHaveBeenCalledWith(["a"]);
  });

  it("renders an empty state", () => {
    renderTable({ status: { type: "empty", title: "Nothing here" } });
    expect(screen.getByText("Nothing here")).toBeInTheDocument();
  });

  it("renders a loading state", () => {
    renderTable({ status: { type: "loading" } });
    expect(screen.getByRole("table")).toBeInTheDocument();
  });

  it("paginates when enabled", async () => {
    const user = userEvent.setup();
    renderTable({ enablePagination: true, initialPageSize: 1 });
    expect(screen.getByText("Zeta")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Next page" }));
    expect(screen.getByText("Alpha")).toBeInTheDocument();
  });
});
