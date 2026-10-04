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
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Employee, LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

function useLeaveRequestFormSchema() {
  const t = useTranslations("Leave");
  return z
    .object({
      employeeId: z.string().min(1, t("validationEmployeeRequired")),
      leaveTypeId: z.string().min(1, t("validationLeaveTypeRequired")),
      dateFrom: z.string().min(1, t("validationDateRequired")),
      dateTo: z.string().min(1, t("validationDateRequired")),
      days: z.string().refine((v) => Number(v) > 0, t("validationDaysInvalid")),
    })
    .refine((data) => data.dateTo >= data.dateFrom, {
      message: t("validationEndAfterStart"),
      path: ["dateTo"],
    });
}
type LeaveRequestFormValues = z.infer<ReturnType<typeof useLeaveRequestFormSchema>>;

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
  const t = useTranslations("Leave");
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
    resolver: zodResolver(useLeaveRequestFormSchema()),
    defaultValues: {
      employeeId: "",
      leaveTypeId: "",
      dateFrom: getLocalDateString(),
      dateTo: getLocalDateString(),
      days: "1",
    },
  });

  function handleSubmit(values: LeaveRequestFormValues) {
    return getSwantaraService()
      .leaveRequests.create(Number(orgId), {
        employeeId: Number(values.employeeId),
        leaveTypeId: Number(values.leaveTypeId),
        dateFrom: values.dateFrom,
        dateTo: values.dateTo,
        days: Number(values.days),
      })
      .then(() => onSave())
      .catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newRequest")}
      description={t("newRequestDescription")}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="employeeId" label={t("fieldEmployee")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldEmployee")}>
                <SelectValue placeholder={t("selectEmployee")} />
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
        <FormField name="leaveTypeId" label={t("fieldLeaveType")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldLeaveType")}>
                <SelectValue placeholder={t("selectLeaveType")} />
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
        <FormField name="dateFrom" label={t("fieldDateFrom")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="dateTo" label={t("fieldDateTo")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="days" label={t("fieldDays")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0.5" step="0.5" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
