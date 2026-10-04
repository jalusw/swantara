"use client";

import { useMutation } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { ApprovalWidget } from "@/app/(org)/approval-requests/_components/approval-widget-section";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { RecordLayout } from "@/components/record-layout";
import { StateBadge } from "@/components/state-badge";
import { WorkflowSteps } from "@/components/workflow-steps";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type {
  ApprovalRequest,
  Contact,
  Department,
  PurchaseRequest,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";

import {
  canApprove,
  canCancel,
  canConfirm,
  canCreateQuoteRequest,
} from "../../_components/purchase-request-utils";

const requisitionSteps = ["draft", "confirmed", "approved", "done"] as const;

function requisitionStateIndex(state: PurchaseRequest["state"]): number {
  const idx = requisitionSteps.indexOf(state as (typeof requisitionSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function getRequisitionStatuses(label?: (state: string) => string) {
  const text = (state: string, fallback: string) => (label ? label(state) : fallback);
  return {
    draft: { label: text("draft", "Draft"), tone: "neutral" as const },
    confirmed: { label: text("confirmed", "Dikonfirmasi"), tone: "info" as const },
    approved: { label: text("approved", "Disetujui"), tone: "success" as const },
    done: { label: text("done", "Selesai"), tone: "success" as const },
    cancelled: { label: text("cancelled", "Dibatalkan"), tone: "danger" as const },
  };
}

export function PurchaseRequestDetail({ orgId, requestId }: { orgId: string; requestId: string }) {
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
  const [quoteRequestDialogOpen, setQuoteRequestDialogOpen] = useState(false);

  const requisitionQuery = useOrgQuery<{ request: PurchaseRequest }>(
    "purchaseRequest",
    requestId,
    (organizationId) =>
      getSwantaraService().purchaseRequests.get(organizationId, Number(requestId)),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );
  const approvalQuery = useOrgListQuery<
    { approvalRequests: ApprovalRequest[] },
    Record<string, never>,
    ApprovalRequest | null
  >(
    "approvalRequests",
    (organizationId) => getSwantaraService().approvalRequests.list(organizationId),
    {},
    {
      select: (data) =>
        data.approvalRequests.find(
          (a) => a.ownerType === "purchase_requisition" && a.ownerId === Number(requestId),
        ) ?? null,
    },
  );

  const request = requisitionQuery.data?.request ?? null;
  const contacts = contactsQuery.data?.contacts ?? [];
  const departments = departmentsQuery.data?.departments ?? [];

  const contactMap = useMemo(
    () => new Map(contacts.map((p) => [p.id, p.displayName || p.name])),
    [contacts],
  );
  const departmentMap = useMemo(
    () => new Map(departments.map((d) => [d.id, d.name])),
    [departments],
  );

  const linkedApproval = approvalQuery.data ?? null;

  const confirmMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().purchaseRequests.confirm(Number(orgId), requestId),
    onSuccess: () => {
      toast.success(t("requestConfirmed"));
      void requisitionQuery.refetch();
    },
    onError: () => toast.error(t("saveFailed")),
  });

  const approveMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().purchaseRequests.approve(Number(orgId), requestId),
    onSuccess: () => {
      toast.success(t("requestApproved"));
      void requisitionQuery.refetch();
    },
    onError: () => toast.error(t("saveFailed")),
  });

  const cancelMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().purchaseRequests.cancel(Number(orgId), requestId),
    onSuccess: () => {
      toast.success(t("requestCancelled"));
      void requisitionQuery.refetch();
    },
    onError: () => toast.error(t("saveFailed")),
  });

  const createQuoteRequestMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().supplierQuoteRequests.fromRequisition(Number(orgId), {
        requestId,
      }),
    onSuccess: () => {
      toast.success(t("quoteRequestCreatedFromRequest"));
      setQuoteRequestDialogOpen(false);
      void requisitionQuery.refetch();
    },
    onError: () => toast.error(t("saveFailed")),
  });

  if (requisitionQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!request) {
    return <p className="text-sm text-muted-foreground">{t("purchaseRequestNotFound")}</p>;
  }

  function requisitionStateLabel(state: string): string {
    try {
      return (t as unknown as (k: string) => string)(`requisitionState.${state}`);
    } catch {
      return state;
    }
  }

  function requisitionStepLabel(step: string): string {
    try {
      return (t as unknown as (k: string) => string)(`requisitionStep.${step}`);
    } catch {
      return String(step);
    }
  }

  const lines = request.lines ?? [];
  const breadcrumbItems = [
    { label: t("allRequisitions"), href: "/purchases/requisitions" },
    { label: request.name ?? `PR-${request.id}` },
  ];
  const steps = requisitionSteps.map((s) => ({ label: requisitionStepLabel(s) }));

  return (
    <>
      <RecordLayout
        breadcrumbItems={breadcrumbItems}
        title={request.name ?? `PR-${request.id}`}
        status={
          <StateBadge
            value={request.state}
            statuses={getRequisitionStatuses(requisitionStateLabel)}
          />
        }
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps steps={steps} currentIndex={requisitionStateIndex(request.state)} />
                <div className="flex flex-wrap gap-2">
                  {canConfirm(request.state) ? (
                    <Button
                      size="sm"
                      disabled={confirmMutation.isPending}
                      onClick={() => confirmMutation.mutate(request.id)}
                    >
                      {t("actionConfirm")}
                    </Button>
                  ) : null}
                  {canApprove(request.state) ? (
                    <Button
                      size="sm"
                      disabled={approveMutation.isPending}
                      onClick={() => approveMutation.mutate(request.id)}
                    >
                      {t("actionApprove")}
                    </Button>
                  ) : null}
                  {canCancel(request.state) ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={cancelMutation.isPending}
                      onClick={() => cancelMutation.mutate(request.id)}
                    >
                      {tCommon("cancel")}
                    </Button>
                  ) : null}
                  {canCreateQuoteRequest(request.state) ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={createQuoteRequestMutation.isPending}
                      onClick={() => setQuoteRequestDialogOpen(true)}
                    >
                      {t("createQuoteRequest")}
                    </Button>
                  ) : null}
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{t("requestHeader")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableRequester")}</span>
                        <span className="">
                          {contactMap.get(request.requesterId) ?? `#${request.requesterId}`}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableDepartment")}</span>
                        <span>
                          {request.departmentId
                            ? (departmentMap.get(request.departmentId) ??
                              `#${request.departmentId}`)
                            : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("fieldNeededBy")}</span>
                        <span>{request.neededBy ? formatDate(request.neededBy) : "—"}</span>
                      </div>
                    </CardContent>
                  </Card>
                  {linkedApproval ? (
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base">{t("approvalTitle")}</CardTitle>
                      </CardHeader>
                      <CardContent>
                        <ApprovalWidget
                          approvalRequest={linkedApproval}
                          orgId={orgId}
                          onDecided={() => void requisitionQuery.refetch()}
                        />
                      </CardContent>
                    </Card>
                  ) : null}
                </div>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{t("orderLines")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {lines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{t("emptyLines")}</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b text-left text-muted-foreground">
                              <th className="pb-2 ">#</th>
                              <th className="pb-2 ">{t("fieldItem")}</th>
                              <th className="pb-2 ">{t("fieldDescription")}</th>
                              <th className="pb-2 text-right">{t("fieldQty")}</th>
                              <th className="pb-2 ">{t("fieldNeededBy")}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {lines.map((line, idx) => (
                              <tr key={line.id} className="border-b last:border-0">
                                <td className="py-2">{idx + 1}</td>
                                <td className="py-2">{line.itemId ?? "—"}</td>
                                <td className="py-2 text-muted-foreground">
                                  {line.description ?? "—"}
                                </td>
                                <td className="py-2 text-right tabular-nums">
                                  {formatNumber(line.qty)}
                                </td>
                                <td className="py-2 text-muted-foreground">
                                  {line.neededBy ? formatDate(line.neededBy) : "—"}
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            ),
          },
        ]}
      />
      <Dialog open={quoteRequestDialogOpen} onOpenChange={setQuoteRequestDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("createQuoteRequest")}</DialogTitle>
            <DialogDescription>{t("createQuoteRequestDescription")}</DialogDescription>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("createQuoteRequestHint")}</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setQuoteRequestDialogOpen(false)}>
              {tCommon("cancel")}
            </Button>
            <Button onClick={() => createQuoteRequestMutation.mutate(request.id)}>
              {t("createQuoteRequest")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
