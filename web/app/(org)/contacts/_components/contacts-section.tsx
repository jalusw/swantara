"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { activeColumn, nameColumn } from "@/lib/utils/table-columns";
import { ContactFormDialog } from "./contact-form-dialog";

export function ContactsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Contacts");
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contacts = query.data?.contacts ?? [];

  function handleCreate() {
    setDialogOpen(false);
    void query.refetch();
  }

  const columns: ColumnDef<Contact>[] = [
    nameColumn<Contact>({
      basePath: "contacts",
      header: t("fieldName"),
      displayNameAccessor: (row) => row.displayName,
    }),
    {
      accessorKey: "isOrganization",
      header: t("fieldType"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {row.original.isOrganization ? t("typeOrganization") : t("typeIndividual")}
        </Badge>
      ),
    },
    {
      accessorKey: "email",
      header: t("fieldEmail"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.email || "—"}</span>,
    },
    activeColumn<Contact>({
      header: t("tableStatus"),
      activeLabel: t("statusActive"),
      inactiveLabel: t("statusInactive"),
    }),
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={contacts}
        getRowId={(row) => String(row.id)}
        searchKeys={["name", "displayName", "email"]}
        statusKey="active"
        statusOptions={[
          { value: "true", label: t("statusActive") },
          { value: "false", label: t("statusInactive") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("filterByStatus")}
        allLabel={t("allStatuses")}
        ariaLabel={t("allContacts")}
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
            <span>{t("addContact")}</span>
          </Button>
        }
      />
      <ContactFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        orgId={orgId}
        onSave={handleCreate}
      />
    </div>
  );
}
