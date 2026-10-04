"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/combobox";
import { CurrencyField } from "@/components/currency-field";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Department, Employee, JobPosition } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { humanizeKey } from "@/lib/utils/case";

const employmentTypes = ["full_time", "part_time", "contract"] as const;
const wageTypes = ["monthly", "hourly"] as const;
const currencyOptions = ["IDR", "USD", "EUR", "SGD", "GBP", "JPY", "AUD", "CAD"];

function useEmployeeFormSchema() {
  const t = useTranslations("Employees");
  return z.object({
    name: z.string().trim().min(1, t("validationNameRequired")),
    employeeNumber: z.string().trim().min(1, t("validationEmployeeNumberRequired")),
    email: z.string().trim(),
    phone: z.string().trim(),
    userId: z.string().trim(),
    departmentId: z.string().trim(),
    jobPositionId: z.string().trim(),
    hireDate: z.string().trim(),
    employmentType: z.enum(employmentTypes),
    workLocation: z.string().trim(),
    wage: z
      .string()
      .trim()
      .refine((value) => Number(value) >= 0, t("validationWageInvalid")),
    wageType: z.enum(wageTypes),
    currencyCode: z.string().trim(),
  });
}
type EmployeeFormValues = z.infer<ReturnType<typeof useEmployeeFormSchema>>;

