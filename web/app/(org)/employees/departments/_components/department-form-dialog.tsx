"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
  return z.object({
    name: z.string().min(1, "Enter a name."),
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
      void getSwantaraService()
        .departments.update(Number(orgId), initial.id, request)
        .then((result) => {
          toast.success("Department updated");
          onSave(String(result.department.id));
        })
        .catch(() => {
          toast.error("Failed to save department.");
        });
    } else {
      void getSwantaraService()
        .departments.create(Number(orgId), request)
        .then((result) => {
          toast.success("Department created");
          onSave(String(result.department.id));
        })
        .catch(() => {
          toast.error("Failed to save department.");
        });
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit department" : "Add department"}
      description={"Create a new department."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="description" label={"Description"}>
          {({ field, id }) => <Textarea {...field} id={id} placeholder={"Description"} />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
