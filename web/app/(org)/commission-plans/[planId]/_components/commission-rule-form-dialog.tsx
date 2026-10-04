"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Commissions");
  const tCommon = useTranslations("Common");
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

  async function handleSubmit(values: CommissionRuleFormValues) {
    const request = getSwantaraService().commissionPlans.rules.create(
      Number(orgId),
      Number(planId),
      {
        itemCategoryId: null,
        minAmount: values.minAmount,
        maxAmount: values.maxAmount,
        ratePct: values.ratePct,
        fixedAmount: values.fixedAmount,
      },
    );
    toast.promise(request, {
      loading: t("saving"),
      success: () => {
        onSave();
        return t("ruleCreated");
      },
      error: t("createRuleFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("newRule")}</DialogTitle>
          <DialogDescription>{t("newRuleDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="minAmount" label={t("fieldMinAmount")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
              <FormField name="maxAmount" label={t("fieldMaxAmount")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="ratePct" label={t("fieldRatePct")}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" min={0} max={100} step="0.01" />
                )}
              </FormField>
              <FormField name="fixedAmount" label={t("fieldFixedAmount")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
