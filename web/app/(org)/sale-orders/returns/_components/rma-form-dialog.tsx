"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Trash2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, CreateRmaRequest, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { type Disposition, dispositionLabel, dispositionValues } from "./rma-utils";

type RmaFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: () => void;
};

const rmaLineSchema = z.object({
  id: z.string(),
  itemId: z.string().min(1),
  qty: z.string().min(1),
  disposition: z.string().min(1),
});

function useRmaFormSchema() {
  const t = useTranslations("Sales");
  return z.object({
    rmaType: z.enum(["customer_return", "vendor_return"] as const),
    contactId: z.string().min(1, t("validationContactRequired")),
    originOrderType: z.enum(["sale_order", "purchase_order"] as const),
    originOrderId: z.string().min(1, t("validationOriginOrderRequired")),
    reason: z.string(),
    lines: z.array(rmaLineSchema).min(1, t("addAtLeastOneLine")),
  });
}

export function RmaFormDialog({ open, onOpenChange, orgId, onSave }: RmaFormDialogProps) {
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();

  const schema = useRmaFormSchema();
  type Values = z.infer<typeof schema>;

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const products = productsQuery.data?.products ?? [];

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      rmaType: "customer_return",
      contactId: "",
      originOrderType: "sale_order",
      originOrderId: "",
      reason: "",
      lines: [{ id: "1", itemId: "", qty: "1", disposition: "restock" }],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
    keyName: "fieldId",
  });

  const createMutation = useMutation({
    mutationFn: (payload: CreateRmaRequest) =>
      getSwantaraService().rmas.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success(t("rmaCreated"));
      void queryClient.invalidateQueries({ queryKey: ["rmas", orgId] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  function handleSubmit(values: Values) {
    createMutation.mutate({
      organizationId: Number(orgId),
      type: values.rmaType,
      contactId: Number(values.contactId),
      originOrderType: values.originOrderType,
      originOrderId: Number(values.originOrderId),
      reason: values.reason,
      lines: values.lines
        .filter((l) => l.itemId)
        .map((l) => ({
          itemId: Number(l.itemId),
          qty: Number(l.qty) || 1,
          batchId: null,
          disposition: l.disposition as Disposition,
        })),
    });
  }

  function rmaTypeLabel(key: string): string {
    try {
      return (t as unknown as (k: string) => string)(`rmaType.${key}`);
    } catch {
      return key;
    }
  }

  function dispositionOptionLabel(value: string): string {
    try {
      return (t as unknown as (k: string) => string)(`disposition.${value}`);
    } catch {
      return dispositionLabel(value as Disposition);
    }
  }

  const originOrderTypes = [
    { value: "sale_order", label: t("originSaleOrder") },
    { value: "purchase_order", label: t("originPurchaseOrder") },
  ] as const;

  const dispositionOptions = dispositionValues.map((value) => ({
    value,
    label: dispositionOptionLabel(value),
  }));

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newRma")}
      description={t("newRmaDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="rmaType" label={t("fieldRmaType")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("fieldRmaType")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="customer_return">{rmaTypeLabel("customer_return")}</SelectItem>
                  <SelectItem value="vendor_return">{rmaTypeLabel("vendor_return")}</SelectItem>
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="contactId" label={t("fieldContact")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("fieldContact")}>
                  <SelectValue placeholder={t("selectContact")} />
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
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="originOrderType" label={t("fieldOriginOrderType")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("fieldOriginOrderType")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {originOrderTypes.map((ot) => (
                    <SelectItem key={ot.value} value={ot.value}>
                      {ot.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="originOrderId" label={t("fieldOriginOrderId")}>
            {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
          </FormField>
        </div>
        <FormField name="reason" label={t("fieldReason")}>
          {({ field, id }) => <Textarea {...field} id={id} rows={3} />}
        </FormField>
        <div className="flex flex-col gap-3 rounded-md border p-4">
          <div className="flex items-center justify-between">
            <h3 className="text-sm">{t("returnLines")}</h3>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() =>
                append({
                  id: String(Date.now()),
                  itemId: "",
                  qty: "1",
                  disposition: "restock",
                })
              }
            >
              <Plus />
              <span>{t("addLine")}</span>
            </Button>
          </div>
          {fields.map((field, index) => (
            <div key={field.fieldId} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
              <div className="sm:col-span-4">
                <span className="text-xs text-muted-foreground">{t("fieldItem")}</span>
                <FormField name={`lines.${index}.itemId`}>
                  {({ field: lineField, id }) => (
                    <Select value={lineField.value} onValueChange={lineField.onChange}>
                      <SelectTrigger id={id} aria-label={t("fieldItem")}>
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
                  )}
                </FormField>
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{t("fieldQty")}</span>
                <FormField name={`lines.${index}.qty`}>
                  {({ field: lineField, id }) => (
                    <Input {...lineField} id={id} type="number" min="0" step="any" />
                  )}
                </FormField>
              </div>
              <div className="sm:col-span-4">
                <span className="text-xs text-muted-foreground">{t("fieldDisposition")}</span>
                <FormField name={`lines.${index}.disposition`}>
                  {({ field: lineField, id }) => (
                    <Select value={lineField.value} onValueChange={lineField.onChange}>
                      <SelectTrigger id={id} aria-label={t("fieldDisposition")}>
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {dispositionOptions.map((d) => (
                          <SelectItem key={d.value} value={d.value}>
                            {d.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormField>
              </div>
              <div className="sm:col-span-2 flex items-end">
                <Button type="button" variant="ghost" size="sm" onClick={() => remove(index)}>
                  <Trash2 />
                </Button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </EntityFormDialog>
  );
}
