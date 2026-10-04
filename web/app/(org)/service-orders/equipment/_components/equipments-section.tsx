"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Equipment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { EquipmentFormDialog } from "./equipment-form-dialog";

export function EquipmentsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Service");
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ equipments: Equipment[] }, Record<string, never>>(
    "equipments",
    (organizationId) => getSwantaraService().equipments.list(organizationId),
  );

  const equipments = query.data?.equipments ?? [];

  function handleSave(_id: string) {
    setDialogOpen(false);
    void query.refetch();
  }

  const columns: ColumnDef<Equipment>[] = [
    {
      accessorKey: "name",
      header: t("colName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "category",
      header: t("category"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.category || "—"}</span>
      ),
    },
    {
      accessorKey: "location",
      header: t("location"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.location || "—"}</span>
      ),
    },
    {
      accessorKey: "installDate",
      header: t("installDate"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.installDate ? formatDate(String(row.original.installDate)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "warrantyEnd",
      header: t("warrantyEnd"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.warrantyEnd ? formatDate(String(row.original.warrantyEnd)) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("view")}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/service-orders/equipment/${row.original.id}`)}
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
            <CardTitle>{t("totalEquipment")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{equipments.length}</p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={equipments}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={t("equipmentSearchPlaceholder")}
        ariaLabel={t("equipmentTitle")}
        emptyTitle={t("equipmentEmpty")}
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
            <span>{t("newEquipment")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <EquipmentFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
