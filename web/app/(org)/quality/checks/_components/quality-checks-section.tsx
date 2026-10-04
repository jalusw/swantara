"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { QualityCheck } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { canRecordResult, checkResultTone } from "../../_components/quality-utils";
import { RecordResultDialog } from "./record-result-dialog";

export function QualityChecksSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Quality");
  const resultLabel = (result: string) =>
    (t as unknown as (k: string) => string)(`result_${result}`);
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
      header: t("colCheck"),
      cell: ({ row }) => `#${row.original.id}`,
    },
    {
      accessorKey: "itemId",
      header: t("colItem"),
      cell: ({ row }) => (row.original.itemId ? `Item #${row.original.itemId}` : "—"),
    },
    {
      accessorKey: "shipmentId",
      header: t("colShipment"),
      cell: ({ row }) => (row.original.shipmentId ? `#${row.original.shipmentId}` : "—"),
    },
    {
      accessorKey: "productionOrderId",
      header: t("colProductionOrder"),
      cell: ({ row }) =>
        row.original.productionOrderId ? `#${row.original.productionOrderId}` : "—",
    },
    {
      accessorKey: "result",
      header: t("colResult"),
      cell: ({ row }) => (
        <StateBadge
          tone={checkResultTone(row.original.result)}
          label={resultLabel(row.original.result)}
        />
      ),
    },
    {
      accessorKey: "checkedAt",
      header: t("colCheckedAt"),
      cell: ({ row }) => (row.original.checkedAt ? formatDate(row.original.checkedAt) : "—"),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) =>
        canRecordResult(row.original) ? (
          <Button size="sm" variant="ghost" onClick={() => setResultDialogCheck(row.original)}>
            {t("recordResult")}
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
          { value: "pending", label: resultLabel("pending") },
          { value: "pass", label: resultLabel("pass") },
          { value: "fail", label: resultLabel("fail") },
        ]}
        searchPlaceholder={t("checksSearchPlaceholder")}
        filterLabel={t("colResult")}
        allLabel={t("allChecks")}
        ariaLabel={t("checksTitle")}
        emptyTitle={t("checksEmpty")}
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
