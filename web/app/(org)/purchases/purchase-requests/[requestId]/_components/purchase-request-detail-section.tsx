"use client";

import { useMutation } from "@tanstack/react-query";
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

function getRequisitionStatuses() {
  return {
    draft: { label: "Draft", tone: "neutral" as const },
    confirmed: { label: "Confirmed", tone: "info" as const },
    approved: { label: "Approved", tone: "success" as const },
    done: { label: "Done", tone: "success" as const },
    cancelled: { label: "Cancelled", tone: "danger" as const },
  };
}

export function PurchaseRequestDetail({ orgId, requestId }: { orgId: string; requestId: string }) {
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
      toast.success("Request confirmed.");
      void requisitionQuery.refetch();
    },
    onError: () => toast.error("Could not disable the organization."),
  });

  const approveMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().purchaseRequests.approve(Number(orgId), requestId),
    onSuccess: () => {
      toast.success("Request approved.");
      void requisitionQuery.refetch();
    },
    onError: () => toast.error("Could not disable the organization."),
  });

  const cancelMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().purchaseRequests.cancel(Number(orgId), requestId),
    onSuccess: () => {
      toast.success("Request cancelled.");
      void requisitionQuery.refetch();
    },
    onError: () => toast.error("Could not disable the organization."),
  });

  const createQuoteRequestMutation = useMutation({
    mutationFn: (requestId: number) =>
      getSwantaraService().supplierQuoteRequests.fromRequisition(Number(orgId), {
        requestId,
      }),
    onSuccess: () => {
      toast.success("QuoteRequest created from request.");
      setQuoteRequestDialogOpen(false);
      void requisitionQuery.refetch();
    },
    onError: () => toast.error("Could not disable the organization."),
  });

  if (requisitionQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!request) {
    return <p className="text-sm text-muted-foreground">{"Purchase request not found."}</p>;
  }

  const lines = request.lines ?? [];
  const breadcrumbItems = [
    { label: "All purchase requisitions", href: "/purchases/requisitions" },
    { label: request.name ?? `PR-${request.id}` },
  ];
  const steps = requisitionSteps.map((s) => ({ label: String(s) }));

  return (
    <>
      <RecordLayout
        breadcrumbItems={breadcrumbItems}
        title={request.name ?? `PR-${request.id}`}
        status={<StateBadge value={request.state} statuses={getRequisitionStatuses()} />}
        tabs={[
          {
            id: "overview",
            label: "Overview",
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
                      {"Confirm"}
                    </Button>
                  ) : null}
                  {canApprove(request.state) ? (
                    <Button
                      size="sm"
                      disabled={approveMutation.isPending}
                      onClick={() => approveMutation.mutate(request.id)}
                    >
                      {"Approve"}
                    </Button>
                  ) : null}
                  {canCancel(request.state) ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={cancelMutation.isPending}
                      onClick={() => cancelMutation.mutate(request.id)}
                    >
                      {"Cancel"}
                    </Button>
                  ) : null}
                  {canCreateQuoteRequest(request.state) ? (
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={createQuoteRequestMutation.isPending}
                      onClick={() => setQuoteRequestDialogOpen(true)}
                    >
                      {"Create QuoteRequest"}
                    </Button>
                  ) : null}
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{"Request header"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Requester"}</span>
                        <span className="">
                          {contactMap.get(request.requesterId) ?? `#${request.requesterId}`}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Department"}</span>
                        <span>
                          {request.departmentId
                            ? (departmentMap.get(request.departmentId) ??
                              `#${request.departmentId}`)
                            : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Needed by"}</span>
                        <span>{request.neededBy ? formatDate(request.neededBy) : "—"}</span>
                      </div>
                    </CardContent>
                  </Card>
                  {linkedApproval ? (
                    <Card>
                      <CardHeader>
                        <CardTitle className="text-base">{"Approval"}</CardTitle>
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
                    <CardTitle className="text-base">{"Lines"}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {lines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{"No lines."}</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b text-left text-muted-foreground">
                              <th className="pb-2 ">#</th>
                              <th className="pb-2 ">{"Item"}</th>
                              <th className="pb-2 ">{"Description"}</th>
                              <th className="pb-2 text-right">{"Qty"}</th>
                              <th className="pb-2 ">{"Needed by"}</th>
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
            <DialogTitle>{"Create QuoteRequest"}</DialogTitle>
            <DialogDescription>
              {"Convert this approved request into a Request for Quotation."}
            </DialogDescription>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {"The QuoteRequest will be pre-filled with the request lines."}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setQuoteRequestDialogOpen(false)}>
              {"Cancel"}
            </Button>
            <Button onClick={() => createQuoteRequestMutation.mutate(request.id)}>
              {"Create QuoteRequest"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
