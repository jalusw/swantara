import type { CrmLead, CrmStage, PipelineForecast } from "@/lib/services/swantara";

export function weightedRevenue(expectedRevenue: number, probability: number): number {
  return Math.round(((expectedRevenue * probability) / 100) * 100) / 100;
}

export function weightedPipeline(
  opportunities: Pick<CrmLead, "expectedRevenue" | "probability">[],
): number {
  let total = 0;
  for (const opp of opportunities) {
    total += (opp.expectedRevenue * opp.probability) / 100;
  }
  return Math.round(total * 100) / 100;
}

export function winRate(won: number, lost: number): number {
  const total = won + lost;
  if (total === 0) return 0;
  return Math.round((won / total) * 100 * 100) / 100;
}

export function isClosedLead(lead: Pick<CrmLead, "closedAt" | "lostReason" | "isWon">): boolean {
  return lead.closedAt != null || lead.isWon || lead.lostReason != null;
}

export function canPromote(lead: Pick<CrmLead, "type" | "closedAt">): boolean {
  return lead.type === "lead" && lead.closedAt == null;
}

export function canAdvanceStage(
  opportunity: Pick<CrmLead, "type" | "closedAt" | "lostReason">,
): boolean {
  return (
    opportunity.type === "opportunity" &&
    opportunity.closedAt == null &&
    opportunity.lostReason == null
  );
}

export function canWin(
  opportunity: Pick<CrmLead, "type" | "closedAt" | "lostReason" | "isWon">,
): boolean {
  return (
    opportunity.type === "opportunity" &&
    !opportunity.isWon &&
    opportunity.closedAt == null &&
    opportunity.lostReason == null
  );
}

export function canLose(
  opportunity: Pick<CrmLead, "type" | "closedAt" | "lostReason" | "isWon">,
): boolean {
  return (
    opportunity.type === "opportunity" &&
    !opportunity.isWon &&
    opportunity.closedAt == null &&
    opportunity.lostReason == null
  );
}

export function resolveWeightedForecast(pipeline: PipelineForecast | null | undefined): {
  weighted: number;
  total: number;
  rate: number;
} {
  if (!pipeline) return { weighted: 0, total: 0, rate: 0 };
  return {
    weighted: pipeline.weightedPipeline,
    total: pipeline.totalExpectedRevenue,
    rate: pipeline.winRate,
  };
}

export function stageProbabilityBadge(stage: CrmStage | null | undefined): string {
  if (!stage) return "—";
  return `${stage.probability}%`;
}

export function sortedStages(stages: CrmStage[]): CrmStage[] {
  return [...stages].sort((a, b) => a.sequence - b.sequence);
}

export function isWonStage(stage: CrmStage | null | undefined): boolean {
  return Boolean(stage?.isWon);
}

export function isLostStage(stage: CrmStage | null | undefined): boolean {
  if (!stage) return false;
  return !stage.isWon && stage.probability === 0;
}

export function opportunityStageId(opportunity: CrmLead): string {
  return opportunity.stageId != null ? String(opportunity.stageId) : "unassigned";
}
