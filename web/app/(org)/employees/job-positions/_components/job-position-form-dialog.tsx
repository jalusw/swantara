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
import type { Department, JobPosition } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useJobPositionFormSchema() {
  return z.object({
    name: z.string().min(1, "Enter a name."),
    departmentId: z.string(),
  });
}

export function JobPositionFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: JobPosition | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const departments = departmentsQuery.data?.departments ?? [];

  const schema = useJobPositionFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          departmentId: initial.departmentId ? String(initial.departmentId) : "",
        }
      : {
          name: "",
          departmentId: "",
        },
  });

  function handleSubmit(values: Values) {
    const payload = {
      organizationId: Number(orgId),
      name: values.name,
      departmentId: values.departmentId ? Number(values.departmentId) : null,
    };

    if (isEdit && initial) {
      void getSwantaraService()
        .jobPositions.update(Number(orgId), initial.id, payload)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .jobPositions.create(Number(orgId), payload)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit position" : "New position"}
      description={"Manage job positions across departments."}
      form={form}
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="departmentId" label={"Department"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Department"}>
                <SelectValue placeholder={"Department"} />
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
      </div>
    </EntityFormDialog>
  );
}
