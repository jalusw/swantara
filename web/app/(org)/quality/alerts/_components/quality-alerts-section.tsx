"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { KanbanBoard } from "@/components/kanban-board";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { QualityAlert } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import {
  alertStateLabel,
  alertStateSteps,
  canTransitionAlert,
} from "../../_components/quality-utils";

export function QualityAlertsSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();

  const query = useOrgListQuery<{ alerts: QualityAlert[] }, Record<string, never>>(
    "qualityAlerts",
    (organizationId) => getSwantaraService().qualityAlerts.list(organizationId),
  );

  const alerts = query.data?.alerts ?? [];
  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  const columns = alertStateSteps.map((state: QualityAlert["state"]) => ({
    id: state,
    title: (
      <span className="flex items-center gap-1">
        <span>{alertStateLabel(state)}</span>
        <Badge variant="secondary" className="text-[10px] leading-none">
          {alerts.filter((a) => a.state === state).length}
        </Badge>
      </span>
    ),
  }));

  const cards = alerts
    .filter((a) => a.state !== "cancelled")
    .map((alert) => ({
      id: String(alert.id),
      columnId: alert.state,
      title: alert.title || `Alert #${alert.id}`,
      subtitle: alert.description
        ? alert.description.length > 60
          ? `${alert.description.slice(0, 60)}…`
          : alert.description
        : undefined,
      meta: alert.severity ? <Badge variant="outline">{alert.severity}</Badge> : null,
    }));

  const stateMutation = useMutation({
    mutationFn: ({ cardId, next }: { cardId: string; next: QualityAlert["state"] }) =>
      getSwantaraService().qualityAlerts.state(Number(orgId), Number(cardId), { state: next }),
    onSuccess: () => {
      toast.success("Alert status updated.");
      void queryClient.invalidateQueries({ queryKey: ["qualityAlerts"] });
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  function handleMove(cardId: string, columnId: string) {
    const next = columnId as QualityAlert["state"];
    const alert = alerts.find((a) => String(a.id) === cardId);
    if (!alert) return;
    if (alert.state === next) return;
    if (!canTransitionAlert(alert.state, next)) return;

    stateMutation.mutate({ cardId, next });
  }

  function getAvailableTransitions(state: QualityAlert["state"]) {
    const transitions: QualityAlert["state"][] = [];
    if (state === "open") {
      transitions.push("in_progress", "cancelled");
    } else if (state === "in_progress") {
      transitions.push("solved", "cancelled");
    }
    return transitions;
  }

  if (isLoading) {
    return <div className="text-muted-foreground text-sm">{"Loading…"}</div>;
  }

  if (error) {
    return (
      <div className="flex flex-col gap-2">
        <p className="text-destructive text-sm">{error.message}</p>
        <Button variant="outline" size="sm" onClick={() => void query.refetch()}>
          {"Retry"}
        </Button>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <KanbanBoard
        columns={columns}
        cards={cards}
        onMoveCard={handleMove}
        aria-label={"Quality Alerts"}
      />
      <p className="text-xs text-muted-foreground">
        {"Drag cards between columns to update status."}
      </p>
      <div className="flex flex-wrap gap-2 lg:hidden">
        {alerts
          .filter((a) => a.state !== "cancelled" && a.state !== "solved")
          .slice(0, 5)
          .map((alert) => {
            const transitions = getAvailableTransitions(alert.state);
            return (
              <div key={alert.id} className="flex gap-1">
                {transitions.map((next) => (
                  <Button
                    key={next}
                    size="sm"
                    variant="outline"
                    onClick={() => handleMove(String(alert.id), next)}
                  >
                    {alertStateLabel(next)}: {alert.title || `#${alert.id}`}
                  </Button>
                ))}
              </div>
            );
          })}
      </div>
    </div>
  );
}
