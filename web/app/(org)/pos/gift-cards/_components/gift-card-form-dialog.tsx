"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useGiftCardFormSchema() {
  return z.object({
    code: z.string().optional(),
    contactId: z.string().optional(),
    amount: z.string().min(1, "Amount is required"),
    currencyCode: z.string().min(1, "Currency is required"),
    expiryDate: z.string().optional(),
  });
}

export function GiftCardFormDialog({
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
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const contacts = contactsQuery.data?.contacts ?? [];

  const schema = useGiftCardFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      code: "",
      contactId: "",
      amount: "",
      currencyCode: "USD",
      expiryDate: "",
    },
  });

  function handleSubmit(values: Values) {
    void toast.promise(
      getSwantaraService().giftCards.create(Number(orgId), {
        code: values.code || undefined,
        contactId: values.contactId ? Number(values.contactId) : undefined,
        amount: Number(values.amount),
        currencyCode: values.currencyCode,
        expiryDate: values.expiryDate || undefined,
        journalId: 1,
        cashAccountId: 1,
        liabilityAccountId: 1,
      }),
      {
        loading: "Issuing…",
        success: (result) => {
          onSave(String(result.giftCard.id));
          return "Gift card issued";
        },
        error: "Failed to issue gift card",
      },
    );
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Issue gift card"}
      description={"Issue, redeem and manage gift cards."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="code" label={"Code"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Code"} />}
          </FormField>
          <FormField name="amount" label={"Amount"}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="0" step="0.01" placeholder={"Amount"} />
            )}
          </FormField>
          <FormField name="contactId" label={"Customer"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Customer"}>
                  <SelectValue placeholder={"Select customer"} />
                </SelectTrigger>
                <SelectContent>
                  {contacts.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="currencyCode" label={"Currency"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"USD"} />}
          </FormField>
          <FormField name="expiryDate" label={"Expiry date"}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
        </div>
      </div>
    </EntityFormDialog>
  );
}
