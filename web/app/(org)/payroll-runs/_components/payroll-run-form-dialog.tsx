"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
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

function usePayrollRunSchema() {
  const t = useTranslations("Payroll");
  return z.object({
    periodStart: z.string().min(1, t("validationDateRequired")),
    periodEnd: z.string().min(1, t("validationDateRequired")),
  });
}
type PayrollRunValues = z.infer<ReturnType<typeof usePayrollRunSchema>>;

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
  const t = useTranslations("Payroll");
  const tCommon = useTranslations("Common");
  const form = useForm<PayrollRunValues>({
    resolver: zodResolver(usePayrollRunSchema()),
    defaultValues: {
      periodStart: getLocalDateString(),
      periodEnd: getLocalDateString(),
    },
  });

  function handleSubmit(values: PayrollRunValues) {
    return getSwantaraService()
      .payrollRuns.create(Number(orgId), {
        organizationId: Number(orgId),
        periodStart: values.periodStart,
        periodEnd: values.periodEnd,
      })
      .then(({ run }) => onSave(String(run.id)))
      .catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("newRun")}</DialogTitle>
          <DialogDescription>{t("newRunDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="periodStart" label={t("fieldPeriodStart")}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
            <FormField name="periodEnd" label={t("fieldPeriodEnd")}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{t("actionCreate")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
