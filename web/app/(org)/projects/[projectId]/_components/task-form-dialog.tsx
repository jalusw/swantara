"use client";

import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, ProjectTask } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useTaskFormSchema() {
  const t = useTranslations("Projects");
  return z.object({
    name: z.string().min(1, t("validation_taskNameRequired")),
    assigneeId: z.coerce.number().nullable(),
    stage: z.enum(["backlog", "todo", "in_progress", "done"]),
    plannedHours: z.coerce.number().min(0),
    effectiveHours: z.coerce.number().min(0),
    deadline: z.string().nullable(),
    priority: z.coerce.number().nullable(),
    parentTaskId: z.coerce.number().nullable(),
  });
}
type TaskFormValues = z.infer<ReturnType<typeof useTaskFormSchema>>;

export function TaskFormDialog({
  open,
  onOpenChange,
  orgId,
  projectId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  projectId: number;
  initial?: ProjectTask | null;
  onSave: () => void;
}) {
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const dyn = (key: string) => (t as unknown as (k: string) => string)(key);
  const isEdit = Boolean(initial);

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contactOptions = (contactsQuery.data?.contacts ?? []).map((p) => ({
    id: String(p.id),
    name: p.displayName ?? p.name,
  }));

  const schema = useTaskFormSchema();

  const form = useForm<TaskFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initial?.name ?? "",
      assigneeId: initial?.assigneeId ?? null,
      stage: initial?.stage ?? "backlog",
      plannedHours: initial?.plannedHours ?? 0,
      effectiveHours: initial?.effectiveHours ?? 0,
      deadline: initial?.deadline ? getLocalDateString(new Date(initial.deadline)) : null,
      priority: initial?.priority ?? null,
      parentTaskId: initial?.parentTaskId ?? null,
    },
  });

  function handleSubmit(values: TaskFormValues) {
    const payload = {
      name: values.name,
      assigneeId: values.assigneeId || null,
      stage: values.stage,
      plannedHours: values.plannedHours,
      parentTaskId: values.parentTaskId || null,
      deadline: values.deadline ? new Date(values.deadline) : null,
      priority: values.priority,
    };

    const op = isEdit
      ? getSwantaraService().projects.tasks.update(Number(orgId), projectId, initial!.id, payload)
      : getSwantaraService().projects.tasks.create(Number(orgId), projectId, payload);

    return op.then(() => onSave()).catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editTask") : t("newTask")}
      description={t("taskFormDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={t("taskName")}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="assigneeId" label={t("assignee")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("assignee")}>
                <SelectValue placeholder={t("selectAssignee")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {contactOptions.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="stage" label={t("stage")}>
          {({ field, id }) => (
            <Select value={field.value ?? ""} onValueChange={(value) => field.onChange(value)}>
              <SelectTrigger id={id} aria-label={t("stage")}>
                <SelectValue placeholder={t("stage")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="backlog">{dyn("stage_backlog")}</SelectItem>
                <SelectItem value="todo">{dyn("stage_todo")}</SelectItem>
                <SelectItem value="in_progress">{dyn("stage_in_progress")}</SelectItem>
                <SelectItem value="done">{dyn("stage_done")}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="plannedHours" label={t("plannedHours")}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0" step="0.5" inputMode="decimal" />
          )}
        </FormField>
        {isEdit ? (
          <FormField name="effectiveHours" label={t("effectiveHours")}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="0" step="0.5" inputMode="decimal" />
            )}
          </FormField>
        ) : null}
        <FormField name="deadline" label={t("deadline")}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="date"
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value || null)}
            />
          )}
        </FormField>
        <FormField name="priority" label={t("priority")}>
          {({ field, id }) => (
            <Select
              value={field.value != null ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("priority")}>
                <SelectValue placeholder={t("priority")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                <SelectItem value="1">{dyn("priority_1")}</SelectItem>
                <SelectItem value="2">{dyn("priority_2")}</SelectItem>
                <SelectItem value="3">{dyn("priority_3")}</SelectItem>
                <SelectItem value="4">{dyn("priority_4")}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
