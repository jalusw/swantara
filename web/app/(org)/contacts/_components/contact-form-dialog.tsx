"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
  return z.object({
    name: z.string().trim().min(1, "Enter a name."),
    displayName: z.string().trim(),
    isOrganization: z.boolean(),
    email: z
      .string()
      .trim()
      .refine(
        (value) => value === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        "Enter a valid email address.",
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

  const schema = useContactFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
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
        },
  });

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
      void getSwantaraService()
        .contacts.update(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .contacts.create(Number(orgId), {
          ...request,
          addresses: [],
          bankAccounts: [],
          customer: null,
          supplier: null,
        })
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit contact" : "New contact"}
      description={"Unified identity for customers, suppliers, and employees."}
      form={form}
      onSubmit={handleSubmit}
    >
      <div className="flex flex-col gap-6">
        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm">{"Identity"}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={"Name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
            </FormField>
            <FormField name="displayName" label={"Display name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Display name"} />}
            </FormField>
            <FormField name="email" label={"Email"}>
              {({ field, id }) => <Input {...field} id={id} type="email" placeholder={"Email"} />}
            </FormField>
            <FormField name="phone" label={"Phone"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Phone"} />}
            </FormField>
            <FormField name="mobile" label={"Mobile"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Mobile"} />}
            </FormField>
            <FormField name="website" label={"Website"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Website"} />}
            </FormField>
            <FormField name="taxId" label={"Tax ID"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Tax ID"} />}
            </FormField>
            <FormField name="industry" label={"Industry"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Industry"} />}
            </FormField>
            <FormField name="currencyCode" label={"Currency"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Currency"}>
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
            <FormField name="lang" label={"Language"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Language"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="en">English</SelectItem>
                    <SelectItem value="id">Bahasa Indonesia</SelectItem>
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="isOrganization" label={"This is an organization"}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={"This is an organization"}
                />
              )}
            </FormField>
            <FormField name="active" label={"Active"}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={"Active"}
                />
              )}
            </FormField>
          </div>
        </fieldset>
      </div>
    </EntityFormDialog>
  );
}
