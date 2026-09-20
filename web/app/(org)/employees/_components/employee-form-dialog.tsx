"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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

const employeeFormSchema = z.object({
  name: z.string().trim().min(1),
  employeeNumber: z.string().trim().min(1),
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
    .refine((value) => Number(value) >= 0),
  wageType: z.enum(wageTypes),
  currencyCode: z.string().trim(),
});
type EmployeeFormValues = z.infer<typeof employeeFormSchema>;

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

  const form = useForm<EmployeeFormValues>({
    resolver: zodResolver(employeeFormSchema),
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
      void getSwantaraService()
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
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
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
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit employee" : "Add employee"}</DialogTitle>
          <DialogDescription>
            {"Manage your team, roles, and departments across the organization."}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{"Personal information"}</legend>
            <FormField name="name" label={"Name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
            </FormField>
            <FormField name="employeeNumber" label={"Employee number"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Employee number"} />}
            </FormField>
            <FormField name="email" label={"Email"}>
              {({ field, id }) => <Input {...field} id={id} type="email" placeholder={"Email"} />}
            </FormField>
            <FormField name="phone" label={"Phone"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Phone"} />}
            </FormField>
          </fieldset>

          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{"Employment Details"}</legend>
            <FormField name="departmentId" label={"Department"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Department"}>
                    <SelectValue placeholder={"Select department"} />
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
            <FormField name="jobPositionId" label={"Job position"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Job position"}>
                    <SelectValue placeholder={"Select job position"} />
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
            <FormField name="hireDate" label={"Hire date"}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
            <FormField name="employmentType" label={"Employment type"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Employment type"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {employmentTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {humanizeKey(String(type))}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="workLocation" label={"Location"} className="sm:col-span-2">
              {({ field, id }) => <Input {...field} id={id} placeholder={"Location"} />}
            </FormField>
          </fieldset>

          <fieldset className="grid gap-4 rounded-lg border p-4 sm:grid-cols-2">
            <legend className="px-1 text-sm">{"Compensation"}</legend>
            <FormField name="wage" label={"Wage"}>
              {({ field, id }) => (
                <CurrencyField
                  id={id}
                  value={field.value === "" ? null : Number(field.value)}
                  onValueChange={(v) => field.onChange(String(v ?? 0))}
                  currency={form.watch("currencyCode") || "IDR"}
                />
              )}
            </FormField>
            <FormField name="wageType" label={"Wage type"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Wage type"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {wageTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {humanizeKey(String(type))}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="currencyCode" label={"Currency"}>
              {({ field, id }) => (
                <Combobox
                  items={currencyOptions}
                  value={field.value}
                  onValueChange={(val) => field.onChange(val ?? "")}
                >
                  <ComboboxInput id={id} placeholder={"Currency"} />
                  <ComboboxContent>
                    <ComboboxList>
                      {(code: string) => (
                        <ComboboxItem key={code} value={code}>
                          <span className="font-mono">{code}</span>
                        </ComboboxItem>
                      )}
                    </ComboboxList>
                    <ComboboxEmpty>{"No results"}</ComboboxEmpty>
                  </ComboboxContent>
                </Combobox>
              )}
            </FormField>
          </fieldset>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Save"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
