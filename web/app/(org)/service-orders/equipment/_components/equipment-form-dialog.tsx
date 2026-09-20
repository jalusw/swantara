"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useEquipmentFormSchema() {
  return z.object({
    name: z.string().min(1, "Name is required"),
    itemId: z.string().optional(),
    ownerContactId: z.string().optional(),
    location: z.string().optional(),
    category: z.string().optional(),
    installDate: z.string().optional(),
    warrantyEnd: z.string().optional(),
  });
}

export function EquipmentFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: (id: string) => void;
}) {
  const queryClient = useQueryClient();

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const contacts = contactsQuery.data?.contacts ?? [];

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );
  const products = productsQuery.data?.products ?? [];

  const schema = useEquipmentFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      itemId: "",
      ownerContactId: "",
      location: "",
      category: "",
      installDate: "",
      warrantyEnd: "",
    },
  });

  const createMutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().equipments.create(Number(orgId), {
        name: values.name,
        itemId: values.itemId ? Number(values.itemId) : undefined,
        ownerContactId: values.ownerContactId ? Number(values.ownerContactId) : undefined,
        location: values.location || undefined,
        category: values.category || undefined,
        installDate: values.installDate ? new Date(values.installDate) : undefined,
        warrantyEnd: values.warrantyEnd ? new Date(values.warrantyEnd) : undefined,
      }),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["equipments", Number(orgId)] });
      onSave(String(result.equipment.id));
      toast.success("Equipment created");
    },
    onError: () => {
      toast.error("Failed to create equipment");
    },
  });

  function handleSubmit(values: Values) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Create equipment"}
      description={"Manage equipment and asset tracking."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={"Name"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
          </FormField>
          <FormField name="category" label={"Category"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Category"} />}
          </FormField>
          <FormField name="itemId" label={"Item"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Item"}>
                  <SelectValue placeholder={"Select item"} />
                </SelectTrigger>
                <SelectContent>
                  {products.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="ownerContactId" label={"Owner"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Owner"}>
                  <SelectValue placeholder={"Select customer"} />
                </SelectTrigger>
                <SelectContent>
                  {contacts.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="location" label={"Location"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Location"} />}
          </FormField>
          <FormField name="installDate" label={"Install date"}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
          <FormField name="warrantyEnd" label={"Warranty end"}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
        </div>
      </div>
    </EntityFormDialog>
  );
}
