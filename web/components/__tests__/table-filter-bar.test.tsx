import type { ColumnFiltersState } from "@tanstack/react-table";
import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import type { FilterConfig } from "../filter-bar";
import { TableFilterBar } from "../table-filter-bar";

const config: FilterConfig[] = [
  { id: "search", label: "Search", type: "text", placeholder: "Search…" },
  {
    id: "status",
    label: "Status",
    type: "select",
    options: [
      { value: "active", label: "Active" },
      { value: "inactive", label: "Inactive" },
    ],
  },
];

function StatefulTableFilterBar({
  extra,
}: {
  extra?: Partial<React.ComponentProps<typeof TableFilterBar>>;
}) {
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  return (
    <TableFilterBar
      config={config}
      columnFilters={columnFilters}
      onColumnFiltersChange={setColumnFilters}
      globalFilter={globalFilter}
      onGlobalFilterChange={setGlobalFilter}
      onReset={() => {
        setColumnFilters([]);
        setGlobalFilter("");
      }}
      {...extra}
    />
  );
}

describe("TableFilterBar", () => {
  it("renders text and select filters", async () => {
    const user = userEvent.setup();
    renderWithProviders(<StatefulTableFilterBar />);
    expect(screen.getByLabelText("Search")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /filter/i }));
    expect(screen.getByLabelText("Status")).toBeInTheDocument();
  });

  it("dispatches globalFilterChange when text is typed", async () => {
    const user = userEvent.setup();
    renderWithProviders(<StatefulTableFilterBar />);
    await user.type(screen.getByLabelText("Search"), "abc");
    expect(screen.getByLabelText("Search")).toHaveValue("abc");
  });

  it("dispatches columnFiltersChange when select changes", async () => {
    const user = userEvent.setup();
    const onColumnFiltersChange = vi.fn();
    renderWithProviders(
      <TableFilterBar
        config={config}
        columnFilters={[]}
        onColumnFiltersChange={onColumnFiltersChange}
      />,
    );
    await user.click(screen.getByRole("button", { name: /filter/i }));
    await user.selectOptions(screen.getByLabelText("Status"), "active");
    expect(onColumnFiltersChange).toHaveBeenCalled();
  });

  it("has data-slot attribute", () => {
    const { container } = renderWithProviders(
      <TableFilterBar config={config} columnFilters={[]} onColumnFiltersChange={vi.fn()} />,
    );
    expect(container.firstElementChild).toHaveAttribute("data-slot", "table-filter-bar");
  });

  it("reset button clears all filters", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    renderWithProviders(
      <TableFilterBar
        config={config}
        columnFilters={[{ id: "status", value: "active" }]}
        onColumnFiltersChange={vi.fn()}
        globalFilter="test"
        onGlobalFilterChange={vi.fn()}
        onReset={onReset}
      />,
    );
    await user.click(screen.getByRole("button", { name: /hapus semua/i }));
    expect(onReset).toHaveBeenCalledTimes(1);
  });
});
