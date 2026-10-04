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
import type { Contact, Department, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { createId } from "@/lib/utils";

type LineRow = {
  id: string;
  itemId: string;
  description: string;
  qty: string;
  neededBy: string;
};

function usePurchaseRequestFormSchema() {
  const t = useTranslations("Purchases");
  return z.object({
    requesterId: z.string().min(1, t("validationRequesterRequired")),
    departmentId: z.string(),
    neededBy: z.string(),
    note: z.string(),
  });
}
type PurchaseRequestFormValues = z.infer<ReturnType<typeof usePurchaseRequestFormSchema>>;

export function PurchaseRequestFormDialog({
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
  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );
  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const departments = departmentsQuery.data?.departments ?? [];
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

  const schema = usePurchaseRequestFormSchema();

  const form = useForm<PurchaseRequestFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      requesterId: "",
      departmentId: "",
      neededBy: "",
      note: "",
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

  function handleSubmit(values: PurchaseRequestFormValues) {
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
      requesterId: Number(values.requesterId),
      departmentId: values.departmentId ? Number(values.departmentId) : null,
      neededBy: values.neededBy || null,
      lines: lines.map((line) => ({
        itemId: Number(line.itemId),
        description: line.description || null,
        qty: Number(line.qty),
        unitId: null as number | null,
        neededBy: line.neededBy || null,
      })),
    };
    return getSwantaraService()
      .purchaseRequests.create(Number(orgId), request)
      .then(() => {
        toast.success(t("requestCreated"));
        onSave();
      })
      .catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newPurchaseRequest")}
      description={t("newPurchaseRequestDescription")}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="requesterId" label={t("tableRequester")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableRequester")}>
                <SelectValue placeholder={t("selectRequester")} />
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
        <FormField name="departmentId" label={t("tableDepartment")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableDepartment")}>
                <SelectValue placeholder={t("selectDepartment")} />
              </SelectTrigger>
              <SelectContent>
                {departments.map((d) => (
                  <SelectItem key={d.id} value={String(d.id)}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="neededBy" label={t("fieldNeededBy")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="note" label={t("fieldNote")}>
          {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
        </FormField>
      </div>

      <div className="mt-6 flex flex-col gap-3 rounded-md border p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm">{t("requestLines")}</h3>
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
