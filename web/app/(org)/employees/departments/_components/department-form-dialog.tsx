"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Textarea } from "@/components/textarea";
import type { Department } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useDepartmentFormSchema() {
  const t = useTranslations("Employees");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    description: z.string().optional(),
  });
}

export function DepartmentFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Department | null;
  onSave: (id: string) => void;
}) {
  const isEdit = Boolean(initial);
  const t = useTranslations("Employees");

  const schema = useDepartmentFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          description: initial.description ?? "",
        }
      : {
          name: "",
          description: "",
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      organizationId: Number(orgId),
      name: values.name,
      description: values.description || null,
      parentId: initial?.parentId ?? null,
      managerId: initial?.managerId ?? null,
      dimensionId: initial?.dimensionId ?? null,
    };

    if (isEdit && initial) {
      return getSwantaraService()
        .departments.update(Number(orgId), initial.id, request)
        .then((result) => {
          toast.success(t("updated"));
          onSave(String(result.department.id));
        })
        .catch(() => {
          toast.error(t("saveFailed"));
        });
    } else {
      return getSwantaraService()
        .departments.create(Number(orgId), request)
        .then((result) => {
          toast.success(t("created"));
          onSave(String(result.department.id));
        })
        .catch(() => {
          toast.error(t("saveFailed"));
        });
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editDepartment") : t("newDepartment")}
      description={t("departmentsSubtitle")}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="description" label={t("fieldDescription")}>
          {({ field, id }) => <Textarea {...field} id={id} placeholder={t("fieldDescription")} />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
