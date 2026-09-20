"use client";

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
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useCommissionRuleFormSchema() {
  return z.object({
    minAmount: z.coerce.number().min(0),
    maxAmount: z.coerce.number().min(0),
    ratePct: z.coerce.number().min(0).max(100),
    fixedAmount: z.coerce.number().min(0),
  });
}
type CommissionRuleFormValues = z.infer<ReturnType<typeof useCommissionRuleFormSchema>>;

export function CommissionRuleFormDialog({
  open,
  onOpenChange,
  orgId,
  planId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  planId: string;
  onSave: () => void;
}) {
  const schema = useCommissionRuleFormSchema();

  const form = useForm<CommissionRuleFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      minAmount: 0,
      maxAmount: 0,
      ratePct: 0,
      fixedAmount: 0,
    },
  });

  function handleSubmit(values: CommissionRuleFormValues) {
    void toast.promise(
      getSwantaraService().commissionPlans.rules.create(Number(orgId), Number(planId), {
        itemCategoryId: null,
        minAmount: values.minAmount,
        maxAmount: values.maxAmount,
        ratePct: values.ratePct,
        fixedAmount: values.fixedAmount,
      }),
      {
        loading: "Saving…",
        success: () => {
          onSave();
          return "Rule added";
        },
        error: "Failed to add rule",
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Add commission rule"}</DialogTitle>
          <DialogDescription>
            {"Define a commission rule for a item category range."}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="minAmount" label={"Min amount"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
              <FormField name="maxAmount" label={"Max amount"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="ratePct" label={"Rate %"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" min={0} max={100} step="0.01" />
                )}
              </FormField>
              <FormField name="fixedAmount" label={"Fixed amount"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Save"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
