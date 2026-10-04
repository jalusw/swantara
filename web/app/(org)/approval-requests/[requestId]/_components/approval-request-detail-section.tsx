"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { ApprovalWidget } from "@/app/(org)/approval-requests/_components/approval-widget-section";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Skeleton } from "@/components/skeleton";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

type ApprovalRequestDetailProps = {
  orgId: string;
  requestId: string;
};

function approvalStateTone(state: string): string {
  switch (state) {
    case "approved":
      return "border-success text-success";
    case "refused":
      return "border-destructive text-destructive";
    default:
      return "";
  }
}

export function ApprovalRequestDetail({ orgId, requestId }: ApprovalRequestDetailProps) {
  const t = useTranslations("ApprovalRequests");
  const requestState = (state: string) => (t as unknown as (k: string) => string)(`state_${state}`);
  const decisionLabel = (decision: string) =>
    (t as unknown as (k: string) => string)(`decision_${decision}`);
  const router = useRouter();
  const queryClient = useQueryClient();

  const query = useQuery({
    queryKey: ["approvalRequests", orgId, requestId],
    queryFn: () => getSwantaraService().approvalRequests.get(Number(orgId), Number(requestId)),
    enabled: Boolean(orgId && requestId),
  });

  const approvalRequest = query.data?.approvalRequest;
  const steps = approvalRequest?.steps ?? [];

  if (query.isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  if (!approvalRequest) {
    return null;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => router.back()} aria-label={t("back")}>
          <ArrowLeft />
        </Button>
        <h2 className="text-lg">{`AR-${approvalRequest.id}`}</h2>
        <Badge variant="outline" className={approvalStateTone(approvalRequest.state)}>
          {requestState(approvalRequest.state)}
        </Badge>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("summary")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colType")}</span>
              <span>{approvalRequest.ownerType}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colRecord")}</span>
              <span>{`#${approvalRequest.ownerId}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colRequestedBy")}</span>
              <span>{`#${approvalRequest.requestedBy}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colCreatedAt")}</span>
              <span>{formatDate(approvalRequest.createdAt)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colUpdatedAt")}</span>
              <span>{formatDate(approvalRequest.updatedAt)}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("approvalSteps")}</CardTitle>
          </CardHeader>
          <CardContent>
            {steps.length > 0 ? (
              <div className="flex flex-col gap-2">
                {[...steps]
                  .sort((a, b) => a.sequence - b.sequence)
                  .map((step) => (
                    <div
                      key={step.id}
                      className="flex items-center justify-between gap-2 rounded border px-3 py-2"
                    >
                      <div className="flex flex-col gap-0.5">
                        <span className="text-sm">
                          {t("step")} {step.sequence}
                        </span>
                        {step.decidedAt ? (
                          <span className="text-xs text-muted-foreground">
                            {formatDate(step.decidedAt)}
                          </span>
                        ) : null}
                        {step.comment ? (
                          <span className="text-xs text-muted-foreground italic">
                            {step.comment}
                          </span>
                        ) : null}
                      </div>
                      <Badge variant="outline" className={approvalStateTone(step.decision)}>
                        {decisionLabel(step.decision)}
                      </Badge>
                    </div>
                  ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">{t("noSteps")}</p>
            )}
          </CardContent>
        </Card>
      </div>

      <ApprovalWidget
        approvalRequest={approvalRequest}
        orgId={orgId}
        onDecided={() =>
          void queryClient
            .invalidateQueries({
              queryKey: ["approvalRequests", orgId, requestId],
            })
            .then(() =>
              queryClient.invalidateQueries({
                queryKey: ["approvalRequests", orgId],
              }),
            )
        }
      />
    </div>
  );
}
