import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it } from "vitest";
import { FilterBar, type FilterConfig, type FilterValue } from "@/components/filter-bar";
import { renderWithProviders } from "@/lib/tests";

const config: FilterConfig[] = [
  { id: "search", label: "Search", type: "text", placeholder: "Search" },
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

function HostStatefulFilterBar() {
  const [filters, setFilters] = useState<Record<string, FilterValue>>({
    search: null,
    status: null,
  });
  return (
    <FilterBar
      config={config}
      filters={filters}
      onFilterChange={(id, value) => setFilters((previous) => ({ ...previous, [id]: value }))}
      onReset={() => setFilters({ search: null, status: null })}
    />
  );
}

describe("FilterBar", () => {
  it("renders text and a select filter", async () => {
    const user = userEvent.setup();
    renderWithProviders(<HostStatefulFilterBar />);
    expect(screen.getByLabelText("Search")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Filter" }));
    expect(screen.getByLabelText("Status")).toBeInTheDocument();
  });

  it("collects typed text into the controlled filter", async () => {
    const user = userEvent.setup();
    renderWithProviders(<HostStatefulFilterBar />);
    const input = screen.getByLabelText("Search");
    await user.type(input, "invoice");
    expect(input).toHaveValue("invoice");
  });
});
