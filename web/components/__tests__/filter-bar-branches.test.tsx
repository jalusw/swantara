import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  FilterBar,
  type FilterConfig,
  type FiltersMap,
  type FilterValue,
  useFilters,
} from "@/components/filter-bar";
import { renderHookWithProviders, renderWithProviders } from "@/lib/tests";

const FULL_CONFIG: FilterConfig[] = [
  { id: "search", label: "Search", type: "text", placeholder: "Search items" },
  { id: "nickname", label: "Nickname", type: "text" },
  {
    id: "status",
    label: "Status",
    type: "select",
    options: [
      { value: "active", label: "Active" },
      { value: "archived", label: "Archived" },
    ],
  },
  { id: "period", label: "Period", type: "date-range" },
];

function Host({
  config = FULL_CONFIG,
  initial = {},
  onReset,
  actions,
}: {
  config?: FilterConfig[];
  initial?: FiltersMap;
  onReset?: () => void;
  actions?: React.ReactNode;
}) {
  const [filters, setFilters] = useState<FiltersMap>({ ...initial });
  return (
    <FilterBar
      config={config}
      filters={filters}
      onFilterChange={(id, value) => setFilters((previous) => ({ ...previous, [id]: value }))}
      onReset={onReset}
      actions={actions}
    />
  );
}

describe("FilterBar branches", () => {
  it("types into the search input and clears back to null", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    const { rerender } = renderWithProviders(
      <FilterBar config={FULL_CONFIG} filters={{}} onFilterChange={onChange} />,
    );
    const input = screen.getByLabelText("Search");
    expect(input).toHaveValue("");
    await user.type(input, "ab");
    expect(onChange).toHaveBeenCalledWith("search", expect.any(String));
    rerender(
      <FilterBar config={FULL_CONFIG} filters={{ search: "ab" }} onFilterChange={onChange} />,
    );
    fireEvent.change(screen.getByLabelText("Search"), { target: { value: "" } });
    expect(onChange).toHaveBeenCalledWith("search", null);
  });

  it("falls back to generated search placeholder when none is given", () => {
    const config: FilterConfig[] = [{ id: "q", label: "Keyword", type: "text" }];
    renderWithProviders(<FilterBar config={config} filters={{}} onFilterChange={vi.fn()} />);
    expect(screen.getByPlaceholderText("Filter berdasarkan keyword…")).toBeInTheDocument();
  });

  it("expands and collapses the filter panel", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Host />);
    const toggle = screen.getByRole("button", { name: /^Filter/ });
    expect(toggle).toHaveAttribute("aria-expanded", "false");
    await user.click(toggle);
    expect(toggle).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByLabelText("Status")).toBeInTheDocument();
    expect(screen.getByLabelText("Nickname")).toBeInTheDocument();
    await user.click(toggle);
    await waitFor(() => expect(screen.queryByLabelText("Status")).not.toBeInTheDocument());
  });

  it("changes a select filter and selects the empty option", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <FilterBar
        config={FULL_CONFIG}
        filters={{ status: "active" }}
        onFilterChange={onChange}
        onReset={vi.fn()}
      />,
    );
    await user.click(screen.getByRole("button", { name: /^Filter/ }));
    const select = screen.getByLabelText("Status");
    expect(select).toHaveValue("active");
    await user.selectOptions(select, "");
    expect(onChange).toHaveBeenCalledWith("status", null);
    await user.selectOptions(select, "archived");
    expect(onChange).toHaveBeenCalledWith("status", "archived");
  });

  it("renders the date-range picker inside the expanded panel", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Host />);
    await user.click(screen.getByRole("button", { name: /^Filter/ }));
    expect(screen.getByText("Period")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Rentang tanggal" })).toBeInTheDocument();
  });

  it("shows select display label and falls back for unknown values", () => {
    const { rerender } = renderWithProviders(
      <FilterBar config={FULL_CONFIG} filters={{ status: "active" }} onFilterChange={vi.fn()} />,
    );
    expect(screen.getByText("Active")).toBeInTheDocument();
    rerender(
      <FilterBar config={FULL_CONFIG} filters={{ status: "custom" }} onFilterChange={vi.fn()} />,
    );
    expect(screen.getByText("custom")).toBeInTheDocument();
  });

  it("shows date-range badges for from-only and from-to ranges", () => {
    const from = new Date(2026, 0, 5);
    const to = new Date(2026, 0, 9);
    const { rerender } = renderWithProviders(
      <FilterBar config={FULL_CONFIG} filters={{ period: { from } }} onFilterChange={vi.fn()} />,
    );
    expect(screen.getByRole("group", { name: "Filter aktif" })).toBeInTheDocument();
    rerender(
      <FilterBar
        config={FULL_CONFIG}
        filters={{ period: { from, to } }}
        onFilterChange={vi.fn()}
      />,
    );
    expect(
      within(screen.getByRole("group", { name: "Filter aktif" })).getByText(/–/),
    ).toBeInTheDocument();
  });

  it("treats date ranges without from as inactive", () => {
    renderWithProviders(
      <FilterBar
        config={FULL_CONFIG}
        filters={{ period: { from: undefined, to: new Date(2026, 0, 9) } }}
        onFilterChange={vi.fn()}
      />,
    );
    expect(screen.queryByRole("group", { name: "Filter aktif" })).not.toBeInTheDocument();
  });

  it("removes a single filter through the badge button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <FilterBar config={FULL_CONFIG} filters={{ status: "active" }} onFilterChange={onChange} />,
    );
    await user.click(screen.getByRole("button", { name: "Hapus filter Status" }));
    expect(onChange).toHaveBeenCalledWith("status", null);
  });

  it("clears all filters from the panel and the active group", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    renderWithProviders(<Host initial={{ status: "active" }} onReset={onReset} />);
    await user.click(screen.getByRole("button", { name: /^Filter/ }));
    const clearButtons = screen.getAllByRole("button", { name: "Hapus semua" });
    expect(clearButtons.length).toBe(2);
    const firstClear = clearButtons[0];
    if (firstClear === undefined) throw new Error("expected a clear button");
    await user.click(firstClear);
    expect(onReset).toHaveBeenCalled();
    await user.click(
      within(screen.getByRole("group", { name: "Filter aktif" })).getByRole("button", {
        name: "Hapus semua",
      }),
    );
    expect(onReset).toHaveBeenCalledTimes(2);
  });

  it("renders active count badge on the toggle", () => {
    renderWithProviders(
      <FilterBar
        config={FULL_CONFIG}
        filters={{ status: "active", search: "x" }}
        onFilterChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "Filter, 2 aktif" })).toBeInTheDocument();
  });

  it("shows no-active-filters hint when reset exists without search", () => {
    const config: FilterConfig[] = [
      {
        id: "status",
        label: "Status",
        type: "select",
        options: [{ value: "active", label: "Active" }],
      },
    ];
    renderWithProviders(
      <FilterBar config={config} filters={{}} onFilterChange={vi.fn()} onReset={vi.fn()} />,
    );
    expect(screen.getByText("Tidak ada filter aktif")).toBeInTheDocument();
  });

  it("renders without filters button when config has search only", () => {
    const config: FilterConfig[] = [{ id: "q", label: "Keyword", type: "text" }];
    renderWithProviders(<FilterBar config={config} filters={{}} onFilterChange={vi.fn()} />);
    expect(screen.queryByRole("button", { name: /^Filter/ })).not.toBeInTheDocument();
  });

  it("renders without search input and without clear buttons when onReset is missing", () => {
    const config: FilterConfig[] = [
      {
        id: "status",
        label: "Status",
        type: "select",
        options: [{ value: "active", label: "Active" }],
      },
    ];
    renderWithProviders(
      <FilterBar config={config} filters={{ status: "active" }} onFilterChange={vi.fn()} />,
    );
    expect(screen.queryByLabelText("Search")).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Hapus semua" })).not.toBeInTheDocument();
  });

  it("renders custom actions", () => {
    renderWithProviders(<Host actions={<button type="button">Export</button>} />);
    expect(screen.getByRole("button", { name: "Export" })).toBeInTheDocument();
  });

  it("renders empty filters without active group", () => {
    renderWithProviders(<FilterBar config={[]} filters={{}} onFilterChange={vi.fn()} />);
    expect(screen.queryByRole("group", { name: "Filter aktif" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /^Filter/ })).not.toBeInTheDocument();
  });

  it("edits a secondary text filter inside the panel", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <FilterBar config={FULL_CONFIG} filters={{ nickname: "ada" }} onFilterChange={onChange} />,
    );
    await user.click(screen.getByRole("button", { name: /^Filter/ }));
    const input = screen.getByLabelText("Nickname");
    expect(input).toHaveValue("ada");
    await user.clear(input);
    expect(onChange).toHaveBeenCalledWith("nickname", null);
  });
});

