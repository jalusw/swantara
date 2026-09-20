"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import type { ReactNode } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";

export type EntitySectionProps<TData> = {
  orgId: string;
  isLoading: boolean;
  error: Error | null;
  onRetry: () => void;
  columns: ColumnDef<TData, unknown>[];
  data: TData[];
  getRowId: (row: TData, index: number) => string;
  searchKeys: string[];
  searchPlaceholder: string;
  statusKey?: keyof TData;
  statusOptions?: Array<{ value: string; label: string }>;
  filterLabel?: string;
  allLabel?: string;
  ariaLabel: string;
  emptyTitle?: string;
  dialogOpen: boolean;
  onDialogOpenChange: (open: boolean) => void;
  dialog: ReactNode;
  onAddClick: () => void;
  addLabel: string;
  actions?: ReactNode;
};

export function EntitySection<TData>({
  isLoading,
  error,
  onRetry,
  columns,
  data,
  getRowId,
  searchKeys,
  searchPlaceholder,
  statusKey,
  statusOptions,
  filterLabel,
  allLabel,
  ariaLabel,
  emptyTitle,
  dialog,
  onAddClick,
  addLabel,
  actions,
}: EntitySectionProps<TData>) {
  const status = isLoading
    ? { type: "loading" as const }
    : error
      ? { type: "error" as const, message: error.message, onRetry }
      : undefined;

  return (
    <div data-slot="entity-section" className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={data}
        getRowId={getRowId}
        searchKeys={searchKeys}
        searchPlaceholder={searchPlaceholder}
        statusKey={statusKey}
        statusOptions={statusOptions}
        filterLabel={filterLabel}
        allLabel={allLabel}
        ariaLabel={ariaLabel}
        emptyTitle={emptyTitle}
        status={status}
        actions={
          <>
            <Button onClick={onAddClick}>
              <Plus className="size-4" aria-hidden />
              <span>{addLabel}</span>
            </Button>
            {actions}
          </>
        }
      />
      {dialog}
    </div>
  );
}
