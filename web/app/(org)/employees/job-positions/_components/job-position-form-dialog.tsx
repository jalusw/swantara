"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect } from "react";
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
  const t = useTranslations("Employees");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    departmentId: z.string(),
  });
}

function toDefaultValues(initial?: JobPosition | null) {
  return initial
    ? {
        name: initial.name,
        departmentId: initial.departmentId ? String(initial.departmentId) : "",
      }
    : {
        name: "",
        departmentId: "",
      };
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
  const t = useTranslations("Employees");
  const tCommon = useTranslations("Common");

  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );

  const departments = departmentsQuery.data?.departments ?? [];

  const schema = useJobPositionFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: toDefaultValues(initial),
  });

  useEffect(() => {
    if (open) {
      form.reset(toDefaultValues(initial));
    }
  }, [open, initial, form]);

  function handleSubmit(values: Values) {
    const payload = {
      organizationId: Number(orgId),
      name: values.name,
      departmentId: values.departmentId ? Number(values.departmentId) : null,
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .jobPositions.update(Number(orgId), initial.id, payload)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    } else {
      return getSwantaraService()
        .jobPositions.create(Number(orgId), payload)
        .then(() => onSave())
        .catch(() => void toast.error(t("saveFailed")));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editPosition") : t("newPosition")}
      description={t("jobPositionsSubtitle")}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      form={form}
      onSubmit={handleSubmit}
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="departmentId" label={t("fieldDepartment")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldDepartment")}>
                <SelectValue placeholder={t("fieldDepartment")} />
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
