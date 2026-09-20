"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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

const saleOrderFormSchema = z.object({
  contactId: z.string().min(1, "Select a customer."),
  priceBookId: z.string().min(1, "Select a price_book."),
  warehouseId: z.string().min(1, "Select a warehouse."),
  crmLeadId: z.string().min(1, "Select a won opportunity."),
  orderDate: z.string(),
  validityDate: z.string(),
  note: z.string(),
});

type SaleOrderFormValues = z.infer<typeof saleOrderFormSchema>;

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
    resolver: zodResolver(saleOrderFormSchema),
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
      toast.error("Add at least one line.");
      return;
    }
    if (lines.some((line) => !line.itemId || Number(line.qtyOrdered) <= 0)) {
      toast.error("Quantity must be greater than zero.");
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
    void getSwantaraService()
      .saleOrders.create(numericOrgId, request)
      .then(({ order }) => {
        toast.success("Quotation created.");
        onSave(String(order.id));
      })
      .catch(() => toast.error("Could not create the quotation."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"New quotation"}
      description={
        "Create a quotation from a won opportunity. PriceBook pricing is applied per line."
      }
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="contactId" label={"Customer"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Customer"}>
                <SelectValue placeholder={"Select customer"} />
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
        <FormField name="priceBookId" label={"PriceBook"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"PriceBook"}>
                <SelectValue placeholder={"Select price_book"} />
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
        <FormField name="warehouseId" label={"Warehouse"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Warehouse"}>
                <SelectValue placeholder={"Select warehouse"} />
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
        <FormField name="crmLeadId" label={"Opportunity (won)"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Opportunity (won)"}>
                <SelectValue placeholder={"Select opportunity"} />
              </SelectTrigger>
              <SelectContent>
                {opportunities.length === 0 ? (
                  <SelectItem value="" disabled>
                    {"No won opportunities"}
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
        <FormField name="orderDate" label={"Order date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="validityDate" label={"Validity date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="note" label={"Note"}>
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
