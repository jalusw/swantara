"use client";

import { useTranslations } from "next-intl";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Switch } from "@/components/switch";
import type { LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useLeaveTypeFormSchema() {
  const t = useTranslations("Leave");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    paid: z.boolean(),
    allocationDays: z.coerce.number().nullable(),
  });
}

function toDefaultValues(initial?: LeaveType | null) {
  return initial
    ? {
        name: initial.name,
        paid: initial.paid,
        allocationDays: initial.allocationDays,
      }
    : {
        name: "",
        paid: true,
        allocationDays: null,
      };
}

export function LeaveTypeFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: LeaveType | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);
  const t = useTranslations("Leave");
  const tCommon = useTranslations("Common");

  const schema = useLeaveTypeFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: toDefaultValues(initial),
  });

  useEffect(() => {
    if (open) {
      form.reset(toDefaultValues(initial));
    }
  }, [open, initial, form]);

  function handleSubmit(values: Values) {
    const request = {
      name: values.name,
      paid: values.paid,
      allocationDays: values.allocationDays,
      organizationId: Number(orgId),
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .leaveTypes.update(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    } else {
      return getSwantaraService()
        .leaveTypes.create(Number(orgId), request)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editLeaveType") : t("newLeaveType")}
      description={t("leaveTypeDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="paid" label={t("fieldPaid")}>
          {({ field }) => (
            <Switch
              checked={field.value}
              onCheckedChange={(checked) => field.onChange(Boolean(checked))}
              aria-label={t("fieldPaid")}
            />
          )}
        </FormField>
        <FormField name="allocationDays" label={t("fieldAllocationDays")}>
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
      </div>
    </EntityFormDialog>
  );
}
