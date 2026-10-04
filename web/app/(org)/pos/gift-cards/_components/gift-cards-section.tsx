"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { GiftCard } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { GiftCardFormDialog } from "./gift-card-form-dialog";
import { giftCardStateTone } from "./gift-card-utils";

export function GiftCardsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Pos");
  const giftCardState = (state: string) =>
    (t as unknown as (k: string) => string)(`giftCardState_${state}`);
  const router = useRouter();
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ giftCards: GiftCard[] }, Record<string, never>>(
    "giftCards",
    (organizationId) => getSwantaraService().giftCards.list(organizationId),
  );

  const giftCards = query.data?.giftCards ?? [];

  const totalBalance = giftCards.reduce((sum, gc) => sum + gc.balance, 0);

  function handleSave(id: string) {
    setDialogOpen(false);
    void query.refetch();
    router.push(`/pos/gift-cards/${id}`);
  }

  const columns: ColumnDef<GiftCard>[] = [
    {
      accessorKey: "code",
      header: t("colCode"),
      cell: ({ row }) => (
        <a
          href={`/pos/gift-cards/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.code}
        </a>
      ),
    },
    {
      accessorKey: "initialAmount",
      header: t("colInitial"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.initialAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "balance",
      header: t("colBalance"),
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.balance, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "currencyCode",
      header: t("currency"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.currencyCode}</span>
      ),
    },
    {
      accessorKey: "expiryDate",
      header: t("colExpiry"),
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(row.original.expiryDate, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => {
        const tone = giftCardStateTone(row.original.state);
        return (
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
            {giftCardState(row.original.state)}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={t("view")}
          deleteLabel=""
          confirmTitle=""
          confirmDescription=""
          onEdit={() => router.push(`/pos/gift-cards/${row.original.id}`)}
          onDelete={() => {}}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>{t("totalCards")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{giftCards.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{t("totalBalance")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">
              {formatMoney(totalBalance, { currency: DEFAULT_CURRENCY })}
            </p>
          </CardContent>
        </Card>
      </div>
      <InteractiveEntityTable
        columns={columns}
        data={giftCards}
        getRowId={(row) => String(row.id)}
        searchKeys={["code"]}
        statusKey="state"
        statusOptions={[
          { value: "active", label: giftCardState("active") },
          { value: "used", label: giftCardState("used") },
          { value: "expired", label: giftCardState("expired") },
          { value: "cancelled", label: giftCardState("cancelled") },
        ]}
        searchPlaceholder={t("giftCardsSearchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allGiftCards")}
        ariaLabel={t("giftCardsTitle")}
        emptyTitle={t("giftCardsEmpty")}
        status={
          query.isLoading
            ? { type: "loading" }
            : query.isError
              ? {
                  type: "error",
                  message: query.error.message,
                  onRetry: () => void query.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("issueGiftCard")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <GiftCardFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
