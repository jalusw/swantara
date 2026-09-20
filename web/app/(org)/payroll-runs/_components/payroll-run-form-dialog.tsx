"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
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
import { getLocalDateString } from "@/lib/utils";

const payrollRunSchema = z.object({
  periodStart: z.string().min(1),
  periodEnd: z.string().min(1),
});
type PayrollRunValues = z.infer<typeof payrollRunSchema>;

export function PayrollRunFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: (runId: string) => void;
}) {
  const form = useForm<PayrollRunValues>({
    resolver: zodResolver(payrollRunSchema),
    defaultValues: {
      periodStart: getLocalDateString(),
      periodEnd: getLocalDateString(),
    },
  });

  function handleSubmit(values: PayrollRunValues) {
    void getSwantaraService()
      .payrollRuns.create(Number(orgId), {
        organizationId: Number(orgId),
        periodStart: values.periodStart,
        periodEnd: values.periodEnd,
      })
      .then(({ run }) => onSave(String(run.id)));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"New payroll run"}</DialogTitle>
          <DialogDescription>{"Create a payroll run for a period."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="periodStart" label={"Period start"}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
            <FormField name="periodEnd" label={"Period end"}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Create"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
