"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import type { Account, SalaryRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";

const categories = ["earning", "deduction"] as const;
const computeTypes = ["fixed", "percent", "formula"] as const;

const salaryRuleFormSchema = z.object({
  code: z.string().min(1),
  name: z.string().min(1),
  category: z.enum(categories),
  computeType: z.enum(computeTypes),
  amount: z.string(),
  formula: z.string(),
  accountDebitId: z.string(),
  accountCreditId: z.string(),
});
type SalaryRuleFormValues = z.infer<typeof salaryRuleFormSchema>;

export function SalaryRuleFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  accounts,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: SalaryRule | null;
  accounts: Account[];
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const form = useForm<SalaryRuleFormValues>({
    resolver: zodResolver(salaryRuleFormSchema),
    defaultValues: initial
      ? {
          code: initial.code,
          name: initial.name,
          category: initial.category ?? "earning",
          computeType: initial.computeType ?? "fixed",
          amount: initial.amount != null ? String(initial.amount) : "",
          formula: initial.formula ?? "",
          accountDebitId: initial.accountDebitId != null ? String(initial.accountDebitId) : "",
          accountCreditId: initial.accountCreditId != null ? String(initial.accountCreditId) : "",
        }
      : {
          code: "",
          name: "",
          category: "earning",
          computeType: "fixed",
          amount: "",
          formula: "",
          accountDebitId: "",
          accountCreditId: "",
        },
  });

  function handleSubmit(values: SalaryRuleFormValues) {
    const body = {
      code: values.code,
      name: values.name,
      category: values.category,
      computeType: values.computeType,
      amount: values.amount ? Number(values.amount) : null,
      formula: values.formula || null,
      accountDebitId: values.accountDebitId ? Number(values.accountDebitId) : null,
      accountCreditId: values.accountCreditId ? Number(values.accountCreditId) : null,
    };
    if (isEdit && initial) {
      void getSwantaraService()
        .salaryRules.update(Number(orgId), initial.id, body)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .salaryRules.create(Number(orgId), {
          ...body,
          organizationId: Number(orgId),
        })
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit salary rule" : "Create asset category"}
      description={"Create a new salary rule."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="code" label={"Code"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Code"} />}
        </FormField>
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="category" label={"Category"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Category"}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {humanizeKey(String(cat))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="computeType" label={"Compute type"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Compute type"}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {computeTypes.map((ct) => (
                  <SelectItem key={ct} value={ct}>
                    {humanizeKey(String(ct))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="amount" label={"Amount"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="any" />}
        </FormField>
        <FormField name="formula" label={"Formula"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"e.g. base * 0.1"} />}
        </FormField>
        <FormField name="accountDebitId" label={"Debit account"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Debit account"}>
                <SelectValue placeholder={"Select account"} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="accountCreditId" label={"Credit account"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Credit account"}>
                <SelectValue placeholder={"Select account"} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
