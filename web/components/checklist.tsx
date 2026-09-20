"use client";

import { cn } from "@/lib/utils";
import { Checkbox } from "./checkbox";

export type ChecklistItem = {
  id: string;
  label: string;
  description?: string;
};

export type ChecklistProps = {
  items: ChecklistItem[];
  values?: Record<string, boolean>;
  onCheckedChange?: (id: string, checked: boolean) => void;
  readOnly?: boolean;
  className?: string;
};

export function Checklist({
  items,
  values = {},
  onCheckedChange,
  readOnly = false,
  className,
}: ChecklistProps) {
  const editable = !readOnly && Boolean(onCheckedChange);
  const doneCount = items.filter((item) => values[item.id]).length;
  const percent = items.length === 0 ? 0 : Math.round((doneCount / items.length) * 100);

  return (
    <div
      data-slot="checklist"
      className={cn("rounded-md border border-border bg-background p-4", className)}
    >
      <div className="mb-3 flex items-center justify-between gap-3">
        <span className="text-sm">{"Progress"}</span>
        <span className="text-xs text-muted-foreground tabular-nums">
          {doneCount} of {items.length}
        </span>
      </div>
      <div
        role="progressbar"
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={percent}
        aria-label={"Checklist progress"}
        className="h-1.5 w-full overflow-hidden rounded-full bg-muted"
      >
        <div
          className="h-full rounded-full bg-primary transition-all"
          style={{ width: `${percent}%` }}
        />
      </div>
      <ul className="mt-4 space-y-1">
        {items.map((item) => {
          const checked = Boolean(values[item.id]);
          return (
            <li key={item.id}>
              <div
                data-state={checked ? "checked" : "unchecked"}
                className={cn(
                  "flex items-start gap-3 rounded-md p-2 transition-colors",
                  checked && "bg-muted/60",
                )}
              >
                <Checkbox
                  checked={checked}
                  disabled={!editable}
                  aria-label={item.label}
                  onCheckedChange={(next) => onCheckedChange?.(item.id, Boolean(next))}
                  className="mt-0.5"
                />
                <span className="min-w-0 space-y-0.5">
                  <span
                    className={cn("block text-sm", checked && "text-muted-foreground line-through")}
                  >
                    {item.label}
                  </span>
                  {item.description ? (
                    <span className="block text-xs text-muted-foreground">{item.description}</span>
                  ) : null}
                </span>
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
