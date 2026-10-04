"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { StateBadge } from "@/components/state-badge";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Project } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { ProjectFormDialog } from "./project-form-dialog";
import { projectStateTone } from "./project-utils";

export function ProjectsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const dyn = (key: string) => (t as unknown as (k: string) => string)(key);
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
      header: t("colProject"),
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
      header: t("colCustomer"),
      cell: ({ row }) => contactMap.get(row.original.contactId) ?? "—",
    },
    {
      accessorKey: "billingType",
      header: t("colBillingType"),
      cell: ({ row }) => (
        <Badge variant="secondary">{dyn(`billing_${row.original.billingType}`)}</Badge>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => (
        <StateBadge
          tone={projectStateTone(row.original.state)}
          label={dyn(`state_${row.original.state}`)}
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
          {t("openProject")}
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
          { value: "draft", label: dyn("state_draft") },
          { value: "open", label: dyn("state_open") },
          { value: "closed", label: dyn("state_closed") },
          { value: "cancelled", label: dyn("state_cancelled") },
        ]}
        searchPlaceholder={t("searchPlaceholder")}
        filterLabel={t("filterStatus")}
        allLabel={t("allProjects")}
        ariaLabel={t("allProjects")}
        emptyTitle={t("emptyTitle")}
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
            <span>{tCommon("add")}</span>
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
