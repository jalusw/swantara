"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
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
import { canRedeem, canRefund, giftCardStateTone } from "../../_components/gift-card-utils";

export function GiftCardDetail({ orgId, giftCardId }: { orgId: string; giftCardId: string }) {
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
  const giftCardState = (state: string) =>
    (t as unknown as (k: string) => string)(`giftCardState_${state}`);
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
      toast.success(t("saved"));
    },
    onError: () => {
      toast.error(t("actionFailed"));
    },
  });

  function handleAction(action: "redeem" | "refund") {
    if (!card || actionMutation.isPending) return;
    actionMutation.mutate(action);
  }

  if (cardQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!card) {
    return <p className="text-sm text-muted-foreground">{t("giftCardNotFound")}</p>;
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
          {t("redeem")}
        </Button>
      ) : null}
      {canRefund(card.state) ? (
        <Button
          size="sm"
          variant="outline"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("refund")}
        >
          {t("refundTitle")}
        </Button>
      ) : null}
    </div>
  );

  const transactionColumns: ColumnDef<GiftCardTransaction>[] = [
    {
      accessorKey: "type",
      header: t("colType"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{String(row.original.type)}</span>
      ),
    },
    {
      accessorKey: "amount",
      header: t("amount"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.amount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "orderType",
      header: t("colOrderType"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.orderType || "—"}</span>
      ),
    },
    {
      accessorKey: "orderId",
      header: t("colOrder"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.orderId ? `#${row.original.orderId}` : "—"}
        </span>
      ),
    },
  ];

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("giftCardsTitle"), href: "/pos/gift-cards" },
        { label: card.code },
      ]}
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
          {giftCardState(card.state)}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("colInitial")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatMoney(card.initialAmount, { currency: DEFAULT_CURRENCY })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("colBalance")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatMoney(card.balance, { currency: DEFAULT_CURRENCY })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("currency")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{card.currencyCode}</p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("detailsTitle")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("customer")}</dt>
                      <dd className="text-sm">{card.contactId ? `#${card.contactId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("colExpiry")}</dt>
                      <dd className="text-sm">
                        {card.expiryDate ? formatDate(String(card.expiryDate)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("issuedFromOrder")}</dt>
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
          label: t("tabTransactions"),
          content: (
            <InteractiveEntityTable
              columns={transactionColumns}
              data={transactions}
              getRowId={(row) => String(row.id)}
              searchKeys={["type", "orderType"]}
              searchPlaceholder={t("transactionsSearchPlaceholder")}
              ariaLabel={t("tabTransactions")}
              emptyTitle={t("transactionsEmpty")}
            />
          ),
        },
      ]}
    />
  );
}
