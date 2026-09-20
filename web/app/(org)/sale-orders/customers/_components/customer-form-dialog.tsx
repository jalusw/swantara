"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
  return z.object({
    name: z.string().trim().min(1, "Enter a name."),
    displayName: z.string().trim(),
    email: z
      .string()
      .trim()
      .refine(
        (value) => value === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value),
        "Enter a valid email address.",
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
        toast.success("Customer saved.");
        onSave();
      })
      .catch(() => toast.error("Something went wrong."))
      .finally(() => setIsPending(false));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"New customer"}
      description={"Add a new customer with basic identity and contact details."}
      form={form}
      onSubmit={handleSubmit}
      isPending={isPending}
    >
      <div className="flex flex-col gap-6">
        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{"Identity"}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={`${"Name"} *`}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  placeholder={"Name"}
                  autoComplete="organization"
                  required
                  aria-required="true"
                />
              )}
            </FormField>
            <FormField name="displayName" label={"Display name"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  placeholder={"Display name"}
                  autoComplete="organization"
                />
              )}
            </FormField>
          </div>
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{"Contact"}</legend>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="email" label={"Email"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="email"
                  inputMode="email"
                  autoComplete="email"
                  placeholder={"Email"}
                />
              )}
            </FormField>
            <FormField name="phone" label={"Phone"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="tel"
                  inputMode="tel"
                  autoComplete="tel"
                  placeholder={"Phone"}
                />
              )}
            </FormField>
          </div>
        </fieldset>

        <fieldset className="flex flex-col gap-4">
          <legend className="text-sm font-medium">{"Status"}</legend>
          <FormField
            name="active"
            label={"Active"}
            description={"Inactive customers stay in history but are hidden from new orders."}
          >
            {({ field }) => (
              <Switch
                checked={field.value}
                onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                aria-label={"Active"}
              />
            )}
          </FormField>
        </fieldset>
      </div>
    </EntityFormDialog>
  );
}
