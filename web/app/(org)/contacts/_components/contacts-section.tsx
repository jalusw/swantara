"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
      header: "Name",
      displayNameAccessor: (row) => row.displayName,
    }),
    {
      accessorKey: "isOrganization",
      header: "Type",
      cell: ({ row }) => (
        <Badge variant="secondary">
          {row.original.isOrganization ? "Organization" : "Individual"}
        </Badge>
      ),
    },
    {
      accessorKey: "email",
      header: "Email",
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.email || "—"}</span>,
    },
    activeColumn<Contact>({
      header: "Status",
      activeLabel: "Active",
      inactiveLabel: "Inactive",
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
          { value: "true", label: "Active" },
          { value: "false", label: "Inactive" },
        ]}
        searchPlaceholder={"Search contacts…"}
        filterLabel={"Filter by status"}
        allLabel={"All statuses"}
        ariaLabel={"All contacts"}
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
            <span>{"Add contact"}</span>
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
