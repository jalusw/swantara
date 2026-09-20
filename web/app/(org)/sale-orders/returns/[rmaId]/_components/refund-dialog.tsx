"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import type { Journal } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

type RefundDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  rmaId: string;
  journals: Journal[];
  onSave: () => void;
};

function useRefundFormSchema() {
  return z.object({
    journalId: z.string().min(1, "Select a journal."),
    date: z.string().min(1, "Date is required."),
    reference: z.string(),
  });
}

export function RefundDialog({
  open,
  onOpenChange,
  orgId,
  rmaId,
  journals,
  onSave,
}: RefundDialogProps) {
  const queryClient = useQueryClient();

  const schema = useRefundFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      journalId: "",
      date: getLocalDateString(),
      reference: "",
    },
  });

  const accountJournals = journals.filter((j) => j.type === "general");

  const refundMutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().rmas.refund(Number(orgId), Number(rmaId), {
        journalId: Number(values.journalId),
        date: values.date,
        reference: values.reference,
      }),
    onSuccess: () => {
      toast.success("Return refunded successfully");
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  function handleSubmit(values: Values) {
    refundMutation.mutate(values);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{"Refund Return"}</DialogTitle>
          <DialogDescription>{"Create a credit note for the returned goods."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="journalId" label={"Journal"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Journal"}>
                  <SelectValue placeholder={"Select a journal"} />
                </SelectTrigger>
                <SelectContent>
                  {accountJournals.map((j) => (
                    <SelectItem key={j.id} value={String(j.id)}>
                      {j.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="date" label={"Refund date"}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
          <FormField name="reference" label={"Reference"}>
            {({ field, id }) => <Input {...field} id={id} />}
          </FormField>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton loading={refundMutation.isPending}>{"Refund"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
