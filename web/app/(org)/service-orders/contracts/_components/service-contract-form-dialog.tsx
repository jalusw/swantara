"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Equipment, Subscription } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useServiceContractSchema() {
  const t = useTranslations("Service");
  return z.object({
    name: z.string().min(1, t("validation_nameRequired")),
    contactId: z.string().optional(),
    equipmentId: z.string().optional(),
    subscriptionId: z.string().optional(),
    coverage: z.string().optional(),
    slaResponseHours: z.string().optional(),
    dateStart: z.string().optional(),
    dateEnd: z.string().optional(),
  });
}
type ServiceContractValues = z.infer<ReturnType<typeof useServiceContractSchema>>;

export function ServiceContractFormDialog({
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

  const equipmentsQuery = useOrgListQuery<{ equipments: Equipment[] }, Record<string, never>>(
    "equipments",
    (organizationId) => getSwantaraService().equipments.list(organizationId),
  );
  const equipments = equipmentsQuery.data?.equipments ?? [];

  const subscriptionsQuery = useOrgListQuery<
    { subscriptions: Subscription[] },
    Record<string, never>
  >("subscriptions", (organizationId) => getSwantaraService().subscriptions.list(organizationId));
  const subscriptions = subscriptionsQuery.data?.subscriptions ?? [];

  const schema = useServiceContractSchema();

  const form = useForm<ServiceContractValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      contactId: "",
      equipmentId: "",
      subscriptionId: "",
      coverage: "",
      slaResponseHours: "",
      dateStart: "",
      dateEnd: "",
    },
  });

  const createMutation = useMutation({
    mutationFn: (values: ServiceContractValues) =>
      getSwantaraService().serviceContracts.create(Number(orgId), {
        name: values.name,
        contactId: values.contactId ? Number(values.contactId) : undefined,
        equipmentId: values.equipmentId ? Number(values.equipmentId) : undefined,
        subscriptionId: values.subscriptionId ? Number(values.subscriptionId) : undefined,
        coverage: values.coverage || undefined,
        slaResponseHours: values.slaResponseHours ? Number(values.slaResponseHours) : undefined,
        dateStart: values.dateStart ? new Date(values.dateStart) : undefined,
        dateEnd: values.dateEnd ? new Date(values.dateEnd) : undefined,
      }),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["serviceContracts", Number(orgId)] });
      onSave(String(result.serviceContract.id));
      toast.success(t("contractCreated"));
    },
    onError: () => {
      toast.error(t("createFailed"));
    },
  });

  function handleSubmit(values: ServiceContractValues) {
    createMutation.mutate(values);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("newContract")}</DialogTitle>
          <DialogDescription>{t("contractsDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("fieldName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
              </FormField>
              <FormField name="coverage" label={t("coverage")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("coverage")} />}
              </FormField>
              <FormField name="contactId" label={t("customer")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("customer")}>
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
              <FormField name="subscriptionId" label={t("subscription")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("subscription")}>
                      <SelectValue placeholder={t("selectSubscription")} />
                    </SelectTrigger>
                    <SelectContent>
                      {subscriptions.map((s) => (
                        <SelectItem key={s.id} value={String(s.id)}>
                          {s.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="slaResponseHours" label={t("slaResponseHours")}>
                {({ field, id }) => (
                  <Input
                    {...field}
                    id={id}
                    type="number"
                    min="0"
                    placeholder={t("slaResponseHours")}
                  />
                )}
              </FormField>
              <FormField name="dateStart" label={t("startDate")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={t("endDate")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton loading={createMutation.isPending}>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
