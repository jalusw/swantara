"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import type { Employee, ExpenseCategory, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function useExpenseLineSchema() {
  return z.object({
    categoryId: z.string(),
    itemId: z.string(),
    description: z.string(),
    expenseDate: z.string().min(1),
    quantity: z.string().min(1),
    unitPrice: z.string().min(1),
    reimbursable: z.boolean(),
  });
}

function useExpenseFormSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Expenses");
  const lineSchema = useExpenseLineSchema();
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    employeeId: z.string().min(1, t("validationEmployeeRequired")),
    paymentMode: z.enum(["own_account", "organization_account"]),
    lines: z.array(lineSchema).min(1, t("validationLinesRequired")),
  });
}
type ExpenseFormValues = z.infer<ReturnType<typeof useExpenseFormSchema>>;

export function ExpenseFormDialog({
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
  const t = (useTranslations as unknown as (ns: string) => TFn)("Expenses");
  const tCommon = useTranslations("Common");
  const linePrefix = useId();

  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );
  const employees = employeesQuery.data?.employees ?? [];

  const categoriesQuery = useOrgListQuery<{ categories: ExpenseCategory[] }, Record<string, never>>(
    "expenseCategories",
    (organizationId) => getSwantaraService().expenseCategories.list(organizationId),
  );
  const categories = categoriesQuery.data?.categories ?? [];

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );
  const products = productsQuery.data?.products ?? [];

  const schema = useExpenseFormSchema();

  const form = useForm<ExpenseFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      employeeId: "",
      paymentMode: "own_account",
      lines: [
        {
          categoryId: "",
          itemId: "",
          description: "",
          expenseDate: getLocalDateString(),
          quantity: "1",
          unitPrice: "0",
          reimbursable: true,
        },
      ],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });

  async function handleSubmit(values: ExpenseFormValues) {
    const request = getSwantaraService().expenseReports.create(Number(orgId), {
      name: values.name,
      employeeId: Number(values.employeeId),
      paymentMode: values.paymentMode,
      lines: values.lines.map((l) => ({
        categoryId: l.categoryId ? Number(l.categoryId) : null,
        itemId: l.itemId ? Number(l.itemId) : null,
        description: l.description,
        expenseDate: l.expenseDate,
        quantity: Number(l.quantity),
        unitPrice: Number(l.unitPrice),
        taxIds: [],
        currencyCode: "USD",
        dimensionId: null,
        projectId: null,
        reimbursable: l.reimbursable,
        receiptAttachmentId: null,
      })),
    });
    toast.promise(request, {
      loading: t("saving"),
      success: () => {
        onSave();
        return t("toastReportCreated");
      },
      error: t("toastReportFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newReport")}
      description={t("reportDialogDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={t("fieldName")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
          </FormField>
          <FormField name="employeeId" label={t("fieldEmployee")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("fieldEmployee")}>
                  <SelectValue placeholder={t("selectEmployee")} />
                </SelectTrigger>
                <SelectContent>
                  {employees.map((e) => (
                    <SelectItem key={e.id} value={String(e.id)}>
                      #{e.employeeNumber}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="paymentMode" label={t("fieldPaymentMode")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("fieldPaymentMode")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="own_account">
                    {(t as unknown as (k: string) => string)("paymentMode_own_account")}
                  </SelectItem>
                  <SelectItem value="organization_account">
                    {(t as unknown as (k: string) => string)("paymentMode_organization_account")}
                  </SelectItem>
                </SelectContent>
              </Select>
            )}
          </FormField>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm">{t("expenseLines")}</span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() =>
                append({
                  categoryId: "",
                  itemId: "",
                  description: "",
                  expenseDate: getLocalDateString(),
                  quantity: "1",
                  unitPrice: "0",
                  reimbursable: true,
                })
              }
            >
              {t("addLine")}
            </Button>
          </div>
          {fields.map((field, index) => (
            <div
              key={field.id}
              className="grid grid-cols-[1fr_1fr_120px_100px_80px_40px] items-end gap-2"
            >
              <div>
                <label
                  htmlFor={`${linePrefix}-category-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("fieldCategory")}
                </label>
                <Select
                  value={form.watch(`lines.${index}.categoryId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.categoryId`, v ?? "")}
                >
                  <SelectTrigger
                    id={`${linePrefix}-category-${index}`}
                    aria-label={t("fieldCategory")}
                  >
                    <SelectValue placeholder={t("selectCategory")} />
                  </SelectTrigger>
                  <SelectContent>
                    {categories.map((c) => (
                      <SelectItem key={c.id} value={String(c.id)}>
                        {c.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-item-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("fieldItem")}
                </label>
                <Select
                  value={form.watch(`lines.${index}.itemId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={t("fieldItem")}>
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
                  htmlFor={`${linePrefix}-date-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("colDate")}
                </label>
                <Input
                  id={`${linePrefix}-date-${index}`}
                  type="date"
                  {...form.register(`lines.${index}.expenseDate`)}
                />
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-qty-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("colQuantity")}
                </label>
                <Input
                  id={`${linePrefix}-qty-${index}`}
                  type="number"
                  min="1"
                  {...form.register(`lines.${index}.quantity`)}
                />
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-price-${index}`}
                  className="text-muted-foreground text-xs"
                >
                  {t("colUnitPrice")}
                </label>
                <Input
                  id={`${linePrefix}-price-${index}`}
                  type="number"
                  min="0"
                  step="0.01"
                  {...form.register(`lines.${index}.unitPrice`)}
                />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => remove(index)}
                disabled={fields.length <= 1}
                aria-label={tCommon("delete")}
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
