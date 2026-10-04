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
import type { Account, SalaryRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";

const categories = ["earning", "deduction"] as const;
const computeTypes = ["fixed", "percent", "formula"] as const;

function useSalaryRuleFormSchema() {
  const t = useTranslations("Payroll");
  return z.object({
    code: z.string().min(1, t("validationCodeRequired")),
    name: z.string().min(1, t("validationNameRequired")),
    category: z.enum(categories),
    computeType: z.enum(computeTypes),
    amount: z.string(),
    formula: z.string(),
    accountDebitId: z.string(),
    accountCreditId: z.string(),
  });
}
type SalaryRuleFormValues = z.infer<ReturnType<typeof useSalaryRuleFormSchema>>;

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
  const t = useTranslations("Payroll");
  const tCommon = useTranslations("Common");

  const form = useForm<SalaryRuleFormValues>({
    resolver: zodResolver(useSalaryRuleFormSchema()),
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

  function categoryLabel(cat: string): string {
    try {
      return (t as unknown as (k: string) => string)(`category.${cat}`);
    } catch {
      return humanizeKey(String(cat));
    }
  }

  function computeTypeLabel(ct: string): string {
    try {
      return (t as unknown as (k: string) => string)(`computeType.${ct}`);
    } catch {
      return humanizeKey(String(ct));
    }
  }

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
      return getSwantaraService()
        .salaryRules.update(Number(orgId), initial.id, body)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    } else {
      return getSwantaraService()
        .salaryRules.create(Number(orgId), {
          ...body,
          organizationId: Number(orgId),
        })
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editSalaryRule") : t("newSalaryRule")}
      description={t("newSalaryRuleDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="code" label={t("fieldCode")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldCode")} />}
        </FormField>
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="category" label={t("tableCategory")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableCategory")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat} value={cat}>
                    {categoryLabel(String(cat))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="computeType" label={t("tableComputeType")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableComputeType")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {computeTypes.map((ct) => (
                  <SelectItem key={ct} value={ct}>
                    {computeTypeLabel(String(ct))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="amount" label={t("tableAmount")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="any" />}
        </FormField>
        <FormField name="formula" label={t("fieldFormula")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("formulaPlaceholder")} />}
        </FormField>
        <FormField name="accountDebitId" label={t("fieldDebitAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldDebitAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
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
        <FormField name="accountCreditId" label={t("fieldCreditAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldCreditAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
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
