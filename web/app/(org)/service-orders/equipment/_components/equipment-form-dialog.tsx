"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Service");
  return z.object({
    name: z.string().min(1, t("validation_nameRequired")),
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
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
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
      toast.success(t("equipmentCreated"));
    },
    onError: () => {
      toast.error(t("createFailed"));
    },
  });

  function handleSubmit(values: Values) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newEquipment")}
      description={t("equipmentFormDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={t("fieldName")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
          </FormField>
          <FormField name="category" label={t("category")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("category")} />}
          </FormField>
          <FormField name="itemId" label={t("item")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("item")}>
                  <SelectValue placeholder={t("selectItem")} />
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
          <FormField name="ownerContactId" label={t("owner")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("owner")}>
                  <SelectValue placeholder={t("selectCustomer")} />
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
          <FormField name="location" label={t("location")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("location")} />}
          </FormField>
          <FormField name="installDate" label={t("installDate")}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
          <FormField name="warrantyEnd" label={t("warrantyEnd")}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
        </div>
      </div>
    </EntityFormDialog>
  );
}
