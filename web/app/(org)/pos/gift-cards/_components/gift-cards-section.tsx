"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useRouter } from "next/navigation";
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
import { humanizeKey } from "@/lib/utils/case";
import { GiftCardFormDialog } from "./gift-card-form-dialog";
import { giftCardStateTone } from "./gift-card-utils";

export function GiftCardsSection({ orgId }: { orgId: string }) {
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
      header: "Code",
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
      header: "Initial",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.initialAmount, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "balance",
      header: "Balance",
      cell: ({ row }) => (
        <span className="text-sm tabular-nums">
          {formatMoney(row.original.balance, { currency: DEFAULT_CURRENCY })}
        </span>
      ),
    },
    {
      accessorKey: "currencyCode",
      header: "Currency",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">{row.original.currencyCode}</span>
      ),
    },
    {
      accessorKey: "expiryDate",
      header: "Expiry",
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {formatDate(row.original.expiryDate, { nullFallback: "—" })}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
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
            {humanizeKey(String(row.original.state))}
          </Badge>
        );
      },
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"View"}
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
            <CardTitle>{"Total Cards"}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-2xl font-bold tabular-nums">{giftCards.length}</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>{"Total Balance"}</CardTitle>
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
          { value: "active", label: "Active" },
          { value: "used", label: "Used" },
          { value: "expired", label: "Expired" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search gift cards…"}
        filterLabel={"State"}
        allLabel={"All gift cards"}
        ariaLabel={"Gift cards"}
        emptyTitle={"No gift cards yet"}
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
            <span>{"Issue gift card"}</span>
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