export function EmployeeFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Employee | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);
  const t = useTranslations("Employees");
  const tCommon = useTranslations("Common");

  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const jobPositionsQuery = useOrgListQuery<{ jobPositions: JobPosition[] }, Record<string, never>>(
    "jobPositions",
    (organizationId) => getSwantaraService().jobPositions.list(organizationId),
  );

  const departments = departmentsQuery.data?.departments ?? [];
  const jobPositions = jobPositionsQuery.data?.jobPositions ?? [];

  const schema = useEmployeeFormSchema();

  const form = useForm<EmployeeFormValues>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: "",
          employeeNumber: initial.employeeNumber,
          email: "",
          phone: "",
          userId: initial.userId != null ? String(initial.userId) : "",
          departmentId: initial.departmentId != null ? String(initial.departmentId) : "",
          jobPositionId: initial.jobPositionId != null ? String(initial.jobPositionId) : "",
          hireDate: initial.hireDate ?? "",
          employmentType: initial.employmentType,
          workLocation: initial.workLocation ?? "",
          wage: "0",
          wageType: "monthly",
          currencyCode: "IDR",
        }
      : {
          name: "",
          employeeNumber: "",
          email: "",
          phone: "",
          userId: "",
          departmentId: "",
          jobPositionId: "",
          hireDate: "",
          employmentType: "full_time",
          workLocation: "",
          wage: "0",
          wageType: "monthly",
          currencyCode: "IDR",
        },
  });

  function handleSubmit(values: EmployeeFormValues) {
    if (isEdit && initial) {
      return getSwantaraService()
        .employees.update(Number(orgId) || 0, initial.id, {
          name: values.name,
          employeeNumber: values.employeeNumber,
          userId: values.userId === "" ? null : Number(values.userId) || null,
          departmentId: values.departmentId === "" ? null : Number(values.departmentId) || null,
          jobPositionId: values.jobPositionId === "" ? null : Number(values.jobPositionId) || null,
          managerId: initial.managerId,
          hireDate: values.hireDate || null,
          terminationDate: initial.terminationDate,
          employmentType: values.employmentType,
          workLocation: values.workLocation || null,
          active: initial.active,
        })
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    } else {
      return getSwantaraService()
        .employees.create(Number(orgId) || 0, {
          organizationId: Number(orgId) || 0,
          name: values.name,
          email: values.email || null,
          phone: values.phone || null,
          userId: values.userId === "" ? null : Number(values.userId) || null,
          employeeNumber: values.employeeNumber,
          departmentId: values.departmentId === "" ? null : Number(values.departmentId) || null,
          jobPositionId: values.jobPositionId === "" ? null : Number(values.jobPositionId) || null,
          managerId: null,
          hireDate: values.hireDate || null,
          employmentType: values.employmentType,
          workLocation: values.workLocation || null,
          wage: Number(values.wage) || 0,
          wageType: values.wageType,
          currencyCode: values.currencyCode,
        })
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    }
  }

  function employmentTypeLabel(key: string): string {
    try {
      return (t as unknown as (k: string) => string)(`employmentType.${key}`);
    } catch {
      return humanizeKey(String(key));
    }
  }

  function wageTypeLabel(key: string): string {
    try {
      return (t as unknown as (k: string) => string)(`wageType.${key}`);
    } catch {
      return humanizeKey(String(key));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{isEdit ? t("editEmployee") : t("addEmployee")}</DialogTitle>
          <DialogDescription>{t("subtitle")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{t("formPersonalInfo")}</legend>
            <FormField name="name" label={t("fieldName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
            </FormField>
            <FormField name="employeeNumber" label={t("fieldEmployeeNumber")}>
              {({ field, id }) => (
                <Input {...field} id={id} placeholder={t("fieldEmployeeNumber")} />
              )}
            </FormField>
            <FormField name="email" label={t("fieldEmail")}>
              {({ field, id }) => (
                <Input {...field} id={id} type="email" placeholder={t("fieldEmail")} />
              )}
            </FormField>
            <FormField name="phone" label={t("fieldPhone")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldPhone")} />}
            </FormField>
          </fieldset>

          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{t("formEmployment")}</legend>
            <FormField name="departmentId" label={t("fieldDepartment")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldDepartment")}>
                    <SelectValue placeholder={t("selectDepartment")} />
                  </SelectTrigger>
                  <SelectContent>
                    {departments.map((dept) => (
                      <SelectItem key={dept.id} value={String(dept.id)}>
                        {dept.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="jobPositionId" label={t("fieldJobPosition")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldJobPosition")}>
                    <SelectValue placeholder={t("selectJobPosition")} />
                  </SelectTrigger>
                  <SelectContent>
                    {jobPositions.map((pos) => (
                      <SelectItem key={pos.id} value={String(pos.id)}>
                        {pos.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="hireDate" label={t("fieldHireDate")}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
            <FormField name="employmentType" label={t("fieldEmploymentType")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldEmploymentType")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {employmentTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {employmentTypeLabel(String(type))}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="workLocation" label={t("fieldWorkLocation")} className="sm:col-span-2">
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldWorkLocation")} />}
            </FormField>
          </fieldset>

          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{t("formCompensation")}</legend>
            <FormField name="wage" label={t("fieldWage")}>
              {({ field, id }) => (
                <CurrencyField
                  id={id}
                  value={field.value === "" ? null : Number(field.value)}
                  onValueChange={(v) => field.onChange(String(v ?? 0))}
                  currency={form.watch("currencyCode") || "IDR"}
                />
              )}
            </FormField>
            <FormField name="wageType" label={t("fieldWageType")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldWageType")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {wageTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {wageTypeLabel(String(type))}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="currencyCode" label={t("fieldCurrency")}>
              {({ field, id }) => (
                <Combobox
                  items={currencyOptions}
                  value={field.value}
                  onValueChange={(val) => field.onChange(val ?? "")}
                >
                  <ComboboxInput id={id} placeholder={t("fieldCurrency")} />
                  <ComboboxContent>
                    <ComboboxList>
                      {(code: string) => (
                        <ComboboxItem key={code} value={code}>
                          <span className="font-mono">{code}</span>
                        </ComboboxItem>
                      )}
                    </ComboboxList>
                    <ComboboxEmpty>{tCommon("noData")}</ComboboxEmpty>
                  </ComboboxContent>
                </Combobox>
              )}
            </FormField>
          </fieldset>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
