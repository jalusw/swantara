"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useLineResetEffect } from "@/lib/hooks/use-line-reset-effect";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact, Item, Warehouse } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { createId, getLocalDateString } from "@/lib/utils";

type LineRow = {
  id: string;
  itemId: string;
  qtyOrdered: string;
  unitPrice: string;
  discountPct: string;
  description: string;
};

function usePurchaseOrderFormSchema() {
  return z.object({
    supplierId: z.string().min(1, "Select a supplier."),
    warehouseId: z.string(),
    vendorRef: z.string(),
    orderDate: z.string(),
    expectedDate: z.string(),
    incoterm: z.string(),
    note: z.string(),
  });
}
type PurchaseOrderFormValues = z.infer<ReturnType<typeof usePurchaseOrderFormSchema>>;

export function PurchaseOrderFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: () => void;
}) {
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );
  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const contacts = useMemo(() => contactsQuery.data?.contacts ?? [], [contactsQuery.data]);
  const warehouses = useMemo(() => warehousesQuery.data?.warehouses ?? [], [warehousesQuery.data]);
  const products = useMemo(() => productsQuery.data?.products ?? [], [productsQuery.data]);

  const [lines, setLines] = useState<LineRow[]>([
    {
      id: createId(),
      itemId: "",
      qtyOrdered: "1",
      unitPrice: "0",
      discountPct: "0",
      description: "",
    },
  ]);

  const createEmptyLine = useCallback(
    () => ({
      id: createId(),
      itemId: "",
      qtyOrdered: "1",
      unitPrice: "0",
      discountPct: "0",
      description: "",
    }),
    [],
  );

  useLineResetEffect(open, setLines, createEmptyLine);

  const schema = usePurchaseOrderFormSchema();

  const form = useForm<PurchaseOrderFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      supplierId: "",
      warehouseId: warehouses[0] ? String(warehouses[0].id) : "",
      vendorRef: "",
      orderDate: getLocalDateString(),
      expectedDate: "",
      incoterm: "",
      note: "",
    },
  });

  useEffect(() => {
    if (!open) return;
    form.reset({
      supplierId: "",
      warehouseId: warehouses[0] ? String(warehouses[0].id) : "",
      vendorRef: "",
      orderDate: getLocalDateString(),
      expectedDate: "",
      incoterm: "",
      note: "",
    });
  }, [open, form, warehouses]);

  function addLine() {
    setLines((prev) => [
      ...prev,
      {
        id: createId(),
        itemId: "",
        qtyOrdered: "1",
        unitPrice: "0",
        discountPct: "0",
        description: "",
      },
    ]);
  }

  function updateLine(id: string, patch: Partial<LineRow>) {
    setLines((prev) => prev.map((line) => (line.id === id ? { ...line, ...patch } : line)));
  }

  function removeLine(id: string) {
    setLines((prev) => prev.filter((line) => line.id !== id));
  }

  function handleSubmit(values: PurchaseOrderFormValues) {
    if (lines.length === 0) {
      toast.error("Add at least one line.");
      return;
    }
    const invalid = lines.some((line) => !line.itemId || Number(line.qtyOrdered) <= 0);
    if (invalid) {
      toast.error("Quantity must be greater than zero.");
      return;
    }
    const request = {
      organizationId: Number(orgId),
      requestId: null as number | null,
      supplierId: Number(values.supplierId),
      vendorRef: values.vendorRef || null,
      currencyCode: null as string | null,
      warehouseId: values.warehouseId ? Number(values.warehouseId) : null,
      destLocationId: null as number | null,
      orderDate: values.orderDate || null,
      expectedDate: values.expectedDate || null,
      paymentTermId: null as number | null,
      incoterm: values.incoterm || null,
      lines: lines.map((line, index) => ({
        sequence: (index + 1) * 10,
        itemId: Number(line.itemId),
        description: line.description || null,
        qtyOrdered: Number(line.qtyOrdered),
        unitId: null as number | null,
        unitPrice: Number(line.unitPrice),
        discountPct: Number(line.discountPct) || 0,
        taxIds: [] as number[],
        dimensionId: null as number | null,
      })),
    };
    void getSwantaraService()
      .purchaseOrders.create(Number(orgId), request)
      .then(() => {
        toast.success("Purchase order created.");
        onSave();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"New purchase order"}
      description={"Create a purchase order for a supplier."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="supplierId" label={"Supplier"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Supplier"}>
                <SelectValue placeholder={"Select supplier"} />
              </SelectTrigger>
              <SelectContent>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
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
                {warehouses.map((w) => (
                  <SelectItem key={w.id} value={String(w.id)}>
                    {w.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="vendorRef" label={"Supplier reference"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="orderDate" label={"Order date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="expectedDate" label={"Expected date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="incoterm" label={"Incoterm"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="note" label={"Note"}>
          {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
        </FormField>
      </div>

      <div className="mt-6 flex flex-col gap-3 rounded-md border p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm">{"Order lines"}</h3>
          <Button type="button" variant="outline" size="sm" onClick={addLine}>
            {"Add line"}
          </Button>
        </div>
        {lines.length === 0 ? (
          <p className="text-sm text-muted-foreground">{"Add at least one line."}</p>
        ) : null}
        <div className="flex flex-col gap-3">
          {lines.map((line, index) => (
            <div key={line.id} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
              <div className="sm:col-span-4">
                <span className="text-xs text-muted-foreground">{"Item"}</span>
                <Select
                  value={line.itemId}
                  onValueChange={(value) => updateLine(line.id, { itemId: value ?? "" })}
                >
                  <SelectTrigger aria-label={`${"Item"} ${index + 1}`}>
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
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Qty"}</span>
                <Input
                  value={line.qtyOrdered}
                  onChange={(e) => updateLine(line.id, { qtyOrdered: e.target.value })}
                  type="number"
                  min="0"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Unit price"}</span>
                <Input
                  value={line.unitPrice}
                  onChange={(e) => updateLine(line.id, { unitPrice: e.target.value })}
                  type="number"
                  min="0"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Discount %"}</span>
                <Input
                  value={line.discountPct}
                  onChange={(e) => updateLine(line.id, { discountPct: e.target.value })}
                  type="number"
                  min="0"
                  max="100"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2 flex items-end">
                <Button type="button" variant="ghost" size="sm" onClick={() => removeLine(line.id)}>
                  {"Remove"}
                </Button>
              </div>
              <div className="sm:col-span-12">
                <Input
                  value={line.description}
                  onChange={(e) => updateLine(line.id, { description: e.target.value })}
                  placeholder={"Description"}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </EntityFormDialog>
  );
}
