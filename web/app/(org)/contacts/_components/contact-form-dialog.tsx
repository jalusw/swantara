"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Switch } from "@/components/switch";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const currencyOptions = ["USD", "IDR", "EUR", "SGD", "GBP", "JPY", "AUD", "CAD"];

function useContactFormSchema() {
  const t = useTranslations("Contacts");
  return z.object({
    name: z.string().trim().min(1, t("validationNameRequired")),
    displayName: z.string().trim(),
    isOrganization: z.boolean(),
    email: z
      .string()
      .trim()
      .refine(
        (value) => value === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        t("validationEmailInvalid"),
      ),
    phone: z.string().trim(),
    mobile: z.string().trim(),
    website: z.string().trim(),
    taxId: z.string().trim(),
    industry: z.string().trim(),
    currencyCode: z.string().trim(),
    lang: z.string().trim(),
    active: z.boolean(),
  });
}

function toDefaultValues(initial?: Contact | null) {
  return initial
    ? {
        name: initial.name,
        displayName: initial.displayName || "",
        isOrganization: initial.isOrganization,
        email: initial.email || "",
        phone: initial.phone || "",
        mobile: initial.mobile || "",
        website: initial.website || "",
        taxId: initial.taxId || "",
        industry: initial.industry || "",
        currencyCode: initial.currencyCode || "USD",
        lang: initial.lang || "en",
        active: initial.active,
      }
    : {
        name: "",
        displayName: "",
        isOrganization: true,
        email: "",
        phone: "",
        mobile: "",
        website: "",
        taxId: "",
        industry: "",
        currencyCode: "USD",
        lang: "en",
        active: true,
      };
}

export function ContactFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Contact | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);
  const t = useTranslations("Contacts");
  const tCommon = useTranslations("Common");

  const schema = useContactFormSchema();
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
      organizationId: Number(orgId),
      name: values.name,
      displayName: values.displayName || null,
      isOrganization: values.isOrganization,
      parentId: initial?.parentId ?? null,
      email: values.email || null,
      phone: values.phone || null,
      mobile: values.mobile || null,
      website: values.website || null,
      taxId: values.taxId || null,
      industry: values.industry || null,
      currencyCode: values.currencyCode || null,
      lang: values.lang || "en",
      active: values.active,
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .contacts.update(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    } else {
      return getSwantaraService()
        .contacts.create(Number(orgId), {
          ...request,
          addresses: [],
          bankAccounts: [],
          customer: null,
          supplier: null,
        })
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editContact") : t("newContact")}
      description={t("description")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
    >
      <div className="flex flex-col gap-6">
        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm">{t("identity")}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={t("fieldName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
            </FormField>
            <FormField name="displayName" label={t("fieldDisplayName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldDisplayName")} />}
            </FormField>
            <FormField name="email" label={t("fieldEmail")}>
              {({ field, id }) => (
                <Input {...field} id={id} type="email" placeholder={t("fieldEmail")} />
              )}
            </FormField>
            <FormField name="phone" label={t("fieldPhone")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldPhone")} />}
            </FormField>
            <FormField name="mobile" label={t("fieldMobile")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldMobile")} />}
            </FormField>
            <FormField name="website" label={t("fieldWebsite")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldWebsite")} />}
            </FormField>
            <FormField name="taxId" label={t("fieldTaxId")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldTaxId")} />}
            </FormField>
            <FormField name="industry" label={t("fieldIndustry")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldIndustry")} />}
            </FormField>
            <FormField name="currencyCode" label={t("fieldCurrency")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldCurrency")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {currencyOptions.map((code) => (
                      <SelectItem key={code} value={code}>
                        {code}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="lang" label={t("fieldLanguage")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldLanguage")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">{t("languageEnglish")}</SelectItem>
                    <SelectItem value="id">{t("languageIndonesian")}</SelectItem>
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="isOrganization" label={t("fieldIsOrganization")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("fieldIsOrganization")}
                />
              )}
            </FormField>
            <FormField name="active" label={t("statusActive")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("statusActive")}
                />
              )}
            </FormField>
          </div>
        </fieldset>
      </div>
    </EntityFormDialog>
  );
}
