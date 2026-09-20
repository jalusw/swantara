"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
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
  const queryClient = useQueryClient();

  const inventoryJournals = journals.filter((j) => j.type === "purchase");

  const receiveMutation = useMutation({
    mutationFn: (params: { journalId: string; date: string }) =>
      getSwantaraService().rmas.receive(Number(orgId), Number(rmaId), {
        journalId: Number(params.journalId),
        date: params.date,
      }),
    onSuccess: () => {
      toast.success("Return received successfully");
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  return (
    <JournalDateDialog
      title={"Receive Return"}
      description={"Receive the returned goods into inventory."}
      confirmLabel={"Receive"}
      cancelLabel={"Cancel"}
      open={open}
      onOpenChange={onOpenChange}
      onConfirm={(journalId, date) => receiveMutation.mutate({ journalId, date })}
      isPending={receiveMutation.isPending}
      journals={inventoryJournals}
    />
  );
}
