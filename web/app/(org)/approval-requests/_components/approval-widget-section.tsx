"use client";

import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Textarea } from "@/components/textarea";
import type { ApprovalRequest, ApprovalStep } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

const approvalStateClass: Record<string, string> = {
  approved: "border-success text-success",
  refused: "border-destructive text-destructive",
};

type ApprovalWidgetProps = {
  approvalRequest: ApprovalRequest | null;
  orgId: string;
  onDecided?: () => void;
};

function stepDecisionTone(
  decision: ApprovalStep["decision"],
): "neutral" | "success" | "danger" | "info" {
  switch (decision) {
    case "approved":
      return "success";
    case "refused":
      return "danger";
    case "pending":
      return "info";
    default:
      return "neutral";
  }
}

export function ApprovalWidget({ approvalRequest, orgId, onDecided }: ApprovalWidgetProps) {
  const [decideOpen, setDecideOpen] = useState(false);
  const [activeStep, setActiveStep] = useState<ApprovalStep | null>(null);
  const [decisionApprove, setDecisionApprove] = useState(true);
  const [comment, setComment] = useState("");

  const decisionMutation = useMutation({
    mutationFn: ({
      stepId,
      approve,
      comment,
    }: {
      stepId: number;
      approve: boolean;
      comment: string;
    }) =>
      getSwantaraService().approvalRequests.decide(Number(orgId), approvalRequest?.id ?? 0, {
        stepId,
        approve,
        comment,
      }),
    onSuccess: (_data, variables) => {
      toast.success(variables.approve ? "Approved." : "Refused.");
      setDecideOpen(false);
      onDecided?.();
    },
    onError: () => toast.error("Something went wrong."),
  });

  if (!approvalRequest) {
    return null;
  }

  const steps = approvalRequest.steps ?? [];

  function handleDecide(step: ApprovalStep, approve: boolean) {
    setActiveStep(step);
    setDecisionApprove(approve);
    setComment("");
    setDecideOpen(true);
  }

  function submitDecision() {
    if (!activeStep || !approvalRequest) return;
    decisionMutation.mutate({
      stepId: activeStep.id,
      approve: decisionApprove,
      comment,
    });
  }

  return (
    <>
      <div data-slot="approval-widget" className="flex flex-col gap-2 rounded-md border p-3">
        <div className="flex items-center justify-between">
          <span className="text-sm">{"Approval"}</span>
          <Badge variant="outline" className={approvalStateClass[approvalRequest.state] ?? ""}>
            {humanizeKey(String(approvalRequest.state))}
          </Badge>
        </div>
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
                      {"Step"} {step.sequence}
                    </span>
                    {step.decidedAt ? (
                      <span className="text-xs text-muted-foreground">
                        {formatDate(step.decidedAt)}
                      </span>
                    ) : null}
                    {step.comment ? (
                      <span className="text-xs text-muted-foreground italic">{step.comment}</span>
                    ) : null}
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline" className={stepDecisionTone(step.decision)}>
                      {humanizeKey(String(step.decision))}
                    </Badge>
                    {step.decision === "pending" ? (
                      <div className="flex gap-1">
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDecide(step, true)}
                        >
                          {"Approve"}
                        </Button>
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => handleDecide(step, false)}
                        >
                          {"Refuse"}
                        </Button>
                      </div>
                    ) : null}
                  </div>
                </div>
              ))}
          </div>
        ) : (
          <p className="text-xs text-muted-foreground">{"No approval steps configured."}</p>
        )}
      </div>
      <Dialog open={decideOpen} onOpenChange={setDecideOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{decisionApprove ? "Approve step" : "Refuse step"}</DialogTitle>
            <DialogDescription>{"Add an optional comment to your decision."}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Comment"}</span>
            <Textarea
              value={comment}
              onChange={(e) => setComment(e.target.value)}
              rows={3}
              placeholder={"Optional comment…"}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDecideOpen(false)}>
              {"Cancel"}
            </Button>
            <Button onClick={submitDecision} disabled={decisionMutation.isPending}>
              {decisionApprove ? "Approve" : "Refuse"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
