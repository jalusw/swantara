"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { QualityCheck } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import {
  canRecordResult,
  checkResultLabel,
  checkResultTone,
} from "../../_components/quality-utils";
import { RecordResultDialog } from "./record-result-dialog";

export function QualityChecksSection({ orgId }: { orgId: string }) {
  const [resultDialogCheck, setResultDialogCheck] = useState<QualityCheck | null>(null);

  const query = useOrgListQuery<{ checks: QualityCheck[] }, Record<string, never>>(
    "qualityChecks",
    (organizationId) => getSwantaraService().qualityChecks.list(organizationId),
  );

  const checks = query.data?.checks ?? [];
  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  const columns: ColumnDef<QualityCheck>[] = [
    {
      accessorKey: "id",
      header: "Check",
      cell: ({ row }) => `#${row.original.id}`,
    },
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => (row.original.itemId ? `Item #${row.original.itemId}` : "—"),
    },
    {
      accessorKey: "shipmentId",
      header: "Shipment",
      cell: ({ row }) => (row.original.shipmentId ? `#${row.original.shipmentId}` : "—"),
    },
    {
      accessorKey: "productionOrderId",
      header: "MO",
      cell: ({ row }) =>
        row.original.productionOrderId ? `#${row.original.productionOrderId}` : "—",
    },
    {
      accessorKey: "result",
      header: "Result",
      cell: ({ row }) => (
        <StateBadge
          tone={checkResultTone(row.original.result)}
          label={checkResultLabel(row.original.result)}
        />
      ),
    },
    {
      accessorKey: "checkedAt",
      header: "Checked at",
      cell: ({ row }) => (row.original.checkedAt ? formatDate(row.original.checkedAt) : "—"),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        canRecordResult(row.original) ? (
          <Button size="sm" variant="ghost" onClick={() => setResultDialogCheck(row.original)}>
            {"Record result"}
          </Button>
        ) : null,
    },
  ];

  return (
    <>
      <InteractiveEntityTable
        columns={columns}
        data={checks}
        getRowId={(row) => String(row.id)}
        searchKeys={["result"]}
        statusKey="result"
        statusOptions={[
          { value: "pending", label: checkResultLabel("pending") },
          { value: "pass", label: checkResultLabel("pass") },
          { value: "fail", label: checkResultLabel("fail") },
        ]}
        searchPlaceholder={"Search quality checks…"}
        filterLabel={"Result"}
        allLabel={"All checks"}
        ariaLabel={"Quality Checks"}
        emptyTitle={"No quality checks"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
      />
      {resultDialogCheck ? (
        <RecordResultDialog
          open={Boolean(resultDialogCheck)}
          onOpenChange={(open) => {
            if (!open) setResultDialogCheck(null);
          }}
          orgId={orgId}
          check={resultDialogCheck}
          onSave={() => {
            setResultDialogCheck(null);
            void query.refetch();
          }}
        />
      ) : null}
    </>
  );
}
