"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("Service");
  const lineSchema = z.object({
    type: z.string().min(1),
    itemId: z.string().optional(),
    description: z.string().optional(),
    qty: z.string().min(1),
    unitPrice: z.string().optional(),
    billable: z.boolean(),
  });
  const schema = z.object({
    name: z.string().min(1, t("validation_nameRequired")),
    contactId: z.string().optional(),
    equipmentId: z.string().optional(),
    contractId: z.string().optional(),
    type: z.string().min(1, t("validation_typeRequired")),
    priority: z.string().optional(),
    technicianId: z.string().optional(),
    reportedIssue: z.string().optional(),
    lines: z.array(lineSchema).min(1, t("validation_linesRequired")),
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
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
  const orderType = (type: string) => (t as unknown as (k: string) => string)(`type_${type}`);
  const lineType = (type: string) => (t as unknown as (k: string) => string)(`lineType_${type}`);
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

  async function handleSubmit(values: ServiceOrderValues) {
    const request = getSwantaraService().serviceOrders.create(Number(orgId), {
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
    });
    toast.promise(request, {
      loading: t("saving"),
      success: (result) => {
        onSave(String(result.serviceOrder.id));
        return t("orderCreated");
      },
      error: t("createFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newOrder")}
      description={t("formDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="flex flex-col gap-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={t("fieldName")}>
            {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
          </FormField>
          <FormField name="type" label={t("orderType")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("orderType")}>
                  <SelectValue placeholder={t("selectType")} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="repair">{orderType("repair")}</SelectItem>
                  <SelectItem value="maintenance">{orderType("maintenance")}</SelectItem>
                  <SelectItem value="installation">{orderType("installation")}</SelectItem>
                  <SelectItem value="inspection">{orderType("inspection")}</SelectItem>
                </SelectContent>
              </Select>
            )}
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
          <FormField name="contractId" label={t("contract")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("contract")}>
                  <SelectValue placeholder={t("selectContract")} />
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
          <FormField name="technicianId" label={t("technician")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("technician")}>
                  <SelectValue placeholder={t("selectTechnician")} />
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
          <FormField name="priority" label={t("colPriority")}>
            {({ field, id }) => (
              <Input {...field} id={id} type="number" min="1" placeholder={t("colPriority")} />
            )}
          </FormField>
          <FormField name="reportedIssue" label={t("reportedIssue")}>
            {({ field, id }) => (
              <textarea
                {...field}
                id={id}
                className="flex min-h-[60px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                placeholder={t("reportedIssue")}
              />
            )}
          </FormField>
        </div>

        <div className="flex flex-col gap-2">
          <div className="flex items-center justify-between">
            <span className="text-sm">{t("lines")}</span>
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
              {t("addLine")}
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
                  {t("lineType")}
                </label>
                <Select
                  value={form.watch(`lines.${index}.type`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.type`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-type-${index}`} aria-label={t("lineType")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="part">{lineType("part")}</SelectItem>
                    <SelectItem value="labor">{lineType("labor")}</SelectItem>
                    <SelectItem value="expense">{lineType("expense")}</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-item-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {t("item")}
                </label>
                <Select
                  value={form.watch(`lines.${index}.itemId`)}
                  onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                >
                  <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={t("item")}>
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
              </div>
              <div>
                <label
                  htmlFor={`${linePrefix}-qty-${index}`}
                  className="text-xs text-muted-foreground"
                >
                  {t("qty")}
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
                  {t("price")}
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
