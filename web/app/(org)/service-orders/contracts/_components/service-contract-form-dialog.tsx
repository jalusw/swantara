"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
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
  return z.object({
    name: z.string().min(1, "Name is required"),
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
      toast.success("Contract created");
    },
    onError: () => {
      toast.error("Failed to create contract");
    },
  });

  function handleSubmit(values: ServiceContractValues) {
    createMutation.mutate(values);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{"Create service contract"}</DialogTitle>
          <DialogDescription>{"Manage service contracts and SLA agreements."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="coverage" label={"Coverage"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Coverage"} />}
              </FormField>
              <FormField name="contactId" label={"Customer"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Customer"}>
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
              <FormField name="equipmentId" label={"Equipment"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Equipment"}>
                      <SelectValue placeholder={"Select equipment"} />
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
              <FormField name="subscriptionId" label={"Subscriptions"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Subscriptions"}>
                      <SelectValue placeholder={"Select Subscription"} />
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
              <FormField name="slaResponseHours" label={"SLA response hours"}>
                {({ field, id }) => (
                  <Input
                    {...field}
                    id={id}
                    type="number"
                    min="0"
                    placeholder={"SLA response hours"}
                  />
                )}
              </FormField>
              <FormField name="dateStart" label={"Start date"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={"End date"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Save"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
