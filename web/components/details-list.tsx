import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export type DetailsListItem = {
  id: string;
  label: ReactNode;
  value: ReactNode;
};

export type DetailsListProps = {
  items: DetailsListItem[];
  layout?: "rows" | "grid";
  columns?: 2 | 3;
  className?: string;
};

export function DetailsList({ items, layout = "rows", columns = 2, className }: DetailsListProps) {
  if (layout === "grid") {
    return (
      <dl
        data-slot="details-list"
        className={cn(
          "grid gap-x-6 gap-y-4",
          columns === 3 ? "sm:grid-cols-3" : "sm:grid-cols-2",
          className,
        )}
      >
        {items.map((item) => (
          <div key={item.id} className="space-y-1">
            <dt className="text-xs tracking-wide text-muted-foreground uppercase">{item.label}</dt>
            <dd className="text-sm">{item.value}</dd>
          </div>
        ))}
      </dl>
    );
  }

  return (
    <dl
      data-slot="details-list"
      className={cn("divide-y divide-border rounded-md border border-border", className)}
    >
      {items.map((item) => (
        <div key={item.id} className="flex items-baseline justify-between gap-4 px-4 py-3">
          <dt className="text-sm text-muted-foreground">{item.label}</dt>
          <dd className="text-right text-sm">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
