"use client";

import { useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CrmLead, CrmStage, PipelineForecast } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { CrmKpis, CrmPipelineKpiWall } from "./crm-kpis";
import { PipelineBoard } from "./pipeline-board";

export function PipelineSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Crm");
  const queryClient = useQueryClient();

  const leadsQuery = useOrgListQuery<{ leads: CrmLead[] }, Record<string, never>>(
    "crmLeads",
    (organizationId) => getSwantaraService().crmLeads.list(organizationId),
  );
  const opportunitiesQuery = useOrgListQuery<{ opportunities: CrmLead[] }, Record<string, never>>(
    "crmOpportunities",
    (organizationId) => getSwantaraService().crmOpportunities.list(organizationId),
  );
  const stagesQuery = useOrgListQuery<{ stages: CrmStage[] }, Record<string, never>>(
    "crmStages",
    (organizationId) => getSwantaraService().crmStages.list(organizationId),
  );
  const pipelineQuery = useOrgListQuery<{ pipeline: PipelineForecast }, Record<string, never>>(
    "crmPipeline",
    (organizationId) => getSwantaraService().crmPipeline.get(organizationId),
  );

  const leads = leadsQuery.data?.leads ?? [];
  const opportunities = opportunitiesQuery.data?.opportunities ?? [];
  const stages = stagesQuery.data?.stages ?? [];
  const pipeline = pipelineQuery.data?.pipeline ?? null;

  const isLoading =
    leadsQuery.isLoading ||
    opportunitiesQuery.isLoading ||
    stagesQuery.isLoading ||
    pipelineQuery.isLoading;
  const error =
    leadsQuery.error ?? opportunitiesQuery.error ?? stagesQuery.error ?? pipelineQuery.error;

  function refreshAll() {
    void leadsQuery.refetch();
    void opportunitiesQuery.refetch();
    void stagesQuery.refetch();
    void pipelineQuery.refetch();
    void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
    void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
    void queryClient.invalidateQueries({ queryKey: ["crmPipeline"] });
    void queryClient.invalidateQueries({ queryKey: ["crmStages"] });
  }

  if (isLoading) {
    return <p className="text-sm text-muted-foreground">{t("loadingPipeline")}</p>;
  }

  if (error) {
    return (
      <div className="flex flex-col gap-2">
        <p className="text-sm text-destructive">{error.message}</p>
        <button type="button" className="text-sm underline" onClick={refreshAll}>
          {t("retryLoad")}
        </button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <CrmKpis
        pipeline={pipeline}
        opportunities={opportunities}
        leads={leads}
        stagesCount={stages.length}
      />
      <div className="flex flex-col gap-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("kpiWeighted")}</CardTitle>
            <p className="text-sm text-muted-foreground">{t("kpiWeightedHint")}</p>
          </CardHeader>
          <CardContent>
            <CrmPipelineKpiWall pipeline={pipeline} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-base">{t("pipelineTitle")}</CardTitle>
            <p className="text-sm text-muted-foreground">{t("pipelineSubtitle")}</p>
          </CardHeader>
          <CardContent>
            {stages.length === 0 ? (
              <p className="text-sm text-muted-foreground">{t("emptyStages")}</p>
            ) : (
              <PipelineBoard
                orgId={orgId}
                stages={stages}
                opportunities={opportunities}
                onRefresh={refreshAll}
              />
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
