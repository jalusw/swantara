"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { StateBadge } from "@/components/state-badge";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { QualityAlert } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { alertStateTone, alertTraceability } from "../../../_components/quality-utils";

export function QualityAlertDetail({ orgId, alertId }: { orgId: string; alertId: string }) {
  const t = useTranslations("Quality");
  const tCommon = useTranslations("Common");
  const alertState = (state: string) =>
    (t as unknown as (k: string) => string)(`alertState_${state}`);
  const severity = (level: string) => (t as unknown as (k: string) => string)(`severity_${level}`);
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
      toast.success(t("alertStatusUpdated"));
    },
    onError: () => {
      toast.error(t("saveFailed"));
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
    return <div className="text-muted-foreground text-sm">{tCommon("loading")}</div>;
  }

  if (error || !alert) {
    return <div className="text-destructive text-sm">{error?.message ?? t("alertNotFound")}</div>;
  }

  const tone = alertStateTone(alert.state);
  const transitions = getAvailableTransitions(alert.state);

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-lg">{alert.title || t("alertFallback", { id: alert.id })}</h1>
          <StateBadge tone={tone} label={alertState(alert.state)} />
          {alert.severity ? <Badge variant="outline">{severity(alert.severity)}</Badge> : null}
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
              {alertState(next)}
            </Button>
          ))}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("summary")}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">{t("description")}</p>
            <p className="text-sm">{alert.description || "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{t("traceability")}</p>
            <p className="text-sm">{alertTraceability(alert)}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{t("assignedTo")}</p>
            <p className="text-sm">
              {alert.assignedTo ? t("userFallback", { id: alert.assignedTo }) : "—"}
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
