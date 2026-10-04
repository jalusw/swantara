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
import type { Equipment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useMaintenancePlanFormSchema() {
  const t = useTranslations("Service");
  return z.object({
    name: z.string().min(1, t("validation_nameRequired")),
    equipmentId: z.string().min(1, t("validation_equipmentRequired")),
    intervalDays: z.string().min(1, t("validation_intervalRequired")),
    nextDue: z.string().min(1, t("validation_nextDueRequired")),
  });
}

export function MaintenancePlanFormDialog({
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

  const equipmentsQuery = useOrgListQuery<{ equipments: Equipment[] }, Record<string, never>>(
    "equipments",
    (organizationId) => getSwantaraService().equipments.list(organizationId),
  );
  const equipments = equipmentsQuery.data?.equipments ?? [];

  const schema = useMaintenancePlanFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      equipmentId: "",
      intervalDays: "",
      nextDue: "",
    },
  });

  const createMutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().maintenancePlans.create(Number(orgId), {
        name: values.name,
        equipmentId: Number(values.equipmentId),
        intervalDays: Number(values.intervalDays),
        nextDue: values.nextDue,
      }),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["maintenancePlans", Number(orgId)] });
      onSave(String(result.maintenancePlan.id));
      toast.success(t("planCreated"));
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
      title={t("newMaintenancePlan")}
      description={t("plansDescription")}
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
          <FormField name="equipmentId" label={t("equipment")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("equipment")}>
                  <SelectValue placeholder={t("selectEquipment")} />
                </SelectTrigger>
                <SelectContent>
                  {equipments.map((e) => (
                    <SelectItem key={e.id} value={String(e.id)}>
                      {e.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="intervalDays" label={t("intervalDays")}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="1" placeholder={t("intervalDays")} />
            )}
          </FormField>
          <FormField name="nextDue" label={t("nextDueDate")}>
            {({ field, id }) => <Input {...field} id={id} type="date" />}
          </FormField>
        </div>
      </div>
    </EntityFormDialog>
  );
}
