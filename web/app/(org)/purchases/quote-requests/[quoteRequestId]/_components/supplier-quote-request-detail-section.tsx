"use client";

import { useTranslations } from "next-intl";
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

function getQuoteRequestStatuses(label?: (state: string) => string) {
  const text = (state: string, fallback: string) => (label ? label(state) : fallback);
  return {
    draft: { label: text("draft", "Draft"), tone: "neutral" as const },
    sent: { label: text("sent", "Terkirim"), tone: "info" as const },
    done: { label: text("done", "Selesai"), tone: "success" as const },
    cancelled: { label: text("cancelled", "Dibatalkan"), tone: "danger" as const },
  };
}

export function SupplierQuoteRequestDetail({
  orgId,
  quoteRequestId,
}: {
  orgId: string;
  quoteRequestId: string;
}) {
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
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
    return <p className="text-sm text-muted-foreground">{t("quoteRequestNotFound")}</p>;
  }

  function quoteStateLabel(state: string): string {
    try {
      return (t as unknown as (k: string) => string)(`quoteRequestState.${state}`);
    } catch {
      return state;
    }
  }

  function quoteStepLabel(step: string): string {
    try {
      return (t as unknown as (k: string) => string)(`quoteRequestStep.${step}`);
    } catch {
      return String(step);
    }
  }

  function handleSend() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.send(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success(t("quoteRequestSent"));
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleCancel() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.cancel(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success(t("quoteRequestCancelled"));
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleAcceptQuote(quoteId: number) {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.quotes.accept(Number(orgId), quoteRequest.id, quoteId)
      .then(() => {
        toast.success(t("quoteAccepted"));
        void quotesQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  function handleCreatePo() {
    if (!quoteRequest) return;
    void getSwantaraService()
      .supplierQuoteRequests.purchaseOrder(Number(orgId), quoteRequest.id)
      .then(() => {
        toast.success(t("purchaseOrderCreated"));
        setCreatePoOpen(false);
        void quoteRequestQuery.refetch();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  const acceptedQuote = quotes.find((q) => q.state === "accepted");

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: t("allQuoteRequests"), href: "/purchases/quoteRequests" },
          { label: quoteRequest.name ?? `QuoteRequest-${quoteRequest.id}` },
        ]}
        title={quoteRequest.name ?? `QuoteRequest-${quoteRequest.id}`}
        status={
          <StateBadge
            value={quoteRequest.state}
            statuses={getQuoteRequestStatuses(quoteStateLabel)}
          />
        }
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: (
              <div className="flex flex-col gap-4">
                <WorkflowSteps
                  steps={quoteRequestSteps.map((s) => ({ label: quoteStepLabel(s) }))}
                  currentIndex={quoteRequestStateIndex(quoteRequest.state)}
                />
                <div className="flex flex-wrap gap-2">
                  {canSend(quoteRequest.state) ? (
                    <Button size="sm" onClick={handleSend}>
                      {t("sendToSupplier")}
                    </Button>
                  ) : null}
                  {canCancel(quoteRequest.state) ? (
                    <Button size="sm" variant="outline" onClick={handleCancel}>
                      {tCommon("cancel")}
                    </Button>
                  ) : null}
                  {canCreatePo(quoteRequest.state) ? (
                    <Button size="sm" onClick={() => setCreatePoOpen(true)}>
                      {t("createPurchaseOrder")}
                    </Button>
                  ) : null}
                </div>
                <div className="grid gap-4 sm:grid-cols-2">
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">{t("quoteRequestHeader")}</CardTitle>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableSupplier")}</span>
                        <span className="">
                          {quoteRequest.supplierId
                            ? (contactMap.get(quoteRequest.supplierId) ??
                              `#${quoteRequest.supplierId}`)
                            : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableOrderDate")}</span>
                        <span>
                          {quoteRequest.orderDate ? formatDate(quoteRequest.orderDate) : "—"}
                        </span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-muted-foreground">{t("tableQuoteDeadline")}</span>
                        <span>
                          {quoteRequest.quoteDeadline
                            ? formatDate(quoteRequest.quoteDeadline)
                            : "—"}
                        </span>
                      </div>
                      {quoteRequest.notes ? (
                        <div className="flex justify-between">
                          <span className="text-muted-foreground">{t("fieldNote")}</span>
                          <span className="max-w-[200px] truncate">{quoteRequest.notes}</span>
                        </div>
                      ) : null}
                    </CardContent>
                  </Card>
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
          {
            id: "quotes",
            label: t("tabQuotes"),
            content: (
              <Card>
                <CardHeader>
                  <CardTitle className="text-base">{t("supplierQuotes")}</CardTitle>
                  <p className="text-sm text-muted-foreground">{t("supplierQuotesDescription")}</p>
                </CardHeader>
                <CardContent>
                  {quotes.length === 0 ? (
                    <p className="text-sm text-muted-foreground">{t("noQuotes")}</p>
                  ) : (
                    <div className="overflow-x-auto">
                      <table className="w-full text-sm">
                        <thead>
                          <tr className="border-b text-muted-foreground">
                            <th className="pb-2 text-left">{t("tableSupplier")}</th>
                            <th className="pb-2 text-left">{t("tableStatus")}</th>
                            <th className="pb-2 text-right">{t("tableTotal")}</th>
                            <th className="pb-2 text-left">{t("validUntil")}</th>
                            <th className="pb-2 text-right">{t("tableActions")}</th>
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
                                    {t("acceptQuote")}
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
                      {t("acceptedQuoteTotal")}:{" "}
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
            <DialogTitle>{t("createPurchaseOrder")}</DialogTitle>
            <DialogDescription>{t("createPoDescription")}</DialogDescription>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">{t("createPoHint")}</p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreatePoOpen(false)}>
              {tCommon("cancel")}
            </Button>
            <Button onClick={handleCreatePo}>{t("createPoSubmit")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
