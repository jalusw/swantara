"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useId } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Employee, Equipment, Item, ServiceContract } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useServiceOrderSchema() {
  const lineSchema = z.object({
    type: z.string().min(1),
    itemId: z.string().optional(),
    description: z.string().optional(),
    qty: z.string().min(1),
    unitPrice: z.string().optional(),
    billable: z.boolean(),
  });
  const schema = z.object({
    name: z.string().min(1, "Name is required"),
    contactId: z.string().optional(),
    equipmentId: z.string().optional(),
    contractId: z.string().optional(),
    type: z.string().min(1, "Type is required"),
    priority: z.string().optional(),
    technicianId: z.string().optional(),
    reportedIssue: z.string().optional(),
    lines: z.array(lineSchema).min(1, "At least one line is required"),
  });
  return schema;
}
type ServiceOrderValues = z.infer<ReturnType<typeof useServiceOrderSchema>>;

export function ServiceOrderFormDialog({
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
  const linePrefix = useId();

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

  const contractsQuery = useOrgListQuery<
    { serviceContracts: ServiceContract[] },
    Record<string, never>
  >("serviceContracts", (organizationId) =>
    getSwantaraService().serviceContracts.list(organizationId),
  );
  const contracts = contractsQuery.data?.serviceContracts ?? [];

  const employeesQuery = useOrgListQuery<{ employees: Employee[] }, Record<string, never>>(
    "employees",
    (organizationId) => getSwantaraService().employees.list(organizationId),
  );
  const employees = employeesQuery.data?.employees ?? [];

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );
  const products = productsQuery.data?.products ?? [];

  const schema = useServiceOrderSchema();

  const form = useForm<ServiceOrderValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      contactId: "",
      equipmentId: "",
      contractId: "",
      type: "",
      priority: "1",
      technicianId: "",
      reportedIssue: "",
      lines: [
        {
          type: "labor",
          itemId: "",
          description: "",
          qty: "1",
          unitPrice: "0",
          billable: true,
        },
      ],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });

  function handleSubmit(values: ServiceOrderValues) {
    void toast.promise(
      getSwantaraService().serviceOrders.create(Number(orgId), {
        name: values.name,
        contactId: values.contactId ? Number(values.contactId) : undefined,
        equipmentId: values.equipmentId ? Number(values.equipmentId) : undefined,
        contractId: values.contractId ? Number(values.contractId) : undefined,
        type: values.type as "repair" | "maintenance" | "installation" | "inspection",
        priority: values.priority ? Number(values.priority) : undefined,
        technicianId: values.technicianId ? Number(values.technicianId) : undefined,
        reportedIssue: values.reportedIssue || undefined,
        lines: values.lines.map((l) => ({
          type: l.type as "part" | "labor" | "expense",
          itemId: l.itemId ? Number(l.itemId) : undefined,
          description: l.description || undefined,
          qty: Number(l.qty),
          unitPrice: l.unitPrice ? Number(l.unitPrice) : undefined,
          billable: l.billable,
        })),
      }),
      {
        loading: "Saving…",
        success: (result) => {
          onSave(String(result.serviceOrder.id));
          return "Service order created";
        },
        error: "Failed to create service order",
      },
    );
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Create service order"}
      description={"Description"}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={"Name"}>
            {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
          </FormField>
          <FormField name="type" label={"Type"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Type"}>
                  <SelectValue placeholder={"Select type"} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="repair">{"Repair"}</SelectItem>
                  <SelectItem value="maintenance">{"Maintenance"}</SelectItem>
                  <SelectItem value="installation">{"Installation"}</SelectItem>
                  <SelectItem value="inspection">{"Inspection"}</SelectItem>
                </SelectContent>
              </Select>
            )}
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
          <FormField name="contractId" label={"Contract"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Contract"}>
                  <SelectValue placeholder={"Select Contract"} />
                </SelectTrigger>
                <SelectContent>
                  {contracts.map((c) => (
                    <SelectItem key={c.id} value={String(c.id)}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="technicianId" label={"Technician"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Technician"}>
                  <SelectValue placeholder={"Select Technician"} />
                </SelectTrigger>
                <SelectContent>
                  {employees.map((e) => (
                    <SelectItem key={e.id} value={String(e.id)}>
                      {e.employeeNumber}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="priority" label={"Priority"}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="1" placeholder={"Priority"} />
            )}
          </FormField>
          <FormField name="reportedIssue" label={"Reported issue"}>
            {({ field, id }) => (
              <textarea
                {...field}
                id={id}
                className="flex min-h-[60px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                placeholder={"Reported issue"}
              />
            )}
          </FormField>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm">{"Lines"}</span>
            <Button
              type="button"
              size="sm"
              variant="outline"
              onClick={() =>
                append({
                  type: "labor",
                  itemId: "",
                  description: "",
                  qty: "1",
                  unitPrice: "0",
                  billable: true,
                })
              }
            >
              {"Add line"}
            </Button>
          </div>
          {fields.map((field, index) => (
            <div
              key={field.id}
              className="grid grid-cols-[100px_1fr_80px_100px_40px] items-end gap-2"
            >
              <div>
                <label
                  htmlFor={`${linePrefix}-type-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {"Line Type"}
                </label>
                <Select
                  value={form.watch(`lines.${index}.type`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.type`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-type-${index}`} aria-label={"Line Type"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="part">{"Part"}</SelectItem>
                    <SelectItem value="labor">{"Labor"}</SelectItem>
                    <SelectItem value="expense">{"Expense"}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-item-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {"Item"}
                </label>
                <Select
                  value={form.watch(`lines.${index}.itemId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={"Item"}>
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
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-qty-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {"Qty"}
                </label>
                <Input
                  id={`${linePrefix}-qty-${index}`}
                  type="number"
                  min="1"
                  {...form.register(`lines.${index}.qty`)}
                />
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-price-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {"Price"}
                </label>
                <Input
                  id={`${linePrefix}-price-${index}`}
                  type="number"
                  min="0"
                  step="0.01"
                  {...form.register(`lines.${index}.unitPrice`)}
                />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => remove(index)}
                disabled={fields.length <= 1}
              >
                x
              </Button>
            </div>
          ))}
        </div>
      </div>
    </EntityFormDialog>
  );
}
