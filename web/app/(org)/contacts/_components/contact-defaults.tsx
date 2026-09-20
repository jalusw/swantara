"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
  const schema = useCustomerDefaultsFormSchema();

  const form = useForm<CustomerDefaultsFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      creditLimit: "",
    },
  });

  function handleSubmit(values: CustomerDefaultsFormValues) {
    void getSwantaraService()
      .contacts.customer(Number(orgId), Number(contactId), {
        customerPaymentTermId: null,
        creditLimit: values.creditLimit ? Number(values.creditLimit) : null,
        receivableAccountId: null,
        active: true,
      })
      .then(() => {
        toast.success("Defaults saved.");
        onRefetch();
      })
      .catch(() => {
        toast.error("Something went wrong. Please try again.");
      });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Customer"}</CardTitle>
        <CardDescription>{"Enables the customer role and its defaults."}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="creditLimit" label={"Credit limit"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  step="any"
                  inputMode="decimal"
                  placeholder={"Credit limit"}
                />
              )}
            </FormField>
          </div>
          <SubmitButton>{"Save defaults"}</SubmitButton>
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
  const schema = useSupplierDefaultsFormSchema();

  const form = useForm<SupplierDefaultsFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {},
  });

  function handleSubmit(_values: SupplierDefaultsFormValues) {
    void getSwantaraService()
      .contacts.supplier(Number(orgId), Number(contactId), {
        vendorPaymentTermId: null,
        payableAccountId: null,
        active: true,
      })
      .then(() => {
        toast.success("Defaults saved.");
        onRefetch();
      })
      .catch(() => {
        toast.error("Something went wrong. Please try again.");
      });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Supplier"}</CardTitle>
        <CardDescription>{"Enables the supplier role and its defaults."}</CardDescription>
      </CardHeader>
      <CardContent>
        <Form form={form} onSubmit={handleSubmit}>
          <SubmitButton>{"Save defaults"}</SubmitButton>
        </Form>
      </CardContent>
    </Card>
  );
}
