"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import type { Warehouse } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useWarehouseFormSchema() {
  return z.object({
    name: z.string().min(1, "Warehouse name is required."),
    code: z.string(),
    line1: z.string(),
    line2: z.string(),
    city: z.string(),
    state: z.string(),
    postalCode: z.string(),
    countryCode: z.string(),
  });
}

export function WarehouseFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Warehouse | null;
  onSave: (id: string) => void;
}) {
  const isEdit = Boolean(initial);

  const schema = useWarehouseFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          code: initial.code || "",
          line1: initial.line1 || "",
          line2: initial.line2 || "",
          city: initial.city || "",
          state: initial.state || "",
          postalCode: initial.postalCode || "",
          countryCode: initial.countryCode || "",
        }
      : {
          name: "",
          code: "",
          line1: "",
          line2: "",
          city: "",
          state: "",
          postalCode: "",
          countryCode: "",
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      organizationId: Number(orgId),
      name: values.name,
      code: values.code || null,
      line1: values.line1 || null,
      line2: values.line2 || null,
      city: values.city || null,
      state: values.state || null,
      postalCode: values.postalCode || null,
      countryCode: values.countryCode || null,
    };

    if (isEdit && initial) {
      void getSwantaraService()
        .inventory.updateWarehouse(Number(orgId), initial.id, request)
        .then((result) => onSave(String(result.warehouse.id)))
        .catch(() => toast.error("Failed to save warehouse."));
    } else {
      void getSwantaraService()
        .inventory.createWarehouse(Number(orgId), request)
        .then((result) => onSave(String(result.warehouse.id)))
        .catch(() => toast.error("Failed to save warehouse."));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit warehouse" : "New warehouse"}
      description={"Manage warehouse locations and their usage types."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="code" label={"Code"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Code"} />}
        </FormField>
        <FormField name="line1" label={"Address line 1"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Address line 1"} />}
        </FormField>
        <FormField name="line2" label={"Address line 2"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Address line 2"} />}
        </FormField>
        <FormField name="city" label={"City"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"City"} />}
        </FormField>
        <FormField name="state" label={"State"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"State"} />}
        </FormField>
        <FormField name="postalCode" label={"Postal code"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Postal code"} />}
        </FormField>
        <FormField name="countryCode" label={"Country code"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Country code"} />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
