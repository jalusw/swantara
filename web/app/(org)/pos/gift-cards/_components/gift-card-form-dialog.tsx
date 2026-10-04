"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Pos");
  return z.object({
    code: z.string().optional(),
    contactId: z.string().optional(),
    amount: z.string().min(1, t("validation_amountRequired")),
    currencyCode: z.string().min(1, t("validation_currencyRequired")),
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
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
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

  async function handleSubmit(values: Values) {
    const request = getSwantaraService().giftCards.create(Number(orgId), {
      code: values.code || undefined,
      contactId: values.contactId ? Number(values.contactId) : undefined,
      amount: Number(values.amount),
      currencyCode: values.currencyCode,
      expiryDate: values.expiryDate || undefined,
      journalId: 1,
      cashAccountId: 1,
      liabilityAccountId: 1,
    });
    toast.promise(request, {
      loading: t("issuing"),
      success: (result) => {
        onSave(String(result.giftCard.id));
        return t("giftCardIssued");
      },
      error: t("giftCardIssueFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("issueGiftCard")}
      description={t("giftCardsDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="code" label={t("colCode")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("colCode")} />}
          </FormField>
          <FormField name="amount" label={t("amount")}>
            {({ field, id }) => (
              <Input
                {...field}
                id={id}
                type="number"
                min="0"
                step="0.01"
                placeholder={t("amount")}
              />
            )}
          </FormField>
          <FormField name="contactId" label={t("customer")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("customer")}>
                  <SelectValue placeholder={t("selectCustomer")} />
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
          <FormField name="currencyCode" label={t("currency")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"USD"} />}
          </FormField>
          <FormField name="expiryDate" label={t("colExpiry")}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
        </div>
      </div>
    </EntityFormDialog>
  );
}
