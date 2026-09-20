"use client";

import { XIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { ConfirmDialog } from "./confirm-dialog";

export type BulkAction = {
  label: string;
  variant?: "outline" | "destructive" | "default";
  onRun: () => void | Promise<void>;
  confirm?: {
    title?: string;
    description?: string;
    confirmLabel?: string;
  };
};

export type BulkActionsBarProps = {
  selected: number;
  actions?: BulkAction[];
  onClear?: () => void;
  className?: string;
};

export function BulkActionsBar({
  selected,
  actions = [],
  onClear,
  className,
}: BulkActionsBarProps) {
  if (selected === 0) {
    return null;
  }

  return (
    // biome-ignore lint/a11y/useSemanticElements: grouped action controls; grouping semantics come from role on a generic container
    <div
      data-slot="bulk-actions-bar"
      role="group"
      aria-label={"Bulk actions"}
      className={cn(
        "flex flex-wrap items-center gap-2 rounded-lg border border-border bg-muted/40 px-3 py-2",
        className,
      )}
    >
      <span className="text-sm tabular-nums" aria-live="polite">
        {selected} selected
      </span>
      <span className="h-4 w-px bg-border" aria-hidden />
      <div className="flex flex-wrap items-center gap-2">
        {actions.map((action) => {
          const button = (
            <Button
              variant={action.variant ?? "outline"}
              size="sm"
              onClick={() => {
                void action.onRun();
              }}
            >
              {action.label}
            </Button>
          );
          if (action.confirm) {
            return (
              <ConfirmDialog
                key={action.label}
                title={action.confirm.title ?? `Run "${action.label}"?`}
                description={action.confirm.description ?? "This will affect all selected rows."}
                confirmLabel={action.confirm.confirmLabel ?? "Confirm"}
                trigger={button}
                onConfirm={() => {
                  void action.onRun();
                }}
              />
            );
          }
          return <span key={action.label}>{button}</span>;
        })}
      </div>
      {onClear ? (
        <Button variant="ghost" size="sm" onClick={onClear} className="ml-auto">
          <XIcon aria-hidden />
          Clear
        </Button>
      ) : null}
    </div>
  );
}
