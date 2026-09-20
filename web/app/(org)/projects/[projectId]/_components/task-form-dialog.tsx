"use client";

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
  return z.object({
    name: z.string().min(1, "Task name is required."),
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

    void op.then(() => onSave()).catch(() => toast.error("Something went wrong."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Task updated." : "New task"}
      description={"Task"}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={"Task"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="assigneeId" label={"Assignee"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Assignee"}>
                <SelectValue placeholder={"Select assignee"} />
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
        <FormField name="stage" label={"Stage"}>
          {({ field, id }) => (
            <Select value={field.value ?? ""} onValueChange={(value) => field.onChange(value)}>
              <SelectTrigger id={id} aria-label={"Stage"}>
                <SelectValue placeholder={"Stage"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="backlog">{"Backlog"}</SelectItem>
                <SelectItem value="todo">{"To Do"}</SelectItem>
                <SelectItem value="in_progress">{"In Progress"}</SelectItem>
                <SelectItem value="done">{"Done"}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="plannedHours" label={"Planned hours"}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0" step="0.5" inputMode="decimal" />
          )}
        </FormField>
        {isEdit ? (
          <FormField name="effectiveHours" label={"Effective hours"}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="0" step="0.5" inputMode="decimal" />
            )}
          </FormField>
        ) : null}
        <FormField name="deadline" label={"Deadline"}>
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
        <FormField name="priority" label={"Priority"}>
          {({ field, id }) => (
            <Select
              value={field.value != null ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Priority"}>
                <SelectValue placeholder={"Priority"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                <SelectItem value="1">{"Low"}</SelectItem>
                <SelectItem value="2">{"Medium"}</SelectItem>
                <SelectItem value="3">{"High"}</SelectItem>
                <SelectItem value="4">{"Urgent"}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
