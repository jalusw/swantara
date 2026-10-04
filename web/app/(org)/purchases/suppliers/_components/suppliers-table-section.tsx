"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";

export type SupplierRow = {
  id: string;
  name: string;
  contact: string;
  category: string;
  status: "active" | "review" | "inactive";
  rating: number;
  leadTime: string;
  orders: number;
  since: string;
};

function toSupplierRow(contact: Contact): SupplierRow {
  return {
    id: String(contact.id),
    name: contact.displayName ?? contact.name,
    contact: contact.email ?? "",
    category: contact.industry ?? "general",
    status: contact.active ? "active" : "inactive",
    rating: 0,
    leadTime: "—",
    orders: 0,
    since: contact.createdAt ? new Date(contact.createdAt).toISOString().split("T")[0]! : "",
  };
}

export function SuppliersTable() {
  const t = useTranslations("Purchases");
  const query = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "suppliers",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const suppliers = (query.data?.contacts ?? []).map(toSupplierRow);

  function statusLabel(status: SupplierRow["status"]): string {
    try {
      return (t as unknown as (k: string) => string)(`supplierStatus.${status}`);
    } catch {
      return humanizeKey(String(status));
    }
  }

  const columns: ColumnDef<SupplierRow>[] = [
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => (
        <div className="flex flex-col">
          <span className="">{row.original.name}</span>
          <span className="text-xs text-muted-foreground">{row.original.contact}</span>
        </div>
      ),
    },
    {
      accessorKey: "category",
      header: t("tableCategory"),
      cell: ({ row }) => <Badge variant="secondary">{row.original.category}</Badge>,
    },
    {
      accessorKey: "rating",
      header: t("tableRating"),
      cell: ({ row }) => (
        <span className=" tabular-nums">
          {row.original.rating > 0 ? row.original.rating.toFixed(1) : "—"}
          {row.original.rating > 0 && <span className="text-muted-foreground"> / 5</span>}
        </span>
      ),
    },
    {
      accessorKey: "leadTime",
      header: t("tableLeadTime"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.leadTime}</span>,
    },
    {
      accessorKey: "orders",
      header: t("tableOrders"),
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{row.original.orders}</span>
      ),
    },
    {
      accessorKey: "status",
      header: t("tableStatus"),
      cell: ({ row }) => {
        const toneMap: Record<SupplierRow["status"], "success" | "warning" | "neutral"> = {
          active: "success",
          review: "warning",
          inactive: "neutral",
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
      header: t("tableSince"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.since}</span>,
    },
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={suppliers}
      getRowId={(row) => row.id}
      searchKeys={["name", "contact", "category"]}
      statusKey="status"
      statusOptions={[
        { value: "active", label: t("supplierStatusActive") },
        { value: "review", label: t("supplierStatusReview") },
        { value: "inactive", label: t("supplierStatusInactive") },
      ]}
      searchPlaceholder={t("searchSuppliersPlaceholder")}
      filterLabel={t("filterByStatus")}
      allLabel={t("allStatuses")}
      ariaLabel={t("allSuppliers")}
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
        <Button size="sm">
          <Plus />
          <span>{t("addSupplier")}</span>
        </Button>
      }
    />
  );
}
