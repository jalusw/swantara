"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Project } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";
import { ProjectFormDialog } from "./project-form-dialog";
import { billingTypeLabel, projectStateTone } from "./project-utils";

export function ProjectsSection({ orgId }: { orgId: string }) {
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ projects: Project[] }, Record<string, never>>(
    "projects",
    (organizationId) => getSwantaraService().projects.list(organizationId),
  );

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const projects = query.data?.projects ?? [];
  const contactMap = new Map(
    (contactsQuery.data?.contacts ?? []).map((p) => [p.id, p.displayName ?? p.name]),
  );

  const isLoading = query.isLoading || contactsQuery.isLoading;
  const error = query.isError ? query.error : contactsQuery.isError ? contactsQuery.error : null;

  const columns: ColumnDef<Project>[] = [
    {
      accessorKey: "name",
      header: "Project",
      cell: ({ row }) => (
        <button
          type="button"
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring text-left"
          onClick={() => router.push(`/projects/${row.original.id}`)}
        >
          {row.original.name}
        </button>
      ),
    },
    {
      accessorKey: "contactId",
      header: "Customer",
      cell: ({ row }) => contactMap.get(row.original.contactId) ?? "—",
    },
    {
      accessorKey: "billingType",
      header: "Billing type",
      cell: ({ row }) => (
        <Badge variant="secondary">{billingTypeLabel(row.original.billingType)}</Badge>
      ),
    },
    {
      accessorKey: "state",
      header: "Status",
      cell: ({ row }) => (
        <StateBadge
          tone={projectStateTone(row.original.state)}
          label={humanizeKey(String(row.original.state))}
        />
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <Button
          size="sm"
          variant="ghost"
          onClick={() => router.push(`/projects/${row.original.id}`)}
        >
          {"Edit project"}
        </Button>
      ),
    },
  ];

  return (
    <>
      <InteractiveEntityTable
        columns={columns}
        data={projects}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "open", label: "Open" },
          { value: "closed", label: "Closed" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search projects…"}
        filterLabel={"Status"}
        allLabel={"All projects"}
        ariaLabel={"All projects"}
        emptyTitle={"No projects"}
        status={
          isLoading
            ? { type: "loading" }
            : error
              ? {
                  type: "error",
                  message: error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New project"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <ProjectFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={() => {
            setDialogOpen(false);
            void query.refetch();
          }}
        />
      ) : null}
    </>
  );
}
