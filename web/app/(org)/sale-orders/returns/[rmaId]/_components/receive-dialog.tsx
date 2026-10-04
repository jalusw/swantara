"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { JournalDateDialog } from "@/components/journal-date-dialog";
import type { Journal } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type ReceiveDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  rmaId: string;
  journals: Journal[];
  onSave: () => void;
};

export function ReceiveDialog({
  open,
  onOpenChange,
  orgId,
  rmaId,
  journals,
  onSave,
}: ReceiveDialogProps) {
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();

  const inventoryJournals = journals.filter((j) => j.type === "purchase");

  const receiveMutation = useMutation({
    mutationFn: (params: { journalId: string; date: string }) =>
      getSwantaraService().rmas.receive(Number(orgId), Number(rmaId), {
        journalId: Number(params.journalId),
        date: params.date,
      }),
    onSuccess: () => {
      toast.success(t("returnReceived"));
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  return (
    <JournalDateDialog
      title={t("receiveReturn")}
      description={t("receiveReturnDescription")}
      confirmLabel={t("actionReceive")}
      cancelLabel={tCommon("cancel")}
      open={open}
      onOpenChange={onOpenChange}
      onConfirm={(journalId, date) => receiveMutation.mutate({ journalId, date })}
      isPending={receiveMutation.isPending}
      journals={inventoryJournals}
    />
  );
}
