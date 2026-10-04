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
import { Switch } from "@/components/switch";
import type { ReorderRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function useReorderRuleFormSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  return z.object({
    itemId: z.coerce.number().min(1, t("validationItemRequired")),
    minQty: z.coerce.number().min(0, t("validationNonNegative")),
    maxQty: z.coerce.number().min(0, t("validationNonNegative")),
    qtyMultiple: z.coerce.number().min(1, t("validationMinOne")),
    leadTimeDays: z.coerce.number().nullable(),
    active: z.boolean(),
  });
}

export function ReorderRuleFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: ReorderRule | null;
  onSave: () => void;
}) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  const tCommon = useTranslations("Common");
  const isEdit = Boolean(initial);

  const schema = useReorderRuleFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          itemId: initial.itemId,
          minQty: initial.minQty,
          maxQty: initial.maxQty,
          qtyMultiple: initial.qtyMultiple,
          leadTimeDays: initial.leadTimeDays,
          active: initial.active,
        }
      : {
          itemId: 0,
          minQty: 0,
          maxQty: 0,
          qtyMultiple: 1,
          leadTimeDays: null,
          active: true,
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      itemId: values.itemId,
      warehouseId: initial?.warehouseId ?? null,
      locationId: initial?.locationId ?? null,
      minQty: values.minQty,
      maxQty: values.maxQty,
      qtyMultiple: values.qtyMultiple,
      leadTimeDays: values.leadTimeDays,
      active: values.active,
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .inventory.updateReorderRule(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => void toast.error(t("toastFailed")));
    } else {
      return getSwantaraService()
        .inventory.createReorderRule(Number(orgId), request)
        .then(() => onSave())
        .catch(() => void toast.error(t("toastFailed")));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? t("editReorderRule") : t("newReorderRule")}</DialogTitle>
          <DialogDescription>{t("reorderRulesDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="itemId" label={t("fieldItem")}>
              {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
            </FormField>
            <div className="grid gap-4 sm:grid-cols-3">
              <FormField name="minQty" label={t("fieldMinQty")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
              </FormField>
              <FormField name="maxQty" label={t("fieldMaxQty")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
              </FormField>
              <FormField name="qtyMultiple" label={t("fieldQtyMultiple")}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
              </FormField>
            </div>
            <FormField name="leadTimeDays" label={t("fieldLeadTime")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  min="0"
                  value={field.value ?? ""}
                  onChange={(event) => {
                    const value = event.target.value;
                    field.onChange(value === "" ? null : Number(value));
                  }}
                />
              )}
            </FormField>
            <FormField name="active" label={t("active")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("active")}
                />
              )}
            </FormField>
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
