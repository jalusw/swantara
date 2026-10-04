"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { KanbanBoard } from "@/components/kanban-board";
import type { CrmLead, CrmStage } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
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
  const t = useTranslations("Crm");
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
      const isOverdue =
        opp.expectedClose != null && new Date(opp.expectedClose).getTime() < Date.now();
      return {
        id: String(opp.id),
        columnId: opp.stageId != null ? String(opp.stageId) : (columns[0]?.id ?? "unassigned"),
        title: opp.name,
        subtitle: (
          <span className="font-medium text-foreground tabular-nums">
            {formatNumber(opp.expectedRevenue)}
            <span className="font-normal text-muted-foreground"> · {opp.probability}%</span>
          </span>
        ),
        meta: (
          <span className="flex flex-wrap items-center gap-1.5">
            {opp.contactName ? <span className="truncate">{opp.contactName}</span> : null}
            {opp.expectedClose ? (
              <Badge variant={isOverdue ? "destructive" : "outline"} className="tabular-nums">
                {formatDate(opp.expectedClose)}
              </Badge>
            ) : null}
          </span>
        ),
      };
    });

  const winMutation = useMutation({
    mutationFn: (cardId: string) =>
      getSwantaraService().crmOpportunities.win(Number(orgId), Number(cardId)),
    onSuccess: () => {
      toast.success(t("markedWon"));
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
      toast.success(t("stageUpdated"));
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
      {opportunities.some((o) => o.isWon) ? (
        <div className="flex flex-wrap gap-2">
          {opportunities
            .filter((o) => o.isWon)
            .slice(0, 3)
            .map((opp) => (
              <div
                key={opp.id}
                className="flex items-center gap-2 rounded-lg border border-success/30 bg-success/10 px-3 py-2 text-sm"
              >
                <span className="max-w-48 truncate font-medium">{opp.name}</span>
                <Badge variant="default">{t("wonBadge")}</Badge>
                <span className="text-xs text-muted-foreground tabular-nums">
                  {formatNumber(opp.expectedRevenue)}
                </span>
                <Button size="sm" variant="outline" asChild>
                  <a href={`/crm?won=${opp.id}`}>{t("createQuote")}</a>
                </Button>
              </div>
            ))}
        </div>
      ) : null}
      <KanbanBoard
        columns={columns}
        cards={cards}
        onMoveCard={handleMove}
        aria-label={t("pipelineTitle")}
      />
      <p className="text-xs text-muted-foreground">{t("boardHint")}</p>
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
                {t("markWon")}: {opp.name}
              </Button>
              <Button size="sm" variant="ghost" onClick={() => setLoseId(String(opp.id))}>
                {t("markLost")}
              </Button>
            </div>
          ))}
      </div>
    </div>
  );
}
