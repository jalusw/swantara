"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useLineResetEffect } from "@/lib/hooks/use-line-reset-effect";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  CrmLead,
  Item,
  ItemVariant,
  PriceBook,
  Warehouse,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { createEmptyLine, type LineRow, SaleOrderLinesEditor } from "./sale-order-lines-editor";

function useSaleOrderFormSchema() {
  const t = useTranslations("Sales");
  return z.object({
    contactId: z.string().min(1, t("validationCustomerRequired")),
    priceBookId: z.string().min(1, t("validationPriceBookRequired")),
    warehouseId: z.string().min(1, t("validationWarehouseRequired")),
    crmLeadId: z.string().min(1, t("validationOpportunityRequired")),
    orderDate: z.string(),
    validityDate: z.string(),
    note: z.string(),
  });
}

type SaleOrderFormValues = z.infer<ReturnType<typeof useSaleOrderFormSchema>>;

type SaleOrderFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initialOpportunityId?: string | null;
  onSave: (saleOrderId: string) => void;
};

export function SaleOrderFormDialog({
  open,
  onOpenChange,
  orgId,
  initialOpportunityId,
  onSave,
}: SaleOrderFormDialogProps) {
  const t = useTranslations("Sales");
  const tCommon = useTranslations("Common");
  const numericOrgId = Number(orgId);

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const priceBooksQuery = useOrgListQuery<{ priceBooks: PriceBook[] }, Record<string, never>>(
    "price_books",
    (organizationId) => getSwantaraService().priceBooks.list(organizationId),
  );
  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );
  const opportunitiesQuery = useOrgListQuery<{ opportunities: CrmLead[] }, Record<string, never>>(
    "crmOpportunities",
    (organizationId) => getSwantaraService().crmOpportunities.list(organizationId),
  );
  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const priceBooks = useMemo(() => priceBooksQuery.data?.priceBooks ?? [], [priceBooksQuery.data]);
  const warehouses = useMemo(() => warehousesQuery.data?.warehouses ?? [], [warehousesQuery.data]);
  const opportunities = useMemo(
    () => (opportunitiesQuery.data?.opportunities ?? []).filter((opportunity) => opportunity.isWon),
    [opportunitiesQuery.data],
  );
  const products = productsQuery.data?.products ?? [];

  const [lines, setLines] = useState<LineRow[]>([createEmptyLine()]);
  const [variantsByProduct, setVariantsByProduct] = useState<Record<string, ItemVariant[]>>({});

  useLineResetEffect(open, setLines, createEmptyLine);

  const form = useForm<SaleOrderFormValues>({
    resolver: zodResolver(useSaleOrderFormSchema()),
    defaultValues: {
      contactId: "",
      priceBookId: "",
      warehouseId: "",
      crmLeadId: initialOpportunityId ?? "",
      orderDate: getLocalDateString(),
      validityDate: "",
      note: "",
    },
  });

  useEffect(() => {
    if (!open) return;
    form.reset({
      contactId: "",
      priceBookId: priceBooks[0] ? String(priceBooks[0].id) : "",
      warehouseId: warehouses[0] ? String(warehouses[0].id) : "",
      crmLeadId: initialOpportunityId ?? "",
      orderDate: getLocalDateString(),
      validityDate: "",
      note: "",
    });
  }, [open, form, priceBooks, warehouses, initialOpportunityId]);

  function loadVariants(itemId: string) {
    if (variantsByProduct[itemId]) return;
    void getSwantaraService()
      .products.variants.list(numericOrgId, Number(itemId))
      .then((data) => setVariantsByProduct((prev) => ({ ...prev, [itemId]: data.variants })))
      .catch(() => setVariantsByProduct((prev) => ({ ...prev, [itemId]: [] })));
  }

  function handleSubmit(values: SaleOrderFormValues) {
    if (lines.length === 0) {
      toast.error(t("addAtLeastOneLine"));
      return;
    }
    if (lines.some((line) => !line.itemId || Number(line.qtyOrdered) <= 0)) {
      toast.error(t("quantityMustBePositive"));
      return;
    }
    const request = {
      organizationId: numericOrgId,
      contactId: Number(values.contactId),
      priceBookId: Number(values.priceBookId),
      warehouseId: Number(values.warehouseId),
      crmLeadId: Number(values.crmLeadId),
      salespersonId: null as number | null,
      salesGroupId: null as number | null,
      shipAddressId: null as number | null,
      billAddressId: null as number | null,
      orderDate: values.orderDate || null,
      expectedDate: null as string | null,
      validityDate: values.validityDate || null,
      paymentTermId: null as number | null,
      incoterm: null as string | null,
      customerPoRef: null as string | null,
      note: values.note || null,
      lines: lines.map((line, index) => ({
        sequence: (index + 1) * 10,
        itemId: Number(line.itemId),
        description: line.description || null,
        qtyOrdered: Number(line.qtyOrdered),
        unitId: null as number | null,
        discountPct: Number(line.discountPct) || 0,
        taxIds: [] as number[],
        dimensionId: null as number | null,
      })),
    };
    return getSwantaraService()
      .saleOrders.create(numericOrgId, request)
      .then(({ order }) => {
        toast.success(t("quotationCreated"));
        onSave(String(order.id));
      })
      .catch(() => void toast.error(t("quotationCreateFailed")));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("newQuotation")}
      description={t("newQuotationDescription")}
      form={form}
      onSubmit={handleSubmit}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="contactId" label={t("tableCustomer")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableCustomer")}>
                <SelectValue placeholder={t("selectCustomer")} />
              </SelectTrigger>
              <SelectContent>
                {contacts.map((contact) => (
                  <SelectItem key={contact.id} value={String(contact.id)}>
                    {contact.displayName || contact.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="priceBookId" label={t("fieldPriceBook")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldPriceBook")}>
                <SelectValue placeholder={t("selectPriceBook")} />
              </SelectTrigger>
              <SelectContent>
                {priceBooks.map((priceBook) => (
                  <SelectItem key={priceBook.id} value={String(priceBook.id)}>
                    {priceBook.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="warehouseId" label={t("fieldWarehouse")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldWarehouse")}>
                <SelectValue placeholder={t("selectWarehouse")} />
              </SelectTrigger>
              <SelectContent>
                {warehouses.map((warehouse) => (
                  <SelectItem key={warehouse.id} value={String(warehouse.id)}>
                    {warehouse.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="crmLeadId" label={t("fieldOpportunityWon")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldOpportunityWon")}>
                <SelectValue placeholder={t("selectOpportunity")} />
              </SelectTrigger>
              <SelectContent>
                {opportunities.length === 0 ? (
                  <SelectItem value="" disabled>
                    {t("noWonOpportunities")}
                  </SelectItem>
                ) : (
                  opportunities.map((opportunity) => (
                    <SelectItem key={opportunity.id} value={String(opportunity.id)}>
                      {opportunity.name}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="orderDate" label={t("tableOrderDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="validityDate" label={t("fieldValidityDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="note" label={t("fieldNote")}>
          {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
        </FormField>
      </div>
      <SaleOrderLinesEditor
        lines={lines}
        products={products}
        variantsByProduct={variantsByProduct}
        priceBookId={form.watch("priceBookId")}
        orderDate={form.watch("orderDate")}
        onAdd={() => setLines((prev) => [...prev, createEmptyLine()])}
        onUpdate={(id, patch) =>
          setLines((prev) => prev.map((line) => (line.id === id ? { ...line, ...patch } : line)))
        }
        onRemove={(id) => setLines((prev) => prev.filter((line) => line.id !== id))}
        onSelectItem={loadVariants}
      />
    </EntityFormDialog>
  );
}
