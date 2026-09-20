"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { ServerEntityTable } from "@/app/(org)/_components/server-entity-section";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import type { Contact, ListQuery } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { activeColumn, nameColumn } from "@/lib/utils/table-columns";
import { ContactFormDialog } from "./contact-form-dialog";

export function ContactsServerSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

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
      <ServerEntityTable
        resource="contacts"
        fetcher={(organizationId: number, params: ListQuery) =>
          getSwantaraService().contacts.list(organizationId, params)
        }
        selectData={(response: { contacts: Contact[] }) => response.contacts}
        columns={columns}
        getRowId={(row: Contact) => String(row.id)}
        searchKeys={["name", "email"]}
        statusKey="active"
        statusOptions={[
          { value: "true", label: "Active" },
          { value: "false", label: "Inactive" },
        ]}
        searchPlaceholder={"Search contacts…"}
        filterLabel={"Filter by status"}
        allLabel={"All statuses"}
        ariaLabel={"All contacts"}
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
        onSave={() => setDialogOpen(false)}
      />
    </div>
  );
}
