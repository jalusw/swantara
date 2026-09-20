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
import type { Contact, SupplierQuoteRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { SupplierQuoteRequestFormDialog } from "./supplier-quote-request-form-dialog";

export function SupplierQuoteRequestsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const quoteRequestsQuery = useOrgListQuery<
    { quoteRequests: SupplierQuoteRequest[] },
    Record<string, never>
  >("supplierQuoteRequests", (organizationId) =>
    getSwantaraService().supplierQuoteRequests.list(organizationId),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const quoteRequests = quoteRequestsQuery.data?.quoteRequests ?? [];
  const contacts = contactsQuery.data?.contacts ?? [];
  const contactMap = new Map(contacts.map((p) => [p.id, p.displayName || p.name]));

  function handleSave() {
    setDialogOpen(false);
    void quoteRequestsQuery.refetch();
  }

  const columns: ColumnDef<SupplierQuoteRequest>[] = [
    {
      accessorKey: "name",
      header: "QuoteRequest",
      cell: ({ row }) => (
        <a
          href={`/purchases/quoteRequests/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `QuoteRequest-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "supplierId",
      header: "Supplier",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.supplierId
            ? (contactMap.get(row.original.supplierId) ?? `#${row.original.supplierId}`)
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
            row.original.state === "done"
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
      accessorKey: "orderDate",
      header: "Order date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.orderDate ? formatDate(row.original.orderDate) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "quoteDeadline",
      header: "Quote deadline",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.quoteDeadline ? formatDate(row.original.quoteDeadline) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit QuoteRequest"}
          deleteLabel={"Delete QuoteRequest"}
          confirmTitle={"Delete this QuoteRequest?"}
          confirmDescription={"The QuoteRequest will be removed. This cannot be undone."}
          onEdit={() => router.push(`/purchases/quoteRequests/${row.original.id}`)}
        />
      ),
    },
  ];

  const isLoading = quoteRequestsQuery.isLoading || contactsQuery.isLoading;
  const error = quoteRequestsQuery.isError
    ? quoteRequestsQuery.error
    : contactsQuery.isError
      ? contactsQuery.error
      : null;

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={quoteRequests}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "sent", label: "Sent" },
          { value: "done", label: "Done" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search Quote Requests…"}
        filterLabel={"Status"}
        allLabel={"All Quote Requests"}
        ariaLabel={"All Quote Requests"}
        emptyTitle={"No purchase Quote Requests"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void quoteRequestsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New QuoteRequest"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <SupplierQuoteRequestFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
