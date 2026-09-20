"use client";

import { useMemo, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
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
  Contact,
  SupplierQuote,
  SupplierQuoteRequest,
  SupplierQuoteRequestLine,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { canCancel, canCreatePo, canSend } from "../../_components/supplier-quote-request-utils";

const quoteRequestSteps = ["draft", "sent", "done"] as const;

function quoteRequestStateIndex(state: SupplierQuoteRequest["state"]): number {
  const idx = quoteRequestSteps.indexOf(state as (typeof quoteRequestSteps)[number]);
  return idx >= 0 ? idx : 0;
}

function getQuoteRequestStatuses() {
  return {
    draft: { label: "Draft", tone: "neutral" as const },
    sent: { label: "Sent", tone: "info" as const },
    done: { label: "Done", tone: "success" as const },
    cancelled: { label: "Cancelled", tone: "danger" as const },
  };
}

export function SupplierQuoteRequestDetail({
  orgId,
  quoteRequestId,
}: {
  orgId: string;
  quoteRequestId: string;
}) {
  const [createPoOpen, setCreatePoOpen] = useState(false);

  const quoteRequestQuery = useOrgQuery<{ quoteRequest: SupplierQuoteRequest }>(
    "supplierQuoteRequest",
    quoteRequestId,
    (organizationId) =>
      getSwantaraService().supplierQuoteRequests.get(organizationId, Number(quoteRequestId)),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const quotesQuery = useOrgListQuery<{ quotes: SupplierQuote[] }, Record<string, never>>(
    "supplierQuoteRequestQuotes",
    (organizationId) =>
      getSwantaraService().supplierQuoteRequests.quotes.list(
        organizationId,
        Number(quoteRequestId),
      ),
  );
  const linesQuery = useOrgQuery<{ lines: SupplierQuoteRequestLine[] }>(
    "supplierQuoteRequestLines",
    quoteRequestId,
    (organizationId) =>
      getSwantaraService().supplierQuoteRequests.lines(organizationId, Number(quoteRequestId)),
  );

  const quoteRequest = quoteRequestQuery.data?.quoteRequest ?? null;
  const contacts = contactsQuery.data?.contacts ?? [];
  const quotes = quotesQuery.data?.quotes ?? [];
  const lines = linesQuery.data?.lines ?? [];

  const contactMap = useMemo(
    () => new Map(contacts.map((p) => [p.id, p.displayName || p.name])),
    [contacts],
  );

  if (quoteRequestQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!quoteRequest) {
    return <p className="text-sm text-muted-foreground">{"QuoteRequest not found."}</p>;
  }

  function handleSend() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.send(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success("QuoteRequest sent to supplier.");
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleCancel() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.cancel(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success("QuoteRequest cancelled.");
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleAcceptQuote(quoteId: number) {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.quotes.accept(Number(orgId), quoteRequest.id, quoteId)
      .then(() => {
        toast.success("Quote accepted.");
        void quotesQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleCreatePo() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.purchaseOrder(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success("PO Created");
        setCreatePoOpen(false);
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  const acceptedQuote = quotes.find((q) => q.state === "accepted");

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "All Quote Requests", href: "/purchases/quoteRequests" },
          { label: quoteRequest.name ?? `QuoteRequest-${quoteRequest.id}` },
        ]}
        title={quoteRequest.name ?? `QuoteRequest-${quoteRequest.id}`}
        status={<StateBadge value={quoteRequest.state} statuses={getQuoteRequestStatuses()} />}
        tabs={[
          {
            id: "overview",
            label: "Overview",
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={quoteRequestSteps.map((s) => ({ label: String(s) }))}
                  currentIndex={quoteRequestStateIndex(quoteRequest.state)}
                />
                <div className="flex flex-wrap gap-2">
                  {canSend(quoteRequest.state) ? (
                    <Button size="sm" onClick={handleSend}>
                      {"Send to supplier"}
                    </Button>
                  ) : null}
                  {canCancel(quoteRequest.state) ? (
                    <Button size="sm" variant="outline" onClick={handleCancel}>
                      {"Cancel"}
                    </Button>
                  ) : null}
                  {canCreatePo(quoteRequest.state) ? (
                    <Button size="sm" onClick={() => setCreatePoOpen(true)}>
                      {"Create purchase order"}
                    </Button>
                  ) : null}
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{"QuoteRequest header"}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Supplier"}</span>
                        <span className="">
                          {quoteRequest.supplierId
                            ? (contactMap.get(quoteRequest.supplierId) ??
                              `#${quoteRequest.supplierId}`)
                            : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Order date"}</span>
                        <span>
                          {quoteRequest.orderDate ? formatDate(quoteRequest.orderDate) : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{"Quote deadline"}</span>
                        <span>
                          {quoteRequest.quoteDeadline
                            ? formatDate(quoteRequest.quoteDeadline)
                            : "—"}
                        </span>
                      </div>
                      {quoteRequest.notes ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{"Notes"}</span>
                          <span className="max-w-[200px] truncate">{quoteRequest.notes}</span>
                        </div>
                      ) : null}
                    </CardContent>
                  </Card>
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
          {
            id: "quotes",
            label: "Quotes",
            content: (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">{"Supplier quotes"}</CardTitle>
                  <p className="text-sm text-muted-foreground">
                    {"Quotes received from suppliers for this QuoteRequest."}
                  </p>
                </CardHeader>
                <CardContent>
                  {quotes.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{"No quotes received yet."}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-muted-foreground">
                            <th className="pb-2 text-left">{"Supplier"}</th>
                            <th className="pb-2 text-left">{"Status"}</th>
                            <th className="pb-2 text-right">{"Total"}</th>
                            <th className="pb-2 text-left">{"Valid until"}</th>
                            <th className="pb-2 text-right">{"Actions"}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {quotes.map((quote) => (
                            <tr key={quote.id} className="border-b">
                              <td className="py-2">
                                {contactMap.get(quote.supplierId) ?? `#${quote.supplierId}`}
                              </td>
                              <td className="py-2">
                                <Badge variant={quote.state === "accepted" ? "default" : "outline"}>
                                  {humanizeKey(String(quote.state))}
                                </Badge>
                              </td>
                              <td className="py-2 text-right tabular-nums">
                                {formatNumber(quote.amountTotal)}
                              </td>
                              <td className="py-2 text-muted-foreground">
                                {quote.validUntil ? formatDate(quote.validUntil) : "—"}
                              </td>
                              <td className="py-2 text-right">
                                {quote.state === "submitted" ? (
                                  <Button
                                    size="sm"
                                    variant="outline"
                                    onClick={() => handleAcceptQuote(quote.id)}
                                  >
                                    {"Accept"}
                                  </Button>
                                ) : null}
                              </td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  )}
                  {acceptedQuote ? (
                    <p className="mt-3 text-xs text-muted-foreground">
                      {"Accepted quote total"}:{" "}
                      <span className="">{formatNumber(acceptedQuote.amountTotal)}</span>
                    </p>
                  ) : null}
                </CardContent>
              </Card>
            ),
          },
        ]}
      />
      <Dialog open={createPoOpen} onOpenChange={setCreatePoOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{"Create Purchase Order"}</DialogTitle>
            <DialogDescription>
              {"Convert this QuoteRequest into a purchase order using the accepted quote."}
            </DialogDescription>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            {"A purchase order will be created from the accepted quote lines."}
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreatePoOpen(false)}>
              {"Cancel"}
            </Button>
            <Button onClick={handleCreatePo}>{"Create PO"}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
