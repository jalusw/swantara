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
import { Textarea } from "@/components/textarea";
import { getSwantaraService } from "@/lib/services/swantara";

function useLostReasonFormSchema() {
  return z.object({
    lostReason: z.string().min(1, "Lost reason is required."),
  });
}
type LostReasonValues = z.infer<ReturnType<typeof useLostReasonFormSchema>>;

export function LostReasonDialog({
  open,
  onOpenChange,
  orgId,
  opportunityId,
  onDone,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  opportunityId: string;
  onDone: () => void;
}) {
  const queryClient = useQueryClient();
  const schema = useLostReasonFormSchema();
  const form = useForm<LostReasonValues>({
    resolver: zodResolver(schema),
    defaultValues: { lostReason: "" },
  });

  const loseMutation = useMutation({
    mutationFn: (lostReason: string) =>
      getSwantaraService().crmOpportunities.lose(Number(orgId), Number(opportunityId), {
        lostReason,
      }),
    onSuccess: () => {
      toast.success("Mark lost");
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onDone();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  function handleSubmit(values: LostReasonValues) {
    loseMutation.mutate(values.lostReason);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Mark as lost"}</DialogTitle>
          <DialogDescription>{"Explain why this deal was lost."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="lostReason" label={"Lost reason"}>
            {({ field, id }) => <Textarea {...field} id={id} placeholder={"e.g. Price too high"} />}
          </FormField>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {"Mark as lost"}
            </Button>
            <SubmitButton>{"Mark lost"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
