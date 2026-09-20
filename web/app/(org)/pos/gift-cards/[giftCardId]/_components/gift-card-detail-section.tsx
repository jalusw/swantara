"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { GiftCard, GiftCardTransaction } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { canRedeem, canRefund, giftCardStateTone } from "../../_components/gift-card-utils";

export function GiftCardDetail({ orgId, giftCardId }: { orgId: string; giftCardId: string }) {
  const queryClient = useQueryClient();

  const cardQuery = useOrgQuery<{ giftCard: GiftCard }>("giftCard", giftCardId, (organizationId) =>
    getSwantaraService().giftCards.get(organizationId, Number(giftCardId)),
  );

  const transactionsQuery = useOrgListQuery<
    { giftCardTransactions: GiftCardTransaction[] },
    Record<string, never>
  >("giftCardTransactions", (organizationId) =>
    getSwantaraService().giftCards.transactions(organizationId, Number(giftCardId)),
  );

  const card = cardQuery.data?.giftCard;
  const transactions = transactionsQuery.data?.giftCardTransactions ?? [];

  const actionMutation = useMutation({
    mutationFn: (action: "redeem" | "refund") =>
      getSwantaraService().giftCards[action](Number(orgId), Number(giftCardId), {
        amount: card?.balance ?? 0,
        journalId: 1,
        revenueAccountId: 1,
        refundAccountId: 1,
        liabilityAccountId: 1,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["giftCard"] });
      void queryClient.invalidateQueries({ queryKey: ["giftCardTransactions"] });
      void queryClient.invalidateQueries({ queryKey: ["giftCards"] });
      toast.success("Saved.");
    },
    onError: () => {
      toast.error("Action failed");
    },
  });

  function handleAction(action: "redeem" | "refund") {
    if (!card || actionMutation.isPending) return;
    actionMutation.mutate(action);
  }

  if (cardQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!card) {
    return <p className="text-sm text-muted-foreground">{"Gift card not found."}</p>;
  }

  const tone = giftCardStateTone(card.state);

  const stateActions = (
    <div className="flex items-center gap-2">
      {canRedeem(card.state) ? (
        <Button
          size="sm"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("redeem")}
        >
          {"Redeem"}
        </Button>
      ) : null}
      {canRefund(card.state) ? (
        <Button
          size="sm"
          variant="outline"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("refund")}
        >
          {"Refund"}
        </Button>
      ) : null}
    </div>
  );

  const transactionColumns: ColumnDef<GiftCardTransaction>[] = [
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{String(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: "Amount",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.amount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "orderType",
      header: "Order type",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.orderType || "—"}</span>
      ),
    },
    {
      accessorKey: "orderId",
      header: "Order",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.orderId ? `#${row.original.orderId}` : "—"}
        </span>
      ),
    },
  ];

  return (
    <RecordLayout
      breadcrumbItems={[{ label: "Gift cards", href: "/pos/gift-cards" }, { label: card.code }]}
      title={card.code}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "danger"
                  ? "border-destructive text-destructive"
                  : tone === "info"
                    ? "border-info text-info"
                    : ""
          }
        >
          {humanizeKey(String(card.state))}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Initial amount"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatMoney(card.initialAmount, { currency: DEFAULT_CURRENCY })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Balance"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatMoney(card.balance, { currency: DEFAULT_CURRENCY })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Currency"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{card.currencyCode}</p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{"Details"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Customer"}</dt>
                      <dd className="text-sm">{card.contactId ? `#${card.contactId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Expiry date"}</dt>
                      <dd className="text-sm">
                        {card.expiryDate ? formatDate(String(card.expiryDate)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Issued From Order"}</dt>
                      <dd className="text-sm">
                        {card.issuedFromOrderId ? `#${card.issuedFromOrderId}` : "—"}
                      </dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "transactions",
          label: "Transactions",
          content: (
            <InteractiveEntityTable
              columns={transactionColumns}
              data={transactions}
              getRowId={(row) => String(row.id)}
              searchKeys={["type", "orderType"]}
              searchPlaceholder={"Search Transactions"}
              ariaLabel={"Transactions"}
              emptyTitle={"No transactions."}
            />
          ),
        },
      ]}
    />
  );
}
