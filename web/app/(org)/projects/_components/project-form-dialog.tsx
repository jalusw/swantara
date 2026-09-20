"use client";

import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Dimension, Project, SaleOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useProjectFormSchema() {
  return z
    .object({
      name: z.string().min(1, "Project name is required."),
      contactId: z.coerce.number().min(1, "Select a customer."),
      managerId: z.coerce.number().nullable(),
      billingType: z.enum(["fixed", "time_material", "milestone"]),
      billableRate: z.coerce.number().min(0, "Billable rate is required."),
      dimensionId: z.coerce.number().nullable(),
      saleOrderId: z.coerce.number().nullable(),
      dateStart: z.string().nullable(),
      dateEnd: z.string().nullable(),
    })
    .refine((data) => !data.dateStart || !data.dateEnd || data.dateEnd >= data.dateStart, {
      message: "End date must be after start date",
      path: ["dateEnd"],
    });
}

export function ProjectFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: Project | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const dimensionQuery = useOrgListQuery<{ accounts: Dimension[] }, Record<string, never>>(
    "dimensions",
    (organizationId) => getSwantaraService().dimensions.list(organizationId),
  );

  const saleOrdersQuery = useOrgListQuery<{ orders: SaleOrder[] }, Record<string, never>>(
    "saleOrders",
    (organizationId) => getSwantaraService().saleOrders.list(organizationId),
  );

  const contactOptions = (contactsQuery.data?.contacts ?? []).map((p) => ({
    id: String(p.id),
    name: p.displayName ?? p.name,
  }));

  const dimensionOptions = (dimensionQuery.data?.accounts ?? []).map((a) => ({
    id: String(a.id),
    name: a.name,
  }));

  const saleOrderOptions = (saleOrdersQuery.data?.orders ?? []).map((so) => ({
    id: String(so.id),
    name: so.name,
  }));

  const schema = useProjectFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initial?.name ?? "",
      contactId: initial?.contactId ?? 0,
      managerId: initial?.managerId ?? null,
      billingType: initial?.billingType ?? "time_material",
      billableRate: initial?.billableRate ?? 0,
      dimensionId: initial?.dimensionId ?? null,
      saleOrderId: initial?.saleOrderId ?? null,
      dateStart: initial?.dateStart ? getLocalDateString(new Date(initial.dateStart)) : null,
      dateEnd: initial?.dateEnd ? getLocalDateString(new Date(initial.dateEnd)) : null,
    },
  });

  function handleSubmit(values: Values) {
    const payload = {
      organizationId: Number(orgId),
      name: values.name,
      contactId: values.contactId,
      managerId: values.managerId || null,
      dimensionId: values.dimensionId || null,
      saleOrderId: values.saleOrderId || null,
      billingType: values.billingType,
      billableRate: values.billableRate,
      dateStart: values.dateStart ? new Date(values.dateStart) : null,
      dateEnd: values.dateEnd ? new Date(values.dateEnd) : null,
    };

    const op = isEdit
      ? getSwantaraService().projects.update(Number(orgId), initial!.id, {
          name: payload.name,
          contactId: payload.contactId,
          managerId: payload.managerId,
          dimensionId: payload.dimensionId,
          saleOrderId: payload.saleOrderId,
          billingType: payload.billingType,
          billableRate: payload.billableRate,
          dateStart: payload.dateStart,
          dateEnd: payload.dateEnd,
        })
      : getSwantaraService().projects.create(Number(orgId), payload);

    void op.then(() => onSave()).catch(() => toast.error("Something went wrong."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit project" : "New project"}
      description={"Create a project with billing type, customer and dimension account."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={"Project name"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="contactId" label={"Customer"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(Number(value))}
            >
              <SelectTrigger id={id} aria-label={"Customer"}>
                <SelectValue placeholder={"Select customer"} />
              </SelectTrigger>
              <SelectContent>
                {contactOptions.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="managerId" label={"Manager"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Manager"}>
                <SelectValue placeholder={"Select manager"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {contactOptions.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="billingType" label={"Billing type"}>
          {({ field, id }) => (
            <Select
              value={field.value ?? ""}
              onValueChange={(value) =>
                field.onChange(value as "fixed" | "time_material" | "milestone")
              }
            >
              <SelectTrigger id={id} aria-label={"Billing type"}>
                <SelectValue placeholder={"Billing type"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fixed">{"Fixed"}</SelectItem>
                <SelectItem value="time_material">{"Time & Material"}</SelectItem>
                <SelectItem value="milestone">{"Milestone"}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="billableRate" label={"Billable rate"}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0" step="0.01" inputMode="decimal" />
          )}
        </FormField>
        <FormField name="dimensionId" label={"Dimension account"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Dimension account"}>
                <SelectValue placeholder={"Select dimension account"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {dimensionOptions.map((a) => (
                  <SelectItem key={a.id} value={a.id}>
                    {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="saleOrderId" label={"Sale order"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Sale order"}>
                <SelectValue placeholder={"Select sale order"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {saleOrderOptions.map((so) => (
                  <SelectItem key={so.id} value={so.id}>
                    {so.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="dateStart" label={"Start date"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="date"
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value || null)}
            />
          )}
        </FormField>
        <FormField name="dateEnd" label={"End date"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="date"
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value || null)}
            />
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
