"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { StateBadge } from "@/components/state-badge";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { QualityAlert } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import {
  alertStateLabel,
  alertStateTone,
  alertTraceability,
  severityLabel,
} from "../../../_components/quality-utils";

export function QualityAlertDetail({ orgId, alertId }: { orgId: string; alertId: string }) {
  const queryClient = useQueryClient();

  const query = useOrgQuery<{ alert: QualityAlert }>("qualityAlert", alertId, (organizationId) =>
    getSwantaraService().qualityAlerts.get(organizationId, Number(alertId)),
  );

  const alert = query.data?.alert;
  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  const transitionMutation = useMutation({
    mutationFn: (next: QualityAlert["state"]) =>
      getSwantaraService().qualityAlerts.state(Number(orgId), Number(alertId), { state: next }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["qualityAlert", Number(orgId), Number(alertId)],
      });
      toast.success("Alert status updated.");
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  function handleTransition(next: QualityAlert["state"]) {
    if (!alert) return;
    transitionMutation.mutate(next);
  }

  function getAvailableTransitions(state: QualityAlert["state"]): QualityAlert["state"][] {
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

  if (error || !alert) {
    return (
      <div className="text-destructive text-sm">{error?.message ?? "Quality alert not found."}</div>
    );
  }

  const tone = alertStateTone(alert.state);
  const transitions = getAvailableTransitions(alert.state);

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-lg">{alert.title || `Alert #${alert.id}`}</h1>
          <StateBadge tone={tone} label={alertStateLabel(alert.state)} />
          {alert.severity ? <Badge variant="outline">{severityLabel(alert.severity)}</Badge> : null}
        </div>
        <div className="flex gap-2">
          {transitions.map((next) => (
            <Button
              key={next}
              size="sm"
              variant={
                next === "solved" ? "default" : next === "cancelled" ? "destructive" : "outline"
              }
              onClick={() => handleTransition(next)}
            >
              {alertStateLabel(next)}
            </Button>
          ))}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{"Overview"}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">{"Description"}</p>
            <p className="text-sm">{alert.description || "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Traceability"}</p>
            <p className="text-sm">{alertTraceability(alert)}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Assigned to"}</p>
            <p className="text-sm">{alert.assignedTo ? `User #${alert.assignedTo}` : "—"}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