describe("useFilters branches", () => {
  it("starts empty and counts empty string plus null as inactive", async () => {
    const { act } = await import("@testing-library/react");
    const { result } = renderHookWithProviders(() =>
      useFilters([
        { id: "a", label: "A", type: "text" },
        { id: "b", label: "B", type: "text" },
      ]),
    );
    expect(result.current.activeCount).toBe(0);
    act(() => result.current.setFilter("a", ""));
    expect(result.current.activeCount).toBe(0);
    act(() => result.current.setFilter("a", "x"));
    expect(result.current.activeCount).toBe(1);
  });

  it("sets and resets filters including non-string values", async () => {
    const { renderHook } = await import("@testing-library/react");
    void renderHook;
    const { result } = renderHookWithProviders(() =>
      useFilters(
        [
          { id: "n", label: "N", type: "text" },
          { id: "b", label: "B", type: "text" },
        ],
        { n: 5 as unknown as FilterValue, b: true as unknown as FilterValue },
      ),
    );
    expect(result.current.activeCount).toBe(2);
    const { act } = await import("@testing-library/react");
    act(() => result.current.setFilter("n", 7 as unknown as FilterValue));
    expect(result.current.filters.n).toBe(7);
    act(() => result.current.resetFilters());
    expect(result.current.activeCount).toBe(0);
    expect(result.current.filters.n).toBeNull();
  });
});
