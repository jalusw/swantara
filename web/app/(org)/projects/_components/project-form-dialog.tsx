"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Projects");
  return z
    .object({
      name: z.string().min(1, t("validation_nameRequired")),
      contactId: z.coerce.number().min(1, t("validation_customerRequired")),
      managerId: z.coerce.number().nullable(),
      billingType: z.enum(["fixed", "time_material", "milestone"]),
      billableRate: z.coerce.number().min(0, t("validation_rateRequired")),
      dimensionId: z.coerce.number().nullable(),
      saleOrderId: z.coerce.number().nullable(),
      dateStart: z.string().nullable(),
      dateEnd: z.string().nullable(),
    })
    .refine((data) => !data.dateStart || !data.dateEnd || data.dateEnd >= data.dateStart, {
      message: t("validation_endAfterStart"),
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
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const dyn = (key: string) => (t as unknown as (k: string) => string)(key);
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

    return op.then(() => onSave()).catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editProject") : t("newProject")}
      description={t("formDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="contactId" label={t("customer")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(Number(value))}
            >
              <SelectTrigger id={id} aria-label={t("customer")}>
                <SelectValue placeholder={t("selectCustomer")} />
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
        <FormField name="managerId" label={t("manager")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("manager")}>
                <SelectValue placeholder={t("selectManager")} />
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
        <FormField name="billingType" label={t("billingType")}>
          {({ field, id }) => (
            <Select
              value={field.value ?? ""}
              onValueChange={(value) =>
                field.onChange(value as "fixed" | "time_material" | "milestone")
              }
            >
              <SelectTrigger id={id} aria-label={t("billingType")}>
                <SelectValue placeholder={t("billingType")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="fixed">{dyn("billing_fixed")}</SelectItem>
                <SelectItem value="time_material">{dyn("billing_time_material")}</SelectItem>
                <SelectItem value="milestone">{dyn("billing_milestone")}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="billableRate" label={t("billableRate")}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" min="0" step="0.01" inputMode="decimal" />
          )}
        </FormField>
        <FormField name="dimensionId" label={t("dimensionAccount")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("dimensionAccount")}>
                <SelectValue placeholder={t("selectDimension")} />
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
        <FormField name="saleOrderId" label={t("saleOrder")}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={t("saleOrder")}>
                <SelectValue placeholder={t("selectSaleOrder")} />
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
        <FormField name="dateStart" label={t("startDate")}>
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
        <FormField name="dateEnd" label={t("endDate")}>
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
