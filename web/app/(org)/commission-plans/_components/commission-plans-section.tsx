"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CommissionPlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { CommissionPlanFormDialog } from "./commission-plan-form-dialog";

export function CommissionPlansSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ commissionPlans: CommissionPlan[] }, Record<string, never>>(
    "commissionPlans",
    (organizationId) => getSwantaraService().commissionPlans.list(organizationId),
  );

  const plans = query.data?.commissionPlans ?? [];

  function handleSave(id: string) {
    setDialogOpen(false);
    void query.refetch();
    router.push(`/commission-plans/${id}`);
  }

  const columns: ColumnDef<CommissionPlan>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/commission-plans/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "basis",
      header: "Basis",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{String(row.original.basis)}</span>
      ),
    },
    {
      accessorKey: "active",
      header: "Active",
      cell: ({ row }) => (
        <Badge
          variant="outline"
          className={
            row.original.active
              ? "border-success text-success"
              : "border-muted-foreground text-muted-foreground"
          }
        >
          {row.original.active ? "Yes" : "No"}
        </Badge>
      ),
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
          onEdit={() => router.push(`/commission-plans/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{"Total Plans"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{plans.length}</p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={plans}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search plans…"}
        ariaLabel={"Commission plans"}
        emptyTitle={"No commission plans yet"}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Create plan"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <CommissionPlanFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
