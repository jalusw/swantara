"use client";

import { useQuery } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, QualityPoint } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { normRangeLabel, testTypeLabel } from "../../_components/quality-utils";
import { QualityPointFormDialog } from "./quality-point-form-dialog";

export function QualityPointsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ points: QualityPoint[] }, Record<string, never>>(
    "qualityPoints",
    (organizationId, _params) => getSwantaraService().qualityPoints.list(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId, _params) => getSwantaraService().products.list(organizationId),
  );

  const unitsQuery = useQuery({
    queryKey: ["units"],
    queryFn: () => getSwantaraService().units.list(),
  });

  const points = query.data?.points ?? [];
  const productMap = new Map((productsQuery.data?.products ?? []).map((p) => [p.id, p.name]));
  const uomMap = new Map((unitsQuery.data?.units ?? []).map((u) => [u.id, u.name]));

  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  const columns: ColumnDef<QualityPoint>[] = [
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => (row.original.itemId ? productMap.get(row.original.itemId) : null) ?? "—",
    },
    {
      accessorKey: "operation",
      header: "Operation",
      cell: ({ row }) => row.original.operation ?? "—",
    },
    {
      accessorKey: "testType",
      header: "Test type",
      cell: ({ row }) => <Badge variant="secondary">{testTypeLabel(row.original.testType)}</Badge>,
    },
    {
      accessorKey: "norms",
      header: "Norms",
      cell: ({ row }) => normRangeLabel(row.original),
    },
    {
      accessorKey: "unitId",
      header: "Unit of Measure",
      cell: ({ row }) => (row.original.unitId ? uomMap.get(row.original.unitId) : null) ?? "—",
    },
    {
      id: "actions",
      header: "",
      cell: () => (
        <Button size="sm" variant="ghost">
          {"Edit"}
        </Button>
      ),
    },
  ];

  return (
    <>
      <InteractiveEntityTable
        columns={columns}
        data={points}
        getRowId={(row) => String(row.id)}
        searchKeys={["operation"]}
        searchPlaceholder={"Search quality points…"}
        ariaLabel={"Quality Points"}
        emptyTitle={"No quality points"}
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
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New quality point"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <QualityPointFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={() => {
            setDialogOpen(false);
            void query.refetch();
          }}
        />
      ) : null}
    </>
  );
}
