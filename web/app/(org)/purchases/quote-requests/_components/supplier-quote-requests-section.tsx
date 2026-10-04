"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Purchases");
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

  function stateLabel(state: SupplierQuoteRequest["state"]): string {
    try {
      return (t as unknown as (k: string) => string)(`quoteRequestState.${state}`);
    } catch {
      return humanizeKey(String(state));
    }
  }

  const columns: ColumnDef<SupplierQuoteRequest>[] = [
    {
      accessorKey: "name",
      header: t("tableQuoteRequest"),
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
      header: t("tableSupplier"),
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
      header: t("tableStatus"),
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
          {stateLabel(row.original.state)}
        </Badge>
      ),
    },
    {
      accessorKey: "orderDate",
      header: t("tableOrderDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.orderDate ? formatDate(row.original.orderDate) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "quoteDeadline",
      header: t("tableQuoteDeadline"),
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
          editLabel={t("editQuoteRequest")}
          deleteLabel={t("deleteQuoteRequest")}
          confirmTitle={t("deleteQuoteRequestTitle")}
          confirmDescription={t("deleteQuoteRequestDescription")}
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
          { value: "draft", label: t("quoteRequestStateDraft") },
          { value: "sent", label: t("quoteRequestStateSent") },
          { value: "done", label: t("quoteRequestStateDone") },
          { value: "cancelled", label: t("quoteRequestStateCancelled") },
        ]}
        searchPlaceholder={t("searchQuoteRequestsPlaceholder")}
        filterLabel={t("tableStatus")}
        allLabel={t("allQuoteRequests")}
        ariaLabel={t("allQuoteRequests")}
        emptyTitle={t("emptyQuoteRequests")}
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
            <span>{t("newQuoteRequest")}</span>
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
