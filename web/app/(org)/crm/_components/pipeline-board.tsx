"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { KanbanBoard } from "@/components/kanban-board";
import type { CrmLead, CrmStage } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { isLostStage, isWonStage } from "./crm-utils";
import { LostReasonDialog } from "./lost-reason-dialog";

export function PipelineBoard({
  orgId,
  stages,
  opportunities,
  onRefresh,
}: {
  orgId: string;
  stages: CrmStage[];
  opportunities: CrmLead[];
  onRefresh: () => void;
}) {
  const queryClient = useQueryClient();
  const [loseId, setLoseId] = useState<string | null>(null);

  const columns = stages.map((stage) => ({
    id: String(stage.id),
    title: (
      <span className="flex items-center gap-1">
        <span>{stage.name}</span>
        <Badge variant="secondary" className="text-[10px] leading-none">
          {stage.probability}%
        </Badge>
      </span>
    ),
  }));

  const stageMap = useMemo(() => new Map(stages.map((s) => [String(s.id), s])), [stages]);

  const cards = opportunities
    .filter((opp) => opp.closedAt == null && opp.lostReason == null)
    .map((opp) => {
      const stage = opp.stageId != null ? stageMap.get(String(opp.stageId)) : null;
      return {
        id: String(opp.id),
        columnId: opp.stageId != null ? String(opp.stageId) : (columns[0]?.id ?? "unassigned"),
        title: opp.name,
        subtitle: `${formatNumber(opp.expectedRevenue)} · ${opp.probability}%`,
        meta: stage ? <Badge variant="outline">{stage.name}</Badge> : null,
      };
    });

  const winMutation = useMutation({
    mutationFn: (cardId: string) =>
      getSwantaraService().crmOpportunities.win(Number(orgId), Number(cardId)),
    onSuccess: () => {
      toast.success("Opportunity marked as won.");
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities", Number(orgId)] });
      void queryClient.invalidateQueries({ queryKey: ["crmPipeline", Number(orgId)] });
      onRefresh();
    },
  });

  const advanceStageMutation = useMutation({
    mutationFn: ({ cardId, stageId }: { cardId: string; stageId: number }) =>
      getSwantaraService().crmOpportunities.advanceStage(Number(orgId), Number(cardId), {
        stageId,
      }),
    onSuccess: () => {
      toast.success("Stage updated.");
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities", Number(orgId)] });
      void queryClient.invalidateQueries({ queryKey: ["crmPipeline", Number(orgId)] });
      onRefresh();
    },
  });

  function handleMove(cardId: string, columnId: string) {
    const target = stageMap.get(columnId);
    if (!target) return;
    const opp = opportunities.find((o) => String(o.id) === cardId);
    if (!opp) return;
    if (opp.stageId != null && String(opp.stageId) === columnId) return;

    if (isWonStage(target)) {
      winMutation.mutate(cardId);
      return;
    }
    if (isLostStage(target)) {
      setLoseId(cardId);
      return;
    }
    advanceStageMutation.mutate({ cardId, stageId: Number(columnId) });
  }

  function handleWin(cardId: string) {
    winMutation.mutate(cardId);
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap gap-2">
        {opportunities
          .filter((o) => o.isWon)
          .slice(0, 3)
          .map((opp) => (
            <div
              key={opp.id}
              className="flex items-center gap-2 rounded-md border bg-success/10 px-3 py-2 text-sm"
            >
              <span className="">{opp.name}</span>
              <Badge variant="default">{"Won"}</Badge>
              <Button size="sm" variant="outline" asChild>
                <a href={`/crm?won=${opp.id}`}>{"Create quotation"}</a>
              </Button>
            </div>
          ))}
      </div>
      <KanbanBoard
        columns={columns}
        cards={cards}
        onMoveCard={handleMove}
        aria-label={"Pipeline board"}
      />
      <p className="text-xs text-muted-foreground">{"Drag or use the stage menu to move"}</p>
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
            onRefresh();
          }}
        />
      ) : null}
      <div className="flex flex-wrap gap-2 lg:hidden">
        {opportunities
          .filter((o) => o.closedAt == null && o.lostReason == null)
          .slice(0, 5)
          .map((opp) => (
            <div key={opp.id} className="flex gap-1">
              <Button size="sm" variant="outline" onClick={() => handleWin(String(opp.id))}>
                {"Mark won"}: {opp.name}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setLoseId(String(opp.id))}>
                {"Mark lost"}
              </Button>
            </div>
          ))}
      </div>
    </div>
  );
}
