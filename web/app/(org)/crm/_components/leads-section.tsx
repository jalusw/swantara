"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CrmLead, CrmStage } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { LeadFormDialog } from "./lead-form-dialog";
import { PromoteDialog } from "./promote-section";

export function LeadsSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();

  const [leadDialogOpen, setLeadDialogOpen] = useState(false);
  const [editingLead, setEditingLead] = useState<CrmLead | null>(null);
  const [promoteId, setPromoteId] = useState<string | null>(null);

  const leadsQuery = useOrgListQuery<{ leads: CrmLead[] }, Record<string, never>>(
    "crmLeads",
    (organizationId) => getSwantaraService().crmLeads.list(organizationId),
  );
  const stagesQuery = useOrgListQuery<{ stages: CrmStage[] }, Record<string, never>>(
    "crmStages",
    (organizationId) => getSwantaraService().crmStages.list(organizationId),
  );

  const leads = leadsQuery.data?.leads ?? [];
  const stages = stagesQuery.data?.stages ?? [];
  const stageMap = new Map(stages.map((s) => [String(s.id), s.name]));

  const isLoading = leadsQuery.isLoading || stagesQuery.isLoading;
  const error = leadsQuery.error ?? stagesQuery.error;

  function refresh() {
    void leadsQuery.refetch();
    void stagesQuery.refetch();
    void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
    void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
    void queryClient.invalidateQueries({ queryKey: ["crmStages"] });
  }

  function handleLeadSave() {
    setLeadDialogOpen(false);
    setEditingLead(null);
    refresh();
  }

  const deleteMutation = useMutation({
    mutationFn: (prospectId: number) =>
      getSwantaraService().crmLeads.delete(Number(orgId), prospectId),
    onSuccess: () => {
      toast.success("Deleted.");
      refresh();
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const leadColumns: ColumnDef<CrmLead>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "contactName",
      header: "Contact",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.contactName ?? "—"}</span>
      ),
    },
    {
      accessorKey: "expectedRevenue",
      header: "Expected revenue",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums">{formatNumber(row.original.expectedRevenue)}</span>
      ),
    },
    {
      accessorKey: "probability",
      header: "Probability",
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{row.original.probability}%</span>,
    },
    {
      accessorKey: "stageId",
      header: "Stage",
      cell: ({ row }) => (
        <Badge variant="secondary">
          {row.original.stageId
            ? (stageMap.get(String(row.original.stageId)) ?? String(row.original.stageId))
            : "—"}
        </Badge>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-1">
          {row.original.type === "lead" && row.original.closedAt == null ? (
            <Button
              size="sm"
              variant="outline"
              onClick={() => setPromoteId(String(row.original.id))}
            >
              {"Promote to opportunity"}
            </Button>
          ) : null}
          <RowActions
            editLabel={"Edit lead"}
            deleteLabel={"Delete lead"}
            confirmTitle={"Delete this lead?"}
            confirmDescription={"The lead will be removed permanently."}
            onEdit={() => {
              setEditingLead(row.original);
              setLeadDialogOpen(true);
            }}
            onDelete={() => {
              deleteMutation.mutate(row.original.id);
            }}
          />
        </div>
      ),
    },
  ];

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">{"Leads"}</CardTitle>
            <p className="text-sm text-muted-foreground">
              {"Unqualified contacts. Promote a lead to create an opportunity."}
            </p>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={leadColumns}
            data={leads}
            getRowId={(row) => String(row.id)}
            searchKeys={["name", "contactName", "email"]}
            statusOptions={[]}
            searchPlaceholder={"Search leads…"}
            filterLabel={"Search leads…"}
            allLabel={"All leads"}
            ariaLabel={"All leads"}
            emptyTitle={"No leads"}
            status={
              isLoading
                ? { type: "loading" }
                : error
                  ? {
                      type: "error",
                      message: error.message,
                      onRetry: refresh,
                    }
                  : undefined
            }
            actions={
              <Button
                size="sm"
                onClick={() => {
                  setEditingLead(null);
                  setLeadDialogOpen(true);
                }}
              >
                <Plus />
                <span>{"Add lead"}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      {leadDialogOpen ? (
        <LeadFormDialog
          open={leadDialogOpen}
          onOpenChange={(open) => {
            if (!open) setEditingLead(null);
            setLeadDialogOpen(open);
          }}
          orgId={orgId}
          initial={editingLead}
          type="lead"
          onSave={handleLeadSave}
        />
      ) : null}
      {promoteId ? (
        <PromoteDialog
          open={Boolean(promoteId)}
          onOpenChange={(open) => {
            if (!open) setPromoteId(null);
          }}
          orgId={orgId}
          prospectId={promoteId}
          onPromoted={() => {
            setPromoteId(null);
            refresh();
          }}
        />
      ) : null}
    </>
  );
}
