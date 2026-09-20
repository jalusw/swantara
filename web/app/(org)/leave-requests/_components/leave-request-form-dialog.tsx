"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Employee, LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

const leaveRequestFormSchema = z
  .object({
    employeeId: z.string().min(1),
    leaveTypeId: z.string().min(1),
    dateFrom: z.string().min(1),
    dateTo: z.string().min(1),
    days: z.string().refine((v) => Number(v) > 0),
  })
  .refine((data) => data.dateTo >= data.dateFrom, {
    message: "End date must be after start date",
    path: ["dateTo"],
  });
type LeaveRequestFormValues = z.infer<typeof leaveRequestFormSchema>;

export function LeaveRequestFormDialog({
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
  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );
  const leaveTypesQuery = useOrgListQuery<{ leaveTypes: LeaveType[] }, Record<string, never>>(
    "leaveTypes",
    (organizationId) => getSwantaraService().leaveTypes.list(organizationId),
  );

  const employees = employeesQuery.data?.employees ?? [];
  const leaveTypes = leaveTypesQuery.data?.leaveTypes ?? [];

  const form = useForm<LeaveRequestFormValues>({
    resolver: zodResolver(leaveRequestFormSchema),
    defaultValues: {
      employeeId: "",
      leaveTypeId: "",
      dateFrom: getLocalDateString(),
      dateTo: getLocalDateString(),
      days: "1",
    },
  });

  function handleSubmit(values: LeaveRequestFormValues) {
    void getSwantaraService()
      .leaveRequests.create(Number(orgId), {
        employeeId: Number(values.employeeId),
        leaveTypeId: Number(values.leaveTypeId),
        dateFrom: values.dateFrom,
        dateTo: values.dateTo,
        days: Number(values.days),
      })
      .then(() => onSave())
      .catch(() => toast.error("Something went wrong."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Add leave request"}
      description={"Create a new leave request."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="employeeId" label={"Employee"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Employee"}>
                <SelectValue placeholder={"Select employee"} />
              </SelectTrigger>
              <SelectContent>
                {employees.map((e) => (
                  <SelectItem key={e.id} value={String(e.id)}>
                    {e.employeeNumber}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="leaveTypeId" label={"Leave type"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Leave type"}>
                <SelectValue placeholder={"Select leave type"} />
              </SelectTrigger>
              <SelectContent>
                {leaveTypes.map((lt) => (
                  <SelectItem key={lt.id} value={String(lt.id)}>
                    {lt.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="dateFrom" label={"From"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="dateTo" label={"To"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="days" label={"Days"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0.5" step="0.5" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
