"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useCallback, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useLineResetEffect } from "@/lib/hooks/use-line-reset-effect";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { createId, getLocalDateString } from "@/lib/utils";

type LineRow = {
  id: string;
  itemId: string;
  description: string;
  qty: string;
  neededBy: string;
};

const supplierQuoteRequestSchema = z.object({
  supplierId: z.string(),
  currencyCode: z.string(),
  orderDate: z.string(),
  quoteDeadline: z.string(),
  notes: z.string(),
});
type SupplierQuoteRequestFormValues = z.infer<typeof supplierQuoteRequestSchema>;

export function SupplierQuoteRequestFormDialog({
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
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
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

  const [lines, setLines] = useState<LineRow[]>([
    { id: createId(), itemId: "", description: "", qty: "1", neededBy: "" },
  ]);

  const createEmptyLine = useCallback(
    () => ({
      id: createId(),
      itemId: "",
      description: "",
      qty: "1",
      neededBy: "",
    }),
    [],
  );

  useLineResetEffect(open, setLines, createEmptyLine);

  const form = useForm<SupplierQuoteRequestFormValues>({
    resolver: zodResolver(supplierQuoteRequestSchema),
    defaultValues: {
      supplierId: "",
      currencyCode: "",
      orderDate: getLocalDateString(),
      quoteDeadline: "",
      notes: "",
    },
  });

  function addLine() {
    setLines((prev) => [
      ...prev,
      {
        id: createId(),
        itemId: "",
        description: "",
        qty: "1",
        neededBy: "",
      },
    ]);
  }

  function updateLine(id: string, patch: Partial<LineRow>) {
    setLines((prev) => prev.map((line) => (line.id === id ? { ...line, ...patch } : line)));
  }

  function removeLine(id: string) {
    setLines((prev) => prev.filter((line) => line.id !== id));
  }

  function handleSubmit(values: SupplierQuoteRequestFormValues) {
    if (lines.length === 0) {
      toast.error(t("addAtLeastOneLine"));
      return;
    }
    const invalid = lines.some((line) => !line.itemId || Number(line.qty) <= 0);
    if (invalid) {
      toast.error(t("quantityMustBePositive"));
      return;
    }
    const request = {
      organizationId: Number(orgId),
      requesterId: 0,
      supplierId: values.supplierId ? Number(values.supplierId) : null,
      currencyCode: values.currencyCode || null,
      orderDate: values.orderDate ? new Date(values.orderDate) : null,
      quoteDeadline: values.quoteDeadline ? new Date(values.quoteDeadline) : null,
      notes: values.notes || null,
      lines: lines.map((line) => ({
        itemId: Number(line.itemId),
        description: line.description || null,
        qty: Number(line.qty),
        unitId: null as number | null,
        neededBy: line.neededBy ? new Date(line.neededBy) : null,
      })),
    };
    return getSwantaraService()
      .supplierQuoteRequests.create(Number(orgId), request)
      .then(() => {
        toast.success(t("quoteRequestCreated"));
        onSave();
      })
      .catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newQuoteRequest")}
      description={t("newQuoteRequestDescription")}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="supplierId" label={t("tableSupplier")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableSupplier")}>
                <SelectValue placeholder={t("selectSupplier")} />
              </SelectTrigger>
              <SelectContent>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="currencyCode" label={t("fieldCurrency")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("currencyPlaceholder")} />}
        </FormField>
        <FormField name="orderDate" label={t("tableOrderDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="quoteDeadline" label={t("tableQuoteDeadline")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="notes" label={t("fieldNote")}>
          {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
        </FormField>
      </div>

      <div className="mt-6 flex flex-col gap-3 rounded-md border p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm">{t("quoteRequestLines")}</h3>
          <Button type="button" variant="outline" size="sm" onClick={addLine}>
            {t("addLine")}
          </Button>
        </div>
        {lines.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("addAtLeastOneLine")}</p>
        ) : null}
        <div className="flex flex-col gap-3">
          {lines.map((line, index) => (
            <div key={line.id} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
              <div className="sm:col-span-5">
                <span className="text-xs text-muted-foreground">{t("fieldItem")}</span>
                <Select
                  value={line.itemId}
                  onValueChange={(value) => updateLine(line.id, { itemId: value ?? "" })}
                >
                  <SelectTrigger aria-label={`${t("fieldItem")} ${index + 1}`}>
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
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{t("fieldQty")}</span>
                <Input
                  value={line.qty}
                  onChange={(e) => updateLine(line.id, { qty: e.target.value })}
                  type="number"
                  min="0"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{t("fieldNeededBy")}</span>
                <Input
                  value={line.neededBy}
                  onChange={(e) => updateLine(line.id, { neededBy: e.target.value })}
                  type="date"
                />
              </div>
              <div className="sm:col-span-2 flex items-end">
                <Button type="button" variant="ghost" size="sm" onClick={() => removeLine(line.id)}>
                  {tCommon("delete")}
                </Button>
              </div>
              <div className="sm:col-span-12">
                <Input
                  value={line.description}
                  onChange={(e) => updateLine(line.id, { description: e.target.value })}
                  placeholder={t("fieldDescription")}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </EntityFormDialog>
  );
}
