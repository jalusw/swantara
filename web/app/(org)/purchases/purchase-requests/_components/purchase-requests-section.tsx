"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Department, PurchaseRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { PurchaseRequestFormDialog } from "./purchase-request-form-dialog";

export function PurchaseRequestsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const requisitionsQuery = useOrgListQuery<
    { requisitions: PurchaseRequest[] },
    Record<string, never>
  >("purchaseRequests", (organizationId) =>
    getSwantaraService().purchaseRequests.list(organizationId),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const requisitions = requisitionsQuery.data?.requisitions ?? [];
  const contacts = contactsQuery.data?.contacts ?? [];
  const departments = departmentsQuery.data?.departments ?? [];
  const contactMap = new Map(contacts.map((p) => [p.id, p.displayName || p.name]));
  const departmentMap = new Map(departments.map((d) => [d.id, d.name]));

  function handleSave() {
    setDialogOpen(false);
    void requisitionsQuery.refetch();
  }

  const columns: ColumnDef<PurchaseRequest>[] = [
    {
      accessorKey: "name",
      header: "Request",
      cell: ({ row }) => (
        <a
          href={`/purchases/requisitions/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `PR-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "requesterId",
      header: "Requester",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {contactMap.get(row.original.requesterId) ?? `#${row.original.requesterId}`}
        </span>
      ),
    },
    {
      accessorKey: "departmentId",
      header: "Department",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.departmentId
            ? (departmentMap.get(row.original.departmentId) ?? `#${row.original.departmentId}`)
            : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ row }) => (
        <Badge
          variant={
            row.original.state === "approved" || row.original.state === "done"
              ? "default"
              : row.original.state === "cancelled"
                ? "outline"
                : "secondary"
          }
        >
          {humanizeKey(String(row.original.state))}
        </Badge>
      ),
    },
    {
      accessorKey: "neededBy",
      header: "Needed by",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.neededBy ? formatDate(row.original.neededBy) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "lines",
      header: "Total qty",
      meta: { align: "right" },
      cell: ({ row }) => {
        const lines = row.original.lines ?? [];
        const total = lines.reduce((sum, l) => sum + l.qty, 0);
        return <span className="tabular-nums ">{formatNumber(total)}</span>;
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit request"}
          deleteLabel={"Delete request"}
          confirmTitle={"Delete this request?"}
          confirmDescription={"The purchase request will be removed. This cannot be undone."}
          onEdit={() => router.push(`/purchases/requisitions/${row.original.id}`)}
        />
      ),
    },
  ];

  const isLoading =
    requisitionsQuery.isLoading || contactsQuery.isLoading || departmentsQuery.isLoading;
  const error = requisitionsQuery.isError
    ? requisitionsQuery.error
    : contactsQuery.isError
      ? contactsQuery.error
      : departmentsQuery.isError
        ? departmentsQuery.error
        : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={requisitions}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "confirmed", label: "Confirmed" },
          { value: "approved", label: "Approved" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search requisitions…"}
        filterLabel={"Status"}
        allLabel={"All purchase requisitions"}
        ariaLabel={"All purchase requisitions"}
        emptyTitle={"No purchase requisitions"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void requisitionsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New request"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <PurchaseRequestFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
