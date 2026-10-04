"use client";

import { useTranslations } from "next-intl";
import { useCallback, useState } from "react";
import { toast } from "sonner";
import { Button } from "@/components/button";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type {
  Item,
  PosOrderLineRequest,
  PosPaymentRequest,
  PosSession,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";
import { canPlaceOrder } from "../../../_components/pos-utils";
import { PosCart } from "./pos-cart";
import { PosPayment } from "./pos-payment";
import { PosProductGrid } from "./pos-product-grid";
import { PosReceipt } from "./pos-receipt";

export type CartLine = {
  key: string;
  itemId: number;
  productName: string;
  qty: number;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
};

export function PosRegister({ orgId, sessionId }: { orgId: string; sessionId: string }) {
  const t = useTranslations("Pos");
  const [cart, setCart] = useState<CartLine[]>([]);
  const [paymentOpen, setPaymentOpen] = useState(false);
  const [receipt, setReceipt] = useState<{
    orderId: number;
    orderName: string | null;
    amountTotal: number;
    payments: { method: string; amount: number }[];
  } | null>(null);

  const sessionQuery = useOrgQuery<{ session: PosSession }>(
    "posSession",
    sessionId,
    (organizationId) => getSwantaraService().posSessions.get(organizationId, Number(sessionId)),
  );
  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const session = sessionQuery.data?.session ?? null;
  const products = productsQuery.data?.products ?? [];

  const subtotal = cart.reduce(
    (sum, line) => sum + line.unitPrice * line.qty * (1 - line.discountPct / 100),
    0,
  );

  const addItem = useCallback((item: Item) => {
    setCart((prev) => {
      const existing = prev.find((l) => l.itemId === item.id);
      if (existing) {
        return prev.map((l) => (l.itemId === item.id ? { ...l, qty: l.qty + 1 } : l));
      }
      return [
        ...prev,
        {
          key: `line-${item.id}-${Date.now()}`,
          itemId: item.id,
          productName: item.name,
          qty: 1,
          unitPrice: item.listPrice,
          discountPct: 0,
          taxIds: [],
        },
      ];
    });
  }, []);

  const updateLine = useCallback((key: string, updates: Partial<CartLine>) => {
    setCart((prev) => prev.map((l) => (l.key === key ? { ...l, ...updates } : l)));
  }, []);

  const removeLine = useCallback((key: string) => {
    setCart((prev) => prev.filter((l) => l.key !== key));
  }, []);

  const clearCart = useCallback(() => {
    setCart([]);
  }, []);

  function handlePayment(payments: PosPaymentRequest[]) {
    if (!session || cart.length === 0) return;

    const lines: PosOrderLineRequest[] = cart.map((line) => ({
      itemId: line.itemId,
      qty: line.qty,
      discountPct: line.discountPct,
      taxIds: line.taxIds,
    }));

    return getSwantaraService()
      .posOrders.create(Number(orgId), {
        sessionId: session.id,
        contactId: null,
        lines,
        payments,
      })
      .then(({ order }) => {
        toast.success(t("orderPlaced"));
        setPaymentOpen(false);
        setReceipt({
          orderId: order.id,
          orderName: order.name,
          amountTotal: order.amountTotal,
          payments,
        });
        clearCart();
        void sessionQuery.refetch();
      })
      .catch(() => {});
  }

  if (receipt) {
    return (
      <PosReceipt
        orderId={receipt.orderId}
        orderName={receipt.orderName}
        amountTotal={receipt.amountTotal}
        payments={receipt.payments}
        onNewOrder={() => setReceipt(null)}
      />
    );
  }

  return (
    <div className="flex h-[calc(100dvh-4rem)] flex-col">
      <div className="flex items-center justify-between border-b px-4 py-2">
        <div className="flex items-center gap-3">
          <h1 className="text-sm">{t("registerTitle")}</h1>
          {session ? (
            <span className="text-xs text-muted-foreground">
              {t("session")}: {t("sessionFallback", { id: session.id })}
            </span>
          ) : null}
        </div>
        <div className="flex items-center gap-3">
          <span className="text-sm text-muted-foreground">{t("items")}:</span>
          <span className="text-sm tabular-nums">{cart.reduce((sum, l) => sum + l.qty, 0)}</span>
          <span className="text-sm text-muted-foreground ml-2">{t("colTotal")}:</span>
          <span className="text-base font-bold tabular-nums">{formatMoney(subtotal)}</span>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        <div className="flex-1 overflow-y-auto p-3">
          <PosProductGrid
            products={products}
            onAddItem={addItem}
            searchPlaceholder={t("searchProducts")}
          />
        </div>

        <div className="flex w-80 flex-col border-l">
          <PosCart
            lines={cart}
            subtotal={subtotal}
            onUpdateLine={updateLine}
            onRemoveLine={removeLine}
            onClear={clearCart}
          />
          <div className="border-t p-3">
            <Button
              className="w-full"
              size="lg"
              disabled={cart.length === 0 || (session != null && !canPlaceOrder(session.state))}
              onClick={() => setPaymentOpen(true)}
            >
              {t("pay")} — {formatMoney(subtotal)}
            </Button>
          </div>
        </div>
      </div>

      {paymentOpen ? (
        <PosPayment
          open={paymentOpen}
          onOpenChange={setPaymentOpen}
          total={subtotal}
          onSubmit={handlePayment}
        />
      ) : null}
    </div>
  );
}
