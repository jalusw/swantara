"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { useRouter } from "next/navigation";
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
      toast.error("Could not disable the organization.");
    },
  });

  const doneMutation = useMutation({
    mutationFn: () => getSwantaraService().rmas.done(Number(orgId), Number(rmaId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  const cancelMutation = useMutation({
    mutationFn: () => getSwantaraService().rmas.cancel(Number(orgId), Number(rmaId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
    },
    onError: () => {
      toast.error("Could not disable the organization.");
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

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" onClick={() => router.back()}>
          <ArrowLeft />
        </Button>
        <h2 className="text-lg">{rma.name ?? `RMA-${rma.id}`}</h2>
        <Badge variant="outline" className={rmaStateTone(state)}>
          {rmaStateLabel(state)}
        </Badge>
      </div>

      <WorkflowSteps
        currentIndex={steps.indexOf(state)}
        steps={steps.map((step) => ({
          label: rmaStateLabel(step as RmaState),
          current: step === state,
          completed: steps.indexOf(step) < steps.indexOf(state),
        }))}
      />

      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{"Summary"}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Type"}</span>
              <span>{rmaTypeLabel(rma.type)}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Contact"}</span>
              <span>{`#${rma.contactId}`}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Origin order"}</span>
              <span>
                {rma.originOrderType && rma.originOrderId
                  ? `${rma.originOrderType.toUpperCase()}-${rma.originOrderId}`
                  : "—"}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Reason"}</span>
              <span>{rma.reason || "—"}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{"Created"}</span>
              <span>{formatDate(rma.createdAt)}</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{"Actions"}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col gap-2">
            <div className="flex flex-wrap gap-2">
              {canConfirm(state) && (
                <Button
                  size="sm"
                  onClick={() => confirmMutation.mutate()}
                  disabled={confirmMutation.isPending}
                >
                  {"Confirm"}
                </Button>
              )}
              {canReceive(state) && (
                <Button size="sm" onClick={() => setReceiveDialogOpen(true)}>
                  {"Receive"}
                </Button>
              )}
              {canRefund(state) && (
                <Button size="sm" onClick={() => setRefundDialogOpen(true)}>
                  {"Refund"}
                </Button>
              )}
              {canDone(state) && (
                <Button
                  size="sm"
                  onClick={() => doneMutation.mutate()}
                  disabled={doneMutation.isPending}
                >
                  {"Done"}
                </Button>
              )}
              {canCancel(state) && (
                <Button
                  size="sm"
                  variant="destructive"
                  onClick={() => cancelMutation.mutate()}
                  disabled={cancelMutation.isPending}
                >
                  {"Cancel"}
                </Button>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{"Return lines"}</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b">
                  <th className="py-2 text-left ">{"Item"}</th>
                  <th className="py-2 text-right ">{"Quantity"}</th>
                  <th className="py-2 text-left ">{"Disposition"}</th>
                  <th className="py-2 text-left ">{"Stock move"}</th>
                  <th className="py-2 text-left ">{"Credit note"}</th>
                </tr>
              </thead>
              <tbody>
                {lines.map((line) => (
                  <tr key={line.id} className="border-b">
                    <td className="py-2">{`#${line.itemId}`}</td>
                    <td className="py-2 text-right tabular-nums">{line.qty}</td>
                    <td className="py-2">
                      <Badge variant="outline" className={dispositionTone(line.disposition)}>
                        {dispositionLabel(line.disposition)}
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
                  <td className="py-2 ">{"Total"}</td>
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
