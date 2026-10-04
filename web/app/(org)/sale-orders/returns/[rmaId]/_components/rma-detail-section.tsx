"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Skeleton } from "@/components/skeleton";
import { WorkflowSteps } from "@/components/workflow-steps";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import { getSwantaraService, type Journal, type Rma, type RmaLine } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import {
  canCancel,
  canConfirm,
  canDone,
  canReceive,
  canRefund,
  dispositionLabel,
  dispositionTone,
  type RmaState,
  rmaStateLabel,
  rmaStateTone,
  rmaSteps,
  rmaTypeLabel,
  rmaValidStates,
  totalQty,
} from "../../_components/rma-utils";
import { ReceiveDialog } from "./receive-dialog";
import { RefundDialog } from "./refund-dialog";

type RmaDetailSectionProps = {
  orgId: string;
  rmaId: string;
};

export function RmaDetailSection({ orgId, rmaId }: RmaDetailSectionProps) {
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const router = useRouter();
  const queryClient = useQueryClient();
  const [receiveDialogOpen, setReceiveDialogOpen] = useState(false);
  const [refundDialogOpen, setRefundDialogOpen] = useState(false);

  const rmaQuery = useOrgQuery<{ rma: Rma; lines: RmaLine[] }>("rmas", rmaId, (organizationId) =>
    getSwantaraService().rmas.get(organizationId, Number(rmaId)),
  );

  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
    undefined,
    {
      select: (data) => ({
        journals: data.journals.filter((j) => j.type === "purchase" || j.type === "general"),
      }),
    },
  );

  const confirmMutation = useMutation({
    mutationFn: () => getSwantaraService().rmas.confirm(Number(orgId), Number(rmaId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const doneMutation = useMutation({
    mutationFn: () => getSwantaraService().rmas.done(Number(orgId), Number(rmaId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const cancelMutation = useMutation({
    mutationFn: () => getSwantaraService().rmas.cancel(Number(orgId), Number(rmaId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const rma = rmaQuery.data?.rma;
  const lines = rmaQuery.data?.lines ?? [];
  const journals = journalsQuery.data?.journals ?? [];

  if (rmaQuery.isLoading) {
    return (
      <div className="flex flex-col gap-4">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  if (!rma) {
    return null;
  }

  const rawState = rma.state;
  const state = rmaValidStates.includes(rawState as RmaState) ? (rawState as RmaState) : null;
  if (!state) return null;
  const steps = rmaSteps(state);

  function stateLabel(state: RmaState): string {
    try {
      return (t as unknown as (k: string) => string)(`rmaState.${state}`);
    } catch {
      return rmaStateLabel(state);
    }
  }

  function typeLabel(type: Rma["type"]): string {
    try {
      return (t as unknown as (k: string) => string)(`rmaType.${type}`);
    } catch {
      return rmaTypeLabel(type);
    }
  }

  function dispositionLabelText(disposition: RmaLine["disposition"]): string {
    try {
      return (t as unknown as (k: string) => string)(`disposition.${disposition}`);
    } catch {
      return dispositionLabel(disposition);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => router.back()}>
          <ArrowLeft />
        </Button>
        <h2 className="text-lg">{rma.name ?? `RMA-${rma.id}`}</h2>
        <Badge variant="outline" className={rmaStateTone(state)}>
          {stateLabel(state)}
        </Badge>
      </div>

      <WorkflowSteps
        currentIndex={steps.indexOf(state)}
        steps={steps.map((step) => ({
          label: stateLabel(step as RmaState),
          current: step === state,
          completed: steps.indexOf(step) < steps.indexOf(state),
        }))}
      />

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("sectionSummary")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("tableType")}</span>
              <span>{typeLabel(rma.type)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("fieldContact")}</span>
              <span>{`#${rma.contactId}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("fieldOriginOrderType")}</span>
              <span>
                {rma.originOrderType && rma.originOrderId
                  ? `${rma.originOrderType.toUpperCase()}-${rma.originOrderId}`
                  : "—"}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("fieldReason")}</span>
              <span>{rma.reason || "—"}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("tableCreated")}</span>
              <span>{formatDate(rma.createdAt)}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("sectionActions")}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <div className="flex flex-wrap gap-2">
              {canConfirm(state) && (
                <Button
                  size="sm"
                  onClick={() => confirmMutation.mutate()}
                  disabled={confirmMutation.isPending}
                >
                  {t("actionConfirm")}
                </Button>
              )}
              {canReceive(state) && (
                <Button size="sm" onClick={() => setReceiveDialogOpen(true)}>
                  {t("actionReceive")}
                </Button>
              )}
              {canRefund(state) && (
                <Button size="sm" onClick={() => setRefundDialogOpen(true)}>
                  {t("actionRefund")}
                </Button>
              )}
              {canDone(state) && (
                <Button
                  size="sm"
                  onClick={() => doneMutation.mutate()}
                  disabled={doneMutation.isPending}
                >
                  {t("actionDone")}
                </Button>
              )}
              {canCancel(state) && (
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => cancelMutation.mutate()}
                  disabled={cancelMutation.isPending}
                >
                  {tCommon("cancel")}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("returnLines")}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="py-2 text-left ">{t("fieldItem")}</th>
                  <th className="py-2 text-right ">{t("fieldQty")}</th>
                  <th className="py-2 text-left ">{t("fieldDisposition")}</th>
                  <th className="py-2 text-left ">{t("stockMove")}</th>
                  <th className="py-2 text-left ">{t("creditNote")}</th>
                </tr>
              </thead>
              <tbody>
                {lines.map((line) => (
                  <tr key={line.id} className="border-b">
                    <td className="py-2">{`#${line.itemId}`}</td>
                    <td className="py-2 text-right tabular-nums">{line.qty}</td>
                    <td className="py-2">
                      <Badge variant="outline" className={dispositionTone(line.disposition)}>
                        {dispositionLabelText(line.disposition)}
                      </Badge>
                    </td>
                    <td className="py-2">
                      {line.stockMovementId ? (
                        <span className="text-muted-foreground">{`#${line.stockMovementId}`}</span>
                      ) : (
                        "—"
                      )}
                    </td>
                    <td className="py-2">
                      {line.creditNoteId ? (
                        <span className="text-muted-foreground">{`#${line.creditNoteId}`}</span>
                      ) : (
                        "—"
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr className="border-t">
                  <td className="py-2 ">{t("tableTotal")}</td>
                  <td className="py-2 text-right tabular-nums ">{totalQty(lines)}</td>
                  <td colSpan={3} />
                </tr>
              </tfoot>
            </table>
          </div>
        </CardContent>
      </Card>

      {receiveDialogOpen && (
        <ReceiveDialog
          open={receiveDialogOpen}
          onOpenChange={setReceiveDialogOpen}
          orgId={orgId}
          rmaId={rmaId}
          journals={journals}
          onSave={() => {
            setReceiveDialogOpen(false);
            void queryClient.invalidateQueries({
              queryKey: ["rmas", orgId],
            });
          }}
        />
      )}

      {refundDialogOpen && (
        <RefundDialog
          open={refundDialogOpen}
          onOpenChange={setRefundDialogOpen}
          orgId={orgId}
          rmaId={rmaId}
          journals={journals}
          onSave={() => {
            setRefundDialogOpen(false);
            void queryClient.invalidateQueries({
              queryKey: ["rmas", orgId],
            });
          }}
        />
      )}
    </div>
  );
}
