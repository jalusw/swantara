"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";
import { CustomerFormDialog } from "./customer-form-dialog";

export type CustomerRow = {
  id: string;
  name: string;
  email: string;
  status: "active" | "dormant";
  since: string;
};

function toCustomerRow(contact: Contact): CustomerRow {
  return {
    id: String(contact.id),
    name: contact.displayName ?? contact.name,
    email: contact.email ?? "",
    status: contact.active ? "active" : "dormant",
    since: contact.createdAt ? new Date(contact.createdAt).toISOString().split("T")[0]! : "",
  };
}

export function CustomersTable({ orgId }: { orgId: string }) {
  const t = useTranslations("Sales");
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const customers = (query.data?.contacts ?? []).map(toCustomerRow);

  function handleCreate() {
    setDialogOpen(false);
    void query.refetch();
  }

  function statusLabel(status: CustomerRow["status"]): string {
    try {
      return (t as unknown as (k: string) => string)(`customerStatus.${status}`);
    } catch {
      return humanizeKey(String(status));
    }
  }

  const columns: ColumnDef<CustomerRow>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => (
        <div className="flex min-w-0 flex-col">
          <span className="truncate font-medium">{row.original.name}</span>
          {row.original.email ? (
            <span className="truncate text-xs text-muted-foreground">{row.original.email}</span>
          ) : null}
        </div>
      ),
    },
    {
      accessorKey: "status",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const toneMap: Record<CustomerRow["status"], "success" | "neutral"> = {
          active: "success",
          dormant: "neutral",
        };
        return (
          <StateBadge
            tone={toneMap[row.original.status]}
            label={statusLabel(row.original.status)}
          />
        );
      },
    },
    {
      accessorKey: "since",
      header: t("tableCustomerSince"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.since}</span>,
    },
  ];

  return (
    <>
      <InteractiveEntityTable
        columns={columns}
        data={customers}
        getRowId={(row) => row.id}
        searchKeys={["name", "email"]}
        statusKey="status"
        statusOptions={[
          { value: "active", label: t("statusActive") },
          { value: "dormant", label: t("statusDormant") },
        ]}
        searchPlaceholder={t("searchCustomersPlaceholder")}
        filterLabel={t("filterByStatus")}
        allLabel={t("allStatuses")}
        ariaLabel={t("allCustomers")}
        emptyTitle={t("emptyCustomers")}
        emptyDescription={t("emptyCustomersDescription")}
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
          <Button onClick={() => setDialogOpen(true)}>
            <Plus aria-hidden />
            <span>{t("addCustomer")}</span>
          </Button>
        }
      />
      <CustomerFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        orgId={orgId}
        onSave={handleCreate}
      />
    </>
  );
}
