"use client";

import { Plus, X } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import type { PosPaymentRequest } from "@/lib/services/swantara";
import { formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";

const PAYMENT_METHODS = ["cash", "card", "transfer", "other"];

export function PosPayment({
  open,
  onOpenChange,
  total,
  onSubmit,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  total: number;
  onSubmit: (payments: PosPaymentRequest[]) => void;
}) {
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
  const paymentMethod = (method: string) =>
    (t as unknown as (k: string) => string)(`paymentMethod_${method}`);
  const [payments, setPayments] = useState<PosPaymentRequest[]>([
    { method: "cash", amount: total },
  ]);

  const paidTotal = payments.reduce((sum, p) => sum + p.amount, 0);
  const remaining = total - paidTotal;
  const isBalanced = Math.abs(remaining) < 0.01;

  function addPayment() {
    setPayments((prev) => [...prev, { method: "cash", amount: 0 }]);
  }

  function removePayment(index: number) {
    setPayments((prev) => prev.filter((_, i) => i !== index));
  }

  function updatePayment(index: number, updates: Partial<PosPaymentRequest>) {
    setPayments((prev) => prev.map((p, i) => (i === index ? { ...p, ...updates } : p)));
  }

  function handleSubmit() {
    if (!isBalanced) return;
    onSubmit(payments.filter((p) => p.amount > 0));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("paymentTitle")}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between rounded-lg bg-muted p-3">
            <span className="text-sm">{t("totalDue")}</span>
            <span className="text-lg font-bold tabular-nums">{formatMoney(total)}</span>
          </div>

          <div className="flex flex-col gap-3">
            {payments.map((payment, index) => (
              <div key={`pay-${payment.method}-${index}`} className="flex items-end gap-2">
                <div className="flex-1 flex flex-col gap-1">
                  {index === 0 ? (
                    <span className="text-xs text-muted-foreground">{t("method")}</span>
                  ) : null}
                  <Select
                    value={payment.method}
                    onValueChange={(v) => updatePayment(index, { method: v ?? "cash" })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {PAYMENT_METHODS.map((m) => (
                        <SelectItem key={m} value={m}>
                          {paymentMethod(m)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="flex-1 flex flex-col gap-1">
                  {index === 0 ? (
                    <span className="text-xs text-muted-foreground">{t("amount")}</span>
                  ) : null}
                  <Input
                    type="number"
                    step="any"
                    min="0"
                    value={payment.amount}
                    onChange={(e) =>
                      updatePayment(index, {
                        amount: Number(e.target.value) || 0,
                      })
                    }
                  />
                </div>
                {payments.length > 1 ? (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0 text-destructive"
                    onClick={() => removePayment(index)}
                  >
                    <X className="h-4 w-4" />
                  </Button>
                ) : (
                  <div className="h-9 w-9 shrink-0" />
                )}
              </div>
            ))}
          </div>

          <Button variant="outline" size="sm" className="self-start" onClick={addPayment}>
            <Plus className="mr-1 h-4 w-4" />
            {t("addPayment")}
          </Button>

          <div className="flex items-center justify-between rounded-lg border p-3">
            <span className="text-sm">{t("remaining")}</span>
            <span
              className={cn(
                "text-base font-bold tabular-nums",
                isBalanced ? "text-success" : remaining > 0 ? "text-destructive" : "text-success",
              )}
            >
              {formatMoney(remaining)}
            </span>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!isBalanced}>
            {tCommon("confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
