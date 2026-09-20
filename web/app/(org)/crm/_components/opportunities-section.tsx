"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus, Trophy, XCircle } from "lucide-react";
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
import { formatDate, formatNumber } from "@/lib/utils";
import { LeadFormDialog } from "./lead-form-dialog";
import { LostReasonDialog } from "./lost-reason-dialog";

export function OpportunitiesSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();

  const [leadDialogOpen, setLeadDialogOpen] = useState(false);
  const [editingLead, setEditingLead] = useState<CrmLead | null>(null);
  const [loseId, setLoseId] = useState<string | null>(null);

  const opportunitiesQuery = useOrgListQuery<{ opportunities: CrmLead[] }, Record<string, never>>(
    "crmOpportunities",
    (organizationId) => getSwantaraService().crmOpportunities.list(organizationId),
  );
  const stagesQuery = useOrgListQuery<{ stages: CrmStage[] }, Record<string, never>>(
    "crmStages",
    (organizationId) => getSwantaraService().crmStages.list(organizationId),
  );

  const opportunities = opportunitiesQuery.data?.opportunities ?? [];
  const stages = stagesQuery.data?.stages ?? [];
  const stageMap = new Map(stages.map((s) => [String(s.id), s.name]));

  const isLoading = opportunitiesQuery.isLoading || stagesQuery.isLoading;
  const error = opportunitiesQuery.error ?? stagesQuery.error;

  function refresh() {
    void opportunitiesQuery.refetch();
    void stagesQuery.refetch();
    void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
    void queryClient.invalidateQueries({ queryKey: ["crmPipeline"] });
    void queryClient.invalidateQueries({ queryKey: ["crmStages"] });
  }

  function handleLeadSave() {
    setLeadDialogOpen(false);
    setEditingLead(null);
    refresh();
  }

  const winMutation = useMutation({
    mutationFn: (prospectId: number) =>
      getSwantaraService().crmOpportunities.win(Number(orgId), prospectId),
    onSuccess: () => {
      toast.success("Mark won");
      refresh();
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (prospectId: number) =>
      getSwantaraService().crmOpportunities.delete(Number(orgId), prospectId),
    onSuccess: () => {
      toast.success("Deleted.");
      refresh();
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const oppColumns: ColumnDef<CrmLead>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "stageId",
      header: "Stage",
      cell: ({ row }) => {
        const stageName = row.original.stageId ? stageMap.get(String(row.original.stageId)) : null;
        return (
          <Badge
            variant={
              row.original.isWon ? "default" : row.original.lostReason ? "outline" : "secondary"
            }
          >
            {stageName ?? "—"}
          </Badge>
        );
      },
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
      accessorKey: "weighted",
      header: "Weighted",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">
          {formatNumber(
            Math.round(((row.original.expectedRevenue * row.original.probability) / 100) * 100) /
              100,
          )}
        </span>
      ),
    },
    {
      accessorKey: "expectedClose",
      header: "Expected close",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.expectedClose ? formatDate(row.original.expectedClose) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const isOpen =
          row.original.closedAt == null && row.original.lostReason == null && !row.original.isWon;
        return (
          <div className="flex items-center justify-end gap-1">
            {isOpen ? (
              <>
                <Button
                  size="icon-sm"
                  variant="ghost"
                  aria-label={"Mark won"}
                  onClick={() => winMutation.mutate(row.original.id)}
                  disabled={winMutation.isPending}
                >
                  <Trophy className="size-4" />
                </Button>
                <Button
                  size="icon-sm"
                  variant="ghost"
                  aria-label={"Mark lost"}
                  onClick={() => setLoseId(String(row.original.id))}
                >
                  <XCircle className="size-4" />
                </Button>
              </>
            ) : null}
            {row.original.isWon ? (
              <Button size="sm" variant="outline" asChild>
                <a href={`/crm?quotation=${row.original.id}`}>{"Create quotation"}</a>
              </Button>
            ) : null}
            <RowActions
              editLabel={"Edit opportunity"}
              deleteLabel={"Delete opportunity"}
              confirmTitle={"Delete this opportunity?"}
              confirmDescription={"The opportunity will be removed permanently."}
              onEdit={() => {
                setEditingLead(row.original);
                setLeadDialogOpen(true);
              }}
              onDelete={() => {
                deleteMutation.mutate(row.original.id);
              }}
            />
          </div>
        );
      },
    },
  ];

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="text-base">{"Opportunities"}</CardTitle>
            <p className="text-sm text-muted-foreground">
              {"Qualified deals on the pipeline. Move them through stages to closing."}
            </p>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={oppColumns}
            data={opportunities}
            getRowId={(row) => String(row.id)}
            searchKeys={["name"]}
            statusOptions={[]}
            searchPlaceholder={"Search opportunities…"}
            filterLabel={"Search opportunities…"}
            allLabel={"All opportunities"}
            ariaLabel={"All opportunities"}
            emptyTitle={"No opportunities"}
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
                <span>{"Add opportunity"}</span>
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
          type="opportunity"
          onSave={handleLeadSave}
        />
      ) : null}
      {loseId ? (
        <LostReasonDialog
          open={Boolean(loseId)}
          onOpenChange={(open) => {
            if (!open) setLoseId(null);
          }}
          orgId={orgId}
          opportunityId={loseId}
          onDone={() => {
            setLoseId(null);
            refresh();
          }}
        />
      ) : null}
    </>
  );
}
