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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useDeferralSchema() {
  const t = useTranslations("Deferrals");
  return z.object({
    type: z.string().min(1, t("validationTypeRequired")),
    sourceType: z.string().min(1, t("validationSourceTypeRequired")),
    sourceId: z.coerce.number().min(1, t("validationSourceIdRequired")),
    totalAmount: z.coerce.number().positive(t("validationTotalAmountPositive")),
    method: z.string().min(1, t("validationMethodRequired")),
    periods: z.coerce.number().min(1, t("validationPeriodsRequired")),
    dateStart: z.string().min(1, t("validationStartDateRequired")),
  });
}
type DeferralValues = z.infer<ReturnType<typeof useDeferralSchema>>;

export function DeferralFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: (id: string) => void;
}) {
  const t = useTranslations("Deferrals");
  const schema = useDeferralSchema();

  const form = useForm<DeferralValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: "",
      sourceType: "",
      sourceId: 0,
      totalAmount: 0,
      method: "",
      periods: 1,
      dateStart: "",
    },
  });

  async function handleSubmit(values: DeferralValues) {
    const request = getSwantaraService().deferrals.create(Number(orgId), {
      type: values.type as "deferred_revenue" | "deferred_expense" | "prepaid",
      sourceType: values.sourceType,
      sourceId: values.sourceId,
      contactId: null,
      itemId: null,
      totalAmount: values.totalAmount,
      balanceSheetAccountId: 1,
      plAccountId: 1,
      dimensionId: null,
      method: values.method as "linear" | "manual" | "milestone",
      dateStart: new Date(values.dateStart),
      dateEnd: null,
      periods: values.periods,
      lines: [],
    });
    toast.promise(request, {
      loading: t("saving"),
      success: (result) => {
        onSave(String(result.schedule.id));
        return t("toastDeferralCreated");
      },
      error: t("toastDeferralFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("createDeferral")}</DialogTitle>
          <DialogDescription>{t("createDeferralDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="type" label={t("fieldType")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldType")}>
                      <SelectValue placeholder={t("selectType")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="deferred_revenue">
                        {(t as unknown as (k: string) => string)("deferralType_deferred_revenue")}
                      </SelectItem>
                      <SelectItem value="deferred_expense">
                        {(t as unknown as (k: string) => string)("deferralType_deferred_expense")}
                      </SelectItem>
                      <SelectItem value="prepaid">
                        {(t as unknown as (k: string) => string)("deferralType_prepaid")}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="method" label={t("fieldMethod")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldMethod")}>
                      <SelectValue placeholder={t("selectMethod")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="linear">
                        {(t as unknown as (k: string) => string)("deferMethod_linear")}
                      </SelectItem>
                      <SelectItem value="manual">
                        {(t as unknown as (k: string) => string)("deferMethod_manual")}
                      </SelectItem>
                      <SelectItem value="milestone">
                        {(t as unknown as (k: string) => string)("deferMethod_milestone")}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="sourceType" label={t("fieldSourceType")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldSourceType")} />}
              </FormField>
              <FormField name="sourceId" label={t("fieldSourceId")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="totalAmount" label={t("fieldTotalAmount")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
              <FormField name="periods" label={t("fieldPeriods")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
              </FormField>
            </div>
            <FormField name="dateStart" label={t("fieldStartDate")}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {t("cancel")}
            </Button>
            <SubmitButton>{t("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
