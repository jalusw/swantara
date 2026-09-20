"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { FixedAsset } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { FixedAssetFormDialog } from "./fixed-asset-form-dialog";
import { fixedAssetStateTone } from "./fixed-asset-utils";

export function FixedAssetsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const assetsQuery = useOrgListQuery<{ fixedAssets: FixedAsset[] }, Record<string, never>>(
    "fixedAssets",
    (organizationId) => getSwantaraService().fixedAssets.list(organizationId),
  );

  const assets = assetsQuery.data?.fixedAssets ?? [];

  function handleSave(assetId: string) {
    setDialogOpen(false);
    void assetsQuery.refetch();
    router.push(`/fixed-assets/${assetId}`);
  }

  const columns: ColumnDef<FixedAsset>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/fixed-assets/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "purchaseValue",
      header: "Purchase value",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.purchaseValue, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "acquisitionDate",
      header: "Acquisition date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.acquisitionDate ? formatDate(String(row.original.acquisitionDate)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const tone = fixedAssetStateTone(row.original.state);
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "info"
                    ? "border-info text-info"
                    : tone === "danger"
                      ? "border-destructive text-destructive"
                      : ""
            }
          >
            {humanizeKey(String(row.original.state))}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/fixed-assets/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={assets}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "running", label: "Running" },
          { value: "sold", label: "Sold" },
          { value: "disposed", label: "Disposed" },
        ]}
        searchPlaceholder={"Search assets…"}
        filterLabel={"State"}
        allLabel={"All assets"}
        ariaLabel={"Fixed assets"}
        emptyTitle={"No fixed assets yet"}
        status={
          assetsQuery.isLoading
            ? { type: "loading" }
            : assetsQuery.isError
              ? {
                  type: "error",
                  message: assetsQuery.error.message,
                  onRetry: () => void assetsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Register asset"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <FixedAssetFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
