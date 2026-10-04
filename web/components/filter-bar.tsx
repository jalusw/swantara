"use client";

import { CalendarDaysIcon, ChevronDownIcon, FilterIcon, RotateCcwIcon, XIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { useId, useMemo, useState } from "react";
import { cn } from "@/lib/utils";
import { formatDate } from "@/lib/utils/formatters";
import { Badge } from "./badge";
import { Button } from "./button";
import { DateRangePicker } from "./date-picker";
import { Input } from "./input";

export type DateRangeValue = { from?: Date; to?: Date } | null;

export type FilterValue = string | number | boolean | DateRangeValue;

export type FilterConfig =
  | {
      id: string;
      label: string;
      type: "text";
      placeholder?: string;
    }
  | {
      id: string;
      label: string;
      type: "select";
      options: { value: string; label: string }[];
      placeholder?: string;
    }
  | {
      id: string;
      label: string;
      type: "date-range";
      placeholder?: string;
    };

export type FiltersMap = Record<string, FilterValue>;

function emptyValueFor(_config: FilterConfig): FilterValue {
  return null;
}

function isFilterActive(value: FilterValue): boolean {
  if (value == null || value === "") {
    return false;
  }
  if (typeof value === "object" && "from" in value) {
    return (value as DateRangeValue)?.from != null;
  }
  return true;
}

function toRangeValue(
  value: FilterValue | undefined,
): { from: Date | undefined; to?: Date | undefined } | undefined {
  if (value && typeof value === "object" && "from" in value) {
    const from = (value as DateRangeValue)?.from as Date | undefined;
    const to = (value as { to?: Date | undefined })?.to as Date | undefined;
    return from ? { from, to } : undefined;
  }
  return undefined;
}

function getFilterDisplayValue(filter: FilterConfig, value: FilterValue): string {
  if (filter.type === "select" && typeof value === "string") {
    return filter.options.find((o) => o.value === value)?.label ?? value;
  }
  if (filter.type === "date-range" && value && typeof value === "object" && "from" in value) {
    const range = value as NonNullable<DateRangeValue>;
    const from = range.from ? formatDate(range.from) : "";
    const to = range.to ? formatDate(range.to) : "";
    return to ? `${from} – ${to}` : from;
  }
  return String(value);
}

export type FilterBarProps = {
  config: readonly FilterConfig[];
  filters: FiltersMap;
  onFilterChange: (id: string, value: FilterValue) => void;
  onReset?: () => void;
  className?: string;
  actions?: React.ReactNode;
};

export function FilterBar({
  config,
  filters,
  onFilterChange,
  onReset,
  className,
  actions,
}: FilterBarProps) {
  const tFilters = useTranslations("Filters" as unknown as "Common");
  const tx = tFilters as unknown as (
    key: string,
    values?: Record<string, string | number>,
  ) => string;
  const activeFilters = useMemo(() => {
    return config.filter((filter) => isFilterActive(filters[filter.id] as FilterValue));
  }, [config, filters]);
  const activeCount = activeFilters.length;
  const searchFilter = useMemo(() => config.find((f) => f.type === "text"), [config]);
  const otherFilters = useMemo(
    () => config.filter((f) => f !== searchFilter),
    [config, searchFilter],
  );
  const searchValue = searchFilter ? (filters[searchFilter.id] as FilterValue) : undefined;
  const [expanded, setExpanded] = useState(false);
  const panelId = useId();

  return (
    <div data-slot="filter-bar" className={cn("flex flex-1 flex-col gap-2 min-w-0", className)}>
      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-2">
        {searchFilter ? (
          <Input
            id="filter-search"
            type="search"
            value={typeof searchValue === "string" ? searchValue : ""}
            onChange={(event) => onFilterChange(searchFilter.id, event.target.value || null)}
            placeholder={
              searchFilter.placeholder ??
              tx("filterBy", { label: searchFilter.label.toLowerCase() })
            }
            aria-label={searchFilter.label}
            className="min-h-11 min-w-40 flex-1 px-3 py-0"
          />
        ) : null}
        {otherFilters.length > 0 ? (
          <Button
            variant="outline"
            size="sm"
            aria-expanded={expanded}
            aria-controls={panelId}
            aria-label={
              activeCount > 0 ? tx("filtersActive", { count: activeCount }) : tx("filters")
            }
            onClick={() => setExpanded((prev) => !prev)}
          >
            <FilterIcon aria-hidden />
            {tx("filters")}
            {activeCount > 0 ? (
              <Badge variant="secondary" className="ml-1 px-1.5 py-0 text-xs">
                {activeCount}
              </Badge>
            ) : null}
            <ChevronDownIcon
              className={cn("ml-1 size-3.5 transition-transform", expanded && "rotate-180")}
              aria-hidden
            />
          </Button>
        ) : null}

        {onReset && activeCount === 0 && otherFilters.length > 0 && !searchFilter ? (
          <span className="text-xs text-muted-foreground">{tx("noActiveFilters")}</span>
        ) : null}
        {actions ? (
          <div className="ml-auto flex shrink-0 flex-wrap items-center gap-2">{actions}</div>
        ) : null}
      </div>

      {otherFilters.length > 0 && expanded ? (
        <div
          id={panelId}
          data-slot="filter-panel"
          className="rounded-lg border border-border bg-muted/30 p-3"
        >
          <div className="flex items-center justify-between gap-2">
            <p className="text-sm">{tx("filters")}</p>
            {onReset && activeCount > 0 ? (
              <Button
                variant="ghost"
                size="sm"
                onClick={onReset}
                className="min-h-11 h-11 gap-1 px-2.5 text-xs"
              >
                <RotateCcwIcon className="size-3.5" aria-hidden />
                {tx("clearAll")}
              </Button>
            ) : null}
          </div>
          <div className="mt-3 grid gap-3 sm:grid-cols-2">
            {otherFilters.map((filter) => {
              const value = filters[filter.id] as FilterValue;
              if (filter.type === "text") {
                return (
                  <div key={filter.id} className="flex flex-col gap-1.5">
                    <label
                      htmlFor={`filter-${filter.id}`}
                      className="text-xs text-muted-foreground"
                    >
                      {filter.label}
                    </label>
                    <Input
                      id={`filter-${filter.id}`}
                      type="search"
                      value={typeof value === "string" ? value : ""}
                      onChange={(event) => onFilterChange(filter.id, event.target.value || null)}
                      placeholder={
                        filter.placeholder ?? tx("filterBy", { label: filter.label.toLowerCase() })
                      }
                      aria-label={filter.label}
                      className="min-h-11"
                    />
                  </div>
                );
              }
              if (filter.type === "select") {
                return (
                  <div key={filter.id} className="flex flex-col gap-1.5">
                    <label
                      htmlFor={`filter-${filter.id}`}
                      className="text-xs text-muted-foreground"
                    >
                      {filter.label}
                    </label>
                    <select
                      id={`filter-${filter.id}`}
                      aria-label={filter.label}
                      value={typeof value === "string" ? value : ""}
                      onChange={(event) => onFilterChange(filter.id, event.target.value || null)}
                      className="min-h-11 rounded-md border border-input bg-transparent px-2.5 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring"
                    >
                      <option value="">
                        {filter.placeholder ?? tx("all", { label: filter.label.toLowerCase() })}
                      </option>
                      {filter.options.map((option) => (
                        <option key={option.value} value={option.value}>
                          {option.label}
                        </option>
                      ))}
                    </select>
                  </div>
                );
              }
              return (
                <div key={filter.id} className="flex flex-col gap-1.5">
                  <span className="text-xs text-muted-foreground">{filter.label}</span>
                  <div className="flex items-center gap-1.5">
                    <CalendarDaysIcon className="size-4 text-muted-foreground" aria-hidden />
                    <DateRangePicker
                      selected={toRangeValue(value)}
                      onSelect={(range) => onFilterChange(filter.id, range ?? null)}
                      placeholder={filter.placeholder ?? tx("dateRange")}
                      className="w-full"
                    />
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      ) : null}

      {activeCount > 0 ? (
        // biome-ignore lint/a11y/useSemanticElements: active filters group uses div for layout
        <div
          role="group"
          aria-label={tx("activeFilters")}
          className="flex flex-wrap items-center gap-1.5"
        >
          {activeFilters.map((filter) => {
            const value = filters[filter.id] as FilterValue;
            return (
              <Badge key={filter.id} variant="secondary" className="gap-1 pl-2 pr-1 py-1 text-xs">
                <span className="">{filter.label}:</span>
                <span className="max-w-24 truncate">
                  {getFilterDisplayValue(filter, value as FilterValue)}
                </span>
                <button
                  type="button"
                  aria-label={tx("removeFilter", { label: filter.label })}
                  onClick={() => onFilterChange(filter.id, null)}
                  className="relative ml-1 grid size-5 place-items-center rounded-full hover:bg-muted-foreground/10 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring after:absolute after:-inset-3 after:content-['']"
                >
                  <XIcon className="size-3" aria-hidden />
                </button>
              </Badge>
            );
          })}
          {onReset ? (
            <Button
              variant="ghost"
              size="sm"
              onClick={onReset}
              className="min-h-11 h-11 gap-1 px-2.5 text-xs"
            >
              <RotateCcwIcon className="size-3.5" aria-hidden />
              {tx("clearAll")}
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

export function useFilters(config: readonly FilterConfig[], initial: FiltersMap = {}) {
  const empty = useMemo(() => {
    const base: FiltersMap = {};
    for (const filter of config) {
      base[filter.id] = emptyValueFor(filter);
    }
    return base;
  }, [config]);

  const [filters, setFilters] = useState<FiltersMap>({ ...empty, ...initial });

  function setFilter(id: string, value: FilterValue) {
    setFilters((previous) => ({ ...previous, [id]: value }));
  }

  function resetFilters() {
    setFilters({ ...empty });
  }

  const activeCount = useMemo(() => {
    return Object.values(filters).filter((value) => isFilterActive(value as FilterValue)).length;
  }, [filters]);

  return { filters, setFilter, resetFilters, activeCount };
}
