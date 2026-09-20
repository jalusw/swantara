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
import type { ServiceContract } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { ServiceContractFormDialog } from "./service-contract-form-dialog";
import { serviceContractStateTone } from "./service-contract-utils";

export function ServiceContractsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ serviceContracts: ServiceContract[] }, Record<string, never>>(
    "serviceContracts",
    (organizationId) => getSwantaraService().serviceContracts.list(organizationId),
  );

  const contracts = query.data?.serviceContracts ?? [];

  function handleSave(id: string) {
    setDialogOpen(false);
    void query.refetch();
    router.push(`/service-orders/contracts/${id}`);
  }

  const columns: ColumnDef<ServiceContract>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/service-orders/contracts/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name}
        </a>
      ),
    },
    {
      accessorKey: "coverage",
      header: "Coverage",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.coverage || "—"}</span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: "Start date",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.dateStart ? formatDate(String(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: "End date",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.dateEnd ? formatDate(String(row.original.dateEnd)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => {
        const tone = serviceContractStateTone(row.original.state);
        return (
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : tone === "info"
                      ? "border-info text-info"
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
          onEdit={() => router.push(`/service-orders/contracts/${row.original.id}`)}
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
            <CardTitle>{"Total Contracts"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{contracts.length}</p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={contracts}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "active", label: "Active" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search contracts…"}
        filterLabel={"State"}
        allLabel={"All contracts"}
        ariaLabel={"Service contracts"}
        emptyTitle={"No service contracts yet"}
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
            <span>{"Create contract"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ServiceContractFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
