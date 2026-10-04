"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Switch } from "@/components/switch";
import { getSwantaraService } from "@/lib/services/swantara";

const DEFAULT_VALUES = {
  name: "",
  displayName: "",
  email: "",
  phone: "",
  active: true,
};

function useCustomerFormSchema() {
  const t = useTranslations("Sales");
  return z.object({
    name: z.string().trim().min(1, t("validationNameRequired")),
    displayName: z.string().trim(),
    email: z
      .string()
      .trim()
      .refine(
        (value) => value === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        t("validationEmailInvalid"),
      ),
    phone: z.string().trim(),
    active: z.boolean(),
  });
}

export function CustomerFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: () => void;
}) {
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const [isPending, setIsPending] = useState(false);

  const schema = useCustomerFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: DEFAULT_VALUES,
  });

  useEffect(() => {
    if (open) {
      form.reset(DEFAULT_VALUES);
    }
  }, [open, form]);

  function handleSubmit(values: Values) {
    setIsPending(true);
    void getSwantaraService()
      .contacts.create(Number(orgId), {
        organizationId: Number(orgId),
        name: values.name,
        displayName: values.displayName || null,
        isOrganization: true,
        parentId: null,
        email: values.email || null,
        phone: values.phone || null,
        mobile: null,
        website: null,
        taxId: null,
        industry: null,
        currencyCode: null,
        lang: "en",
        active: values.active,
        addresses: [],
        bankAccounts: [],
        customer: {
          customerPaymentTermId: null,
          creditLimit: null,
          receivableAccountId: null,
          active: values.active,
        },
        supplier: null,
      })
      .then(() => {
        toast.success(t("customerSaved"));
        onSave();
      })
      .catch(() => toast.error(t("saveFailed")))
      .finally(() => setIsPending(false));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newCustomer")}
      description={t("newCustomerDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
    >
      <div className="flex flex-col gap-6">
        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{t("sectionIdentity")}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={`${t("fieldName")} *`}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  placeholder={t("fieldName")}
                  autoComplete="organization"
                  required
                  aria-required="true"
                />
              )}
            </FormField>
            <FormField name="displayName" label={t("fieldDisplayName")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  placeholder={t("fieldDisplayName")}
                  autoComplete="organization"
                />
              )}
            </FormField>
          </div>
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{t("sectionContact")}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="email" label={t("fieldEmail")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="email"
                  inputMode="email"
                  autoComplete="email"
                  placeholder={t("fieldEmail")}
                />
              )}
            </FormField>
            <FormField name="phone" label={t("fieldPhone")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="tel"
                  inputMode="tel"
                  autoComplete="tel"
                  placeholder={t("fieldPhone")}
                />
              )}
            </FormField>
          </div>
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{t("sectionStatus")}</legend>
          <FormField
            name="active"
            label={t("statusActive")}
            description={t("inactiveCustomerHint")}
          >
            {({ field }) => (
              <Switch
                checked={field.value}
                onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                aria-label={t("statusActive")}
              />
            )}
          </FormField>
        </fieldset>
      </div>
    </EntityFormDialog>
  );
}
