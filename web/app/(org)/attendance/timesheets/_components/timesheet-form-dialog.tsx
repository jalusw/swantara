"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useForm, useWatch } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Dimension, Employee, Project } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useTimesheetFormSchema() {
  const t = useTranslations("Attendance");
  return z.object({
    employeeId: z.coerce.number().min(1, t("validationEmployeeRequired")),
    date: z.string().min(1, t("validationDateRequired")),
    hours: z.coerce.number().min(0.5, t("validationMinHours")),
    description: z.string().nullable(),
    dimensionId: z.coerce.number().nullable(),
    projectId: z.coerce.number().nullable(),
    taskId: z.coerce.number().nullable(),
  });
}
type TimesheetFormValues = z.infer<ReturnType<typeof useTimesheetFormSchema>>;

export function TimesheetFormDialog({
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
  const t = useTranslations("Attendance");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();

  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const dimensionsQuery = useOrgListQuery<{ accounts: Dimension[] }, Record<string, never>>(
    "dimensions",
    (organizationId) => getSwantaraService().dimensions.list(organizationId),
  );

  const projectsQuery = useOrgListQuery<{ projects: Project[] }, Record<string, never>>(
    "projects",
    (organizationId) => getSwantaraService().projects.list(organizationId),
  );

  const contactMap = new Map(
    (contactsQuery.data?.contacts ?? []).map((p) => [p.id, p.displayName ?? p.name]),
  );

  const employeeOptions = (employeesQuery.data?.employees ?? []).map((e) => ({
    id: String(e.id),
    name: contactMap.get(e.contactId) ?? e.employeeNumber,
  }));

  const dimensionOptions = (dimensionsQuery.data?.accounts ?? []).map((a) => ({
    id: String(a.id),
    name: a.name,
  }));

  const projectOptions = (projectsQuery.data?.projects ?? []).map((p) => ({
    id: String(p.id),
    name: p.name,
  }));

  const schema = useTimesheetFormSchema();

  const form = useForm<TimesheetFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      employeeId: 0,
      date: getLocalDateString(),
      hours: 8,
      description: null,
      dimensionId: null,
      projectId: null,
      taskId: null,
    },
  });

  const selectedProjectId = useWatch({
    control: form.control,
    name: "projectId",
  });

  const tasksQuery = useQuery({
    queryKey: ["projectTasks", orgId, selectedProjectId],
    queryFn: () =>
      getSwantaraService().projects.tasks.list(Number(orgId), selectedProjectId as number),
    enabled: Boolean(selectedProjectId),
  });

  const taskOptions = (tasksQuery.data?.tasks ?? []).map((task) => ({
    id: String(task.id),
    name: task.name,
  }));

  const createMutation = useMutation({
    mutationFn: (values: TimesheetFormValues) =>
      getSwantaraService().timesheets.create(Number(orgId), {
        employeeId: values.employeeId,
        date: values.date,
        projectId: values.projectId || null,
        taskId: values.taskId || null,
        dimensionId: values.dimensionId,
        hours: values.hours,
        description: values.description,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["timesheets", Number(orgId)] });
      onSave();
      toast.success(t("timesheetCreated"));
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  function handleSubmit(values: TimesheetFormValues) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newTimesheetEntry")}
      description={t("timesheetsSubtitle")}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="employeeId" label={t("fieldEmployee")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(Number(value))}
            >
              <SelectTrigger id={id} aria-label={t("fieldEmployee")}>
                <SelectValue placeholder={t("fieldEmployee")} />
              </SelectTrigger>
              <SelectContent>
                {employeeOptions.map((employee) => (
                  <SelectItem key={employee.id} value={employee.id}>
                    {employee.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="projectId" label={t("fieldProject")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => {
                field.onChange(value ? Number(value) : null);
                form.setValue("taskId", null);
              }}
            >
              <SelectTrigger id={id} aria-label={t("fieldProject")}>
                <SelectValue placeholder={t("noProject")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {projectOptions.map((project) => (
                  <SelectItem key={project.id} value={project.id}>
                    {project.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="taskId" label={t("fieldTask")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
              disabled={!selectedProjectId}
            >
              <SelectTrigger id={id} aria-label={t("fieldTask")}>
                <SelectValue placeholder={t("noTask")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {taskOptions.map((task) => (
                  <SelectItem key={task.id} value={task.id}>
                    {task.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="date" label={t("fieldDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="hours" label={t("fieldHours")}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0.5" step="0.5" inputMode="decimal" />
          )}
        </FormField>
        <FormField name="description" label={t("fieldDescription")}>
          {({ field, id }) => (
            <Textarea
              id={id}
              value={field.value ?? ""}
              onChange={(event) => field.onChange(event.target.value || null)}
            />
          )}
        </FormField>
        <FormField name="dimensionId" label={t("fieldDimension")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("fieldDimension")}>
                <SelectValue placeholder={t("fieldDimension")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {dimensionOptions.map((account) => (
                  <SelectItem key={account.id} value={account.id}>
                    {account.name}
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
