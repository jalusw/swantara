"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { MaintenancePlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { MaintenancePlanFormDialog } from "./maintenance-plan-form-dialog";

export function MaintenancePlansSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ maintenancePlans: MaintenancePlan[] }, Record<string, never>>(
    "maintenancePlans",
    (organizationId) => getSwantaraService().maintenancePlans.list(organizationId),
  );

  const plans = query.data?.maintenancePlans ?? [];

  function handleSave(_id: string) {
    setDialogOpen(false);
    void query.refetch();
  }

  const columns: ColumnDef<MaintenancePlan>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "intervalDays",
      header: "Interval (days)",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">{row.original.intervalDays} days</span>
      ),
    },
    {
      accessorKey: "nextDue",
      header: "Next due",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.nextDue ? formatDate(String(row.original.nextDue)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "active",
      header: "Active",
      cell: ({ row }) => <span className="text-sm">{row.original.active ? "Yes" : "No"}</span>,
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
        ariaLabel={"Maintenance plans"}
        emptyTitle={"No maintenance plans yet"}
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
        <MaintenancePlanFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
