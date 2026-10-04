"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import type { Warehouse } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function useWarehouseFormSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  return z.object({
    name: z.string().min(1, t("validationWarehouseNameRequired")),
    code: z.string(),
    line1: z.string(),
    line2: z.string(),
    city: z.string(),
    state: z.string(),
    postalCode: z.string(),
    countryCode: z.string(),
  });
}

export function WarehouseFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Warehouse | null;
  onSave: (id: string) => void;
}) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Stock");
  const tCommon = useTranslations("Common");
  const isEdit = Boolean(initial);

  const schema = useWarehouseFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          code: initial.code || "",
          line1: initial.line1 || "",
          line2: initial.line2 || "",
          city: initial.city || "",
          state: initial.state || "",
          postalCode: initial.postalCode || "",
          countryCode: initial.countryCode || "",
        }
      : {
          name: "",
          code: "",
          line1: "",
          line2: "",
          city: "",
          state: "",
          postalCode: "",
          countryCode: "",
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      organizationId: Number(orgId),
      name: values.name,
      code: values.code || null,
      line1: values.line1 || null,
      line2: values.line2 || null,
      city: values.city || null,
      state: values.state || null,
      postalCode: values.postalCode || null,
      countryCode: values.countryCode || null,
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .inventory.updateWarehouse(Number(orgId), initial.id, request)
        .then((result) => onSave(String(result.warehouse.id)))
        .catch(() => void toast.error(t("toastWarehouseFailed")));
    } else {
      return getSwantaraService()
        .inventory.createWarehouse(Number(orgId), request)
        .then((result) => onSave(String(result.warehouse.id)))
        .catch(() => void toast.error(t("toastWarehouseFailed")));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editWarehouse") : t("newWarehouse")}
      description={t("warehousesDescription")}
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
        <FormField name="code" label={t("fieldCode")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldCode")} />}
        </FormField>
        <FormField name="line1" label={t("fieldAddress1")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldAddress1")} />}
        </FormField>
        <FormField name="line2" label={t("fieldAddress2")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldAddress2")} />}
        </FormField>
        <FormField name="city" label={t("fieldCity")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldCity")} />}
        </FormField>
        <FormField name="state" label={t("fieldState")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldState")} />}
        </FormField>
        <FormField name="postalCode" label={t("fieldPostalCode")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldPostalCode")} />}
        </FormField>
        <FormField name="countryCode" label={t("fieldCountryCode")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldCountryCode")} />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
