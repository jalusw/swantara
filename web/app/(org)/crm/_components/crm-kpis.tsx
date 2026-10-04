"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { CrmLead, PipelineForecast } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { weightedPipeline, winRate } from "./crm-utils";

export function CrmKpis({
  pipeline,
  opportunities,
  leads,
  stagesCount,
}: {
  pipeline: PipelineForecast | null | undefined;
  opportunities: CrmLead[];
  leads: CrmLead[];
  stagesCount: number;
}) {
  const t = useTranslations("Crm");
  const liveWeighted = weightedPipeline(
    opportunities.filter((o) => o.closedAt == null && o.lostReason == null),
  );
  const liveTotal = opportunities
    .filter((o) => o.closedAt == null && o.lostReason == null)
    .reduce((sum, o) => sum + o.expectedRevenue, 0);
  const won = opportunities.filter((o) => o.isWon).length;
  const lost = opportunities.filter((o) => o.lostReason != null).length;
  const liveWinRate = winRate(won, lost);

  const weighted = pipeline?.weightedPipeline ?? liveWeighted;
  const total = pipeline?.totalExpectedRevenue ?? liveTotal;
  const rate = pipeline?.winRate ?? liveWinRate;

  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">{t("kpiWeighted")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl tabular-nums">{formatNumber(weighted)}</p>
          <p className="text-xs text-muted-foreground">{t("kpiWeightedHint")}</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">{t("kpiExpectedRevenue")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl tabular-nums">{formatNumber(total)}</p>
          <p className="text-xs text-muted-foreground">
            {t("kpiOpenDeals")}:{" "}
            {opportunities.filter((o) => o.closedAt == null && o.lostReason == null).length}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">{t("kpiWinRate")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl tabular-nums">{formatNumber(rate)}%</p>
          <p className="text-xs text-muted-foreground">
            {t("kpiWon")}: {won} · {t("kpiLost")}: {lost}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-sm text-muted-foreground">{t("kpiOpenDeals")}</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-2xl tabular-nums">
            {opportunities.length + leads.filter((l) => l.type === "lead").length}
          </p>
          <p className="text-xs text-muted-foreground">
            {stagesCount} {t("kpiWinRateHint")}
          </p>
        </CardContent>
      </Card>
    </div>
  );
}

export function CrmPipelineKpiWall({
  pipeline,
}: {
  pipeline: PipelineForecast | null | undefined;
}) {
  const t = useTranslations("Crm");
  if (!pipeline || pipeline.stages.length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noForecast")}</p>;
  }
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
      {pipeline.stages.map((stage) => (
        <Card key={stage.stageId}>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm text-muted-foreground">{stage.stageName}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-lg tabular-nums">{formatNumber(stage.weightedRevenue)}</p>
            <p className="text-xs text-muted-foreground">
              {stage.opportunityCount} · {formatNumber(stage.expectedRevenue)}
            </p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
