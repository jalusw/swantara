"use client";

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
import { UI_COPY } from "@/lib/constants/ui-copy";
import type { Journal } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

export type JournalDateDialogProps = {
  title: string;
  description?: string;
  confirmLabel: string;
  cancelLabel?: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (journalId: string, date: string) => void;
  isPending?: boolean;
  journals: Journal[];
};

export function JournalDateDialog({
  title,
  description,
  confirmLabel,
  cancelLabel = UI_COPY.cancel,
  open,
  onOpenChange,
  onConfirm,
  isPending,
  journals,
}: JournalDateDialogProps) {
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent data-slot="journal-date-dialog" className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{UI_COPY.journal}</span>
            <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
              <SelectTrigger aria-label={UI_COPY.journal}>
                <SelectValue placeholder={UI_COPY.selectJournal} />
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
          <div className="flex flex-col gap-1">
            <span className="text-sm">{UI_COPY.date}</span>
            <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {cancelLabel}
          </Button>
          <Button onClick={() => onConfirm(journalId, date)} disabled={!journalId || isPending}>
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
