"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useId } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Item, SubscriptionPlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const subscriptionLineSchema = z.object({
  itemId: z.string().min(1),
  qty: z.string().min(1),
  unitPrice: z.string(),
  discountPct: z.string(),
});

function useSubscriptionFormSchema() {
  const t = useTranslations("Subscriptions");
  return z.object({
    name: z.string().min(1, t("validation_nameRequired")),
    contactId: z.string().min(1, t("validation_customerRequired")),
    planId: z.string().min(1, t("validation_planRequired")),
    currencyCode: z.string().min(1, t("validation_currencyRequired")),
    lines: z.array(subscriptionLineSchema).min(1, t("validation_linesRequired")),
  });
}
type SubscriptionFormValues = z.infer<ReturnType<typeof useSubscriptionFormSchema>>;

export function SubscriptionFormDialog({
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
  const t = useTranslations("Subscriptions");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();
  const linePrefix = useId();

  const plansQuery = useOrgListQuery<{ plans: SubscriptionPlan[] }, Record<string, never>>(
    "subscriptionPlans",
    (organizationId) => getSwantaraService().subscriptionPlans.list(organizationId),
  );
  const plans = plansQuery.data?.plans ?? [];

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const contacts = contactsQuery.data?.contacts ?? [];

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );
  const products = productsQuery.data?.products ?? [];

  const schema = useSubscriptionFormSchema();

  const form = useForm<SubscriptionFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      contactId: "",
      planId: "",
      currencyCode: "USD",
      lines: [{ itemId: "", qty: "1", unitPrice: "0", discountPct: "0" }],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });

  const createMutation = useMutation({
    mutationFn: (values: SubscriptionFormValues) =>
      getSwantaraService().subscriptions.create(Number(orgId), {
        name: values.name,
        contactId: Number(values.contactId),
        planId: Number(values.planId),
        priceBookId: 0,
        currencyCode: values.currencyCode,
        lines: values.lines.map((l) => ({
          itemId: Number(l.itemId),
          qty: Number(l.qty),
          unitPrice: l.unitPrice ? Number(l.unitPrice) : 0,
          discountPct: l.discountPct ? Number(l.discountPct) : 0,
        })),
      }),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["subscriptions", Number(orgId)] });
      onSave(String(result.subscription.id));
      toast.success(t("created"));
    },
    onError: () => {
      toast.error(t("createFailed"));
    },
  });

  function handleSubmit(values: SubscriptionFormValues) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newSubscription")}
      description={t("formDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={t("fieldName")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
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
          <FormField name="planId" label={t("plan")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("plan")}>
                  <SelectValue placeholder={t("selectPlan")} />
                </SelectTrigger>
                <SelectContent>
                  {plans.map((p) => (
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
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm">{t("lines")}</span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() =>
                append({
                  itemId: "",
                  qty: "1",
                  unitPrice: "0",
                  discountPct: "0",
                })
              }
            >
              {t("addLine")}
            </Button>
          </div>
          {fields.map((field, index) => (
            <div
              key={field.id}
              className="grid grid-cols-[1fr_80px_100px_80px_40px] items-end gap-2"
            >
              <div>
                <label
                  htmlFor={`${linePrefix}-item-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("item")}
                </label>
                <Select
                  value={form.watch(`lines.${index}.itemId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={t("item")}>
                    <SelectValue placeholder={t("selectItem")} />
                  </SelectTrigger>
                  <SelectContent>
                    {products.map((p) => (
                      <SelectItem key={p.id} value={String(p.id)}>
                        {p.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-qty-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("qty")}
                </label>
                <Input
                  id={`${linePrefix}-qty-${index}`}
                  type="number"
                  min="1"
                  {...form.register(`lines.${index}.qty`)}
                />
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-price-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("price")}
                </label>
                <Input
                  id={`${linePrefix}-price-${index}`}
                  type="number"
                  min="0"
                  step="0.01"
                  {...form.register(`lines.${index}.unitPrice`)}
                />
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-discount-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("discountPct")}
                </label>
                <Input
                  id={`${linePrefix}-discount-${index}`}
                  type="number"
                  min="0"
                  max="100"
                  {...form.register(`lines.${index}.discountPct`)}
                />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => remove(index)}
                disabled={fields.length <= 1}
              >
                ×
              </Button>
            </div>
          ))}
        </div>
      </div>
    </EntityFormDialog>
  );
}
