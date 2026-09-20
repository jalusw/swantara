"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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

const expenseLineSchema = z.object({
  categoryId: z.string(),
  itemId: z.string(),
  description: z.string(),
  expenseDate: z.string().min(1),
  quantity: z.string().min(1),
  unitPrice: z.string().min(1),
  reimbursable: z.boolean(),
});

function useExpenseFormSchema() {
  return z.object({
    name: z.string().min(1, "Name is required"),
    employeeId: z.string().min(1, "Employee is required"),
    paymentMode: z.enum(["own_account", "organization_account"]),
    lines: z.array(expenseLineSchema).min(1, "At least one line is required"),
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

  function handleSubmit(values: ExpenseFormValues) {
    void toast.promise(
      getSwantaraService().expenseReports.create(Number(orgId), {
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
      }),
      {
        loading: "Saving…",
        success: () => {
          onSave();
          return "Expense report created";
        },
        error: "Failed to create expense report",
      },
    );
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Create expense report"}
      description={"Description"}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={"Name"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
          </FormField>
          <FormField name="employeeId" label={"Employee"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Employee"}>
                  <SelectValue placeholder={"Select employee"} />
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
          <FormField name="paymentMode" label={"Payment mode"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Payment mode"}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="own_account">{"Own account"}</SelectItem>
                  <SelectItem value="organization_account">{"Organization account"}</SelectItem>
                </SelectContent>
              </Select>
            )}
          </FormField>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm">{"Expense lines"}</span>
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
              {"Add line"}
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
                  {"Category"}
                </label>
                <Select
                  value={form.watch(`lines.${index}.categoryId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.categoryId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-category-${index}`} aria-label={"Category"}>
                    <SelectValue placeholder={"Select category"} />
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
                  {"Item"}
                </label>
                <Select
                  value={form.watch(`lines.${index}.itemId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={"Item"}>
                    <SelectValue placeholder={"Select item"} />
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
                  {"Date"}
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
                  {"Quantity"}
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
                  {"Unit price"}
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
