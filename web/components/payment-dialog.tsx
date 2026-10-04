"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import type { Journal } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

type PaymentDialogProps = {
  title: string;
  description?: string;
  confirmLabel: string;
  cancelLabel?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (values: { journalId: string; date: string; amount?: string }) => void;
  isPending?: boolean;
  journals: Journal[];
  showAmount?: boolean;
};

export function PaymentDialog({
  title,
  description,
  confirmLabel,
  cancelLabel,
  open,
  onOpenChange,
  onConfirm,
  isPending,
  journals,
  showAmount = false,
}: PaymentDialogProps) {
  const tCommon = useTranslations("Common");
  const cancel = cancelLabel ?? tCommon("cancel");
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());
  const [amount, setAmount] = useState("");

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-slot="payment-dialog" className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{tCommon("journal")}</span>
            <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
              <SelectTrigger aria-label={tCommon("journal")}>
                <SelectValue placeholder={tCommon("selectJournal")} />
              </SelectTrigger>
              <SelectContent>
                {journals.map((j) => (
                  <SelectItem key={j.id} value={String(j.id)}>
                    {j.name} — {j.type}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          {showAmount ? (
            <div className="flex flex-col gap-1">
              <span className="text-sm">{tCommon("amount")}</span>
              <Input
                type="number"
                min="0"
                step="any"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
              />
            </div>
          ) : null}
          <div className="flex flex-col gap-1">
            <span className="text-sm">{tCommon("date")}</span>
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {cancel}
          </Button>
          <Button
            onClick={() =>
              onConfirm(showAmount ? { journalId, date, amount } : { journalId, date })
            }
            disabled={!journalId || (showAmount && !amount) || isPending}
          >
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
