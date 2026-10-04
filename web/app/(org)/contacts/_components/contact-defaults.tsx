"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export function ContactDefaults({
  orgId,
  contactId,
  contact: _contact,
  onRefetch,
}: {
  orgId: string;
  contactId: string;
  contact: Contact;
  onRefetch: () => void;
}) {
  return (
    <div className="flex flex-col gap-4">
      <CustomerDefaultsPanel orgId={orgId} contactId={contactId} onRefetch={onRefetch} />
      <SupplierDefaultsPanel orgId={orgId} contactId={contactId} onRefetch={onRefetch} />
    </div>
  );
}

function useCustomerDefaultsFormSchema() {
  return z.object({
    creditLimit: z.string(),
  });
}
type CustomerDefaultsFormValues = z.infer<ReturnType<typeof useCustomerDefaultsFormSchema>>;

function CustomerDefaultsPanel({
  orgId,
  contactId,
  onRefetch,
}: {
  orgId: string;
  contactId: string;
  onRefetch: () => void;
}) {
  const t = useTranslations("Contacts");
  const schema = useCustomerDefaultsFormSchema();

  const form = useForm<CustomerDefaultsFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      creditLimit: "",
    },
  });

  function handleSubmit(values: CustomerDefaultsFormValues) {
    return getSwantaraService()
      .contacts.customer(Number(orgId), Number(contactId), {
        customerPaymentTermId: null,
        creditLimit: values.creditLimit ? Number(values.creditLimit) : null,
        receivableAccountId: null,
        active: true,
      })
      .then(() => {
        toast.success(t("defaultsSaved"));
        onRefetch();
      })
      .catch(() => {
        toast.error(t("saveFailedTryAgain"));
      });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("customerDefaultsTitle")}</CardTitle>
        <CardDescription>{t("customerDefaultsDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="creditLimit" label={t("fieldCreditLimit")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  step="any"
                  inputMode="decimal"
                  placeholder={t("fieldCreditLimit")}
                />
              )}
            </FormField>
          </div>
          <SubmitButton>{t("saveDefaults")}</SubmitButton>
        </Form>
      </CardContent>
    </Card>
  );
}

function useSupplierDefaultsFormSchema() {
  return z.object({ _placeholder: z.string().optional() });
}
type SupplierDefaultsFormValues = z.infer<ReturnType<typeof useSupplierDefaultsFormSchema>>;

function SupplierDefaultsPanel({
  orgId,
  contactId,
  onRefetch,
}: {
  orgId: string;
  contactId: string;
  onRefetch: () => void;
}) {
  const t = useTranslations("Contacts");
  const schema = useSupplierDefaultsFormSchema();

  const form = useForm<SupplierDefaultsFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {},
  });

  function handleSubmit(_values: SupplierDefaultsFormValues) {
    return getSwantaraService()
      .contacts.supplier(Number(orgId), Number(contactId), {
        vendorPaymentTermId: null,
        payableAccountId: null,
        active: true,
      })
      .then(() => {
        toast.success(t("defaultsSaved"));
        onRefetch();
      })
      .catch(() => {
        toast.error(t("saveFailedTryAgain"));
      });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("supplierDefaultsTitle")}</CardTitle>
        <CardDescription>{t("supplierDefaultsDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form form={form} onSubmit={handleSubmit}>
          <SubmitButton>{t("saveDefaults")}</SubmitButton>
        </Form>
      </CardContent>
    </Card>
  );
}
