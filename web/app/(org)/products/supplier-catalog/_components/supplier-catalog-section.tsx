"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
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
import type { Item, SupplierProduct } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber, getLocalDateString } from "@/lib/utils";
import { logger } from "@/lib/utils/logger";

const schema = z.object({
  itemId: z.string().min(1),
  supplierId: z.string().min(1),
  vendorSku: z.string(),
  price: z.string(),
  currencyCode: z.string(),
  leadTimeDays: z.string(),
  priority: z.string(),
  minQty: z.string(),
  validFrom: z.string(),
  validTo: z.string(),
});
type Values = z.infer<typeof schema>;

export function SupplierCatalogSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<SupplierProduct | null>(null);

  const query = useOrgListQuery<{ supplierProducts: SupplierProduct[] }, Record<string, never>>(
    "supplierProducts",
    (organizationId) => getSwantaraService().supplierProducts.list(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const items = query.data?.supplierProducts ?? [];
  const products = productsQuery.data?.products ?? [];

  function productName(itemId: number): string {
    return products.find((p) => p.id === itemId)?.name ?? `#${itemId}`;
  }

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: editing
      ? {
          itemId: String(editing.itemId),
          supplierId: String(editing.supplierId),
          vendorSku: editing.vendorSku ?? "",
          price: editing.price != null ? String(editing.price) : "",
          currencyCode: editing.currencyCode ?? "USD",
          leadTimeDays: editing.leadTimeDays != null ? String(editing.leadTimeDays) : "",
          priority: String(editing.priority),
          minQty: String(editing.minQty),
          validFrom: editing.validFrom ? getLocalDateString(new Date(editing.validFrom)) : "",
          validTo: editing.validTo ? getLocalDateString(new Date(editing.validTo)) : "",
        }
      : {
          itemId: "",
          supplierId: "",
          vendorSku: "",
          price: "",
          currencyCode: "USD",
          leadTimeDays: "",
          priority: "1",
          minQty: "1",
          validFrom: "",
          validTo: "",
        },
  });

  function openCreate() {
    setEditing(null);
    form.reset({
      itemId: "",
      supplierId: "",
      vendorSku: "",
      price: "",
      currencyCode: "USD",
      leadTimeDays: "",
      priority: "1",
      minQty: "1",
      validFrom: "",
      validTo: "",
    });
    setDialogOpen(true);
  }

  function openEdit(item: SupplierProduct) {
    setEditing(item);
    form.reset({
      itemId: String(item.itemId),
      supplierId: String(item.supplierId),
      vendorSku: item.vendorSku ?? "",
      price: item.price != null ? String(item.price) : "",
      currencyCode: item.currencyCode ?? "USD",
      leadTimeDays: item.leadTimeDays != null ? String(item.leadTimeDays) : "",
      priority: String(item.priority),
      minQty: String(item.minQty),
      validFrom: item.validFrom ? getLocalDateString(new Date(item.validFrom)) : "",
      validTo: item.validTo ? getLocalDateString(new Date(item.validTo)) : "",
    });
    setDialogOpen(true);
  }

  function handleSubmit(values: Values) {
    const base = {
      supplierId: Number(values.supplierId),
      vendorSku: values.vendorSku || null,
      vendorProductName: null,
      minQty: Number(values.minQty) || 1,
      price: values.price ? Number(values.price) : null,
      currencyCode: values.currencyCode || null,
      leadTimeDays: values.leadTimeDays ? Number(values.leadTimeDays) : null,
      priority: Number(values.priority) || 1,
      validFrom: values.validFrom ? new Date(values.validFrom) : null,
      validTo: values.validTo ? new Date(values.validTo) : null,
    };

    if (editing) {
      void getSwantaraService()
        .supplierProducts.update(Number(orgId), editing.id, base)
        .then(() => {
          toast.success("Supplier item added.");
          setDialogOpen(false);
          setEditing(null);
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update supplier item", error);
          toast.error("Could not disable the organization.");
        });
    } else {
      void getSwantaraService()
        .supplierProducts.create(Number(orgId), {
          ...base,
          itemId: Number(values.itemId),
        })
        .then(() => {
          toast.success("Supplier item added.");
          setDialogOpen(false);
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create supplier item", error);
          toast.error("Could not disable the organization.");
        });
    }
  }

  function handleDelete(item: SupplierProduct) {
    void getSwantaraService()
      .supplierProducts.delete(Number(orgId), item.id)
      .then(() => {
        toast.success("Supplier item removed.");
        void query.refetch();
      })
      .catch((error) => {
        logger.error("Failed to delete supplier item", error);
        toast.error("Could not disable the organization.");
      });
  }

  const columns: ColumnDef<SupplierProduct>[] = [
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => <span className="">{productName(row.original.itemId)}</span>,
    },
    {
      accessorKey: "supplierId",
      header: "Supplier",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {"Supplier #"} {row.original.supplierId}
        </span>
      ),
    },
    {
      accessorKey: "vendorSku",
      header: "Supplier SKU",
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.vendorSku ?? "—"}</span>,
    },
    {
      accessorKey: "price",
      header: "Price",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.price != null ? formatNumber(row.original.price) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "leadTimeDays",
      header: "Lead time (days)",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.leadTimeDays != null ? `${row.original.leadTimeDays}d` : "—"}
        </span>
      ),
    },
    {
      accessorKey: "priority",
      header: "Priority",
      cell: ({ row }) => <Badge variant="secondary">{row.original.priority}</Badge>,
    },
    {
      accessorKey: "validTo",
      header: "Valid to",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.validTo ? formatDate(new Date(row.original.validTo)) : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit supplier item"}
          deleteLabel={"Delete"}
          confirmTitle={"Remove this supplier item?"}
          confirmDescription={"The link between this supplier and item will be removed."}
          confirmLabel="OK"
          onEdit={() => openEdit(row.original)}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle>{"Supplier catalog"}</CardTitle>
            <CardDescription>
              {"Item-to-supplier pricing, lead times, and priority."}
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={columns}
            data={items}
            getRowId={(row) => String(row.id)}
            searchKeys={["vendorSku"]}
            searchPlaceholder={"Search supplier products…"}
            filterLabel={"Priority"}
            allLabel={"All supplier products"}
            ariaLabel={"All supplier products"}
            statusOptions={[]}
            status={
              query.isLoading
                ? { type: "loading" }
                : query.isError
                  ? {
                      type: "error",
                      message: query.error.message,
                      onRetry: () => void query.refetch(),
                    }
                  : undefined
            }
            actions={
              <Button size="sm" onClick={openCreate}>
                <Plus />
                <span>{"Add supplier item"}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{editing ? "Edit supplier item" : "New supplier item"}</DialogTitle>
            <DialogDescription>
              {"Link a supplier to a item with pricing and lead time."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              {!editing ? (
                <FormField name="itemId" label={"Item"}>
                  {({ field, id }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id={id} aria-label={"Item"}>
                        <SelectValue placeholder={"Item"} />
                      </SelectTrigger>
                      <SelectContent>
                        {products.map((p) => (
                          <SelectItem key={p.id} value={String(p.id)}>
                            {p.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormField>
              ) : (
                <FormField name="itemId" label={"Item"}>
                  {({ field, id }) => (
                    <Input {...field} id={id} disabled value={productName(Number(field.value))} />
                  )}
                </FormField>
              )}
              <FormField name="supplierId" label={"Supplier"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" placeholder={"Supplier"} />
                )}
              </FormField>
              <FormField name="vendorSku" label={"Supplier SKU"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Supplier SKU"} />}
              </FormField>
              <FormField name="price" label={"Price"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" step="any" placeholder={"Price"} />
                )}
              </FormField>
              <FormField name="currencyCode" label={"Currency"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Currency"} />}
              </FormField>
              <FormField name="leadTimeDays" label={"Lead time (days)"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" placeholder={"Lead time (days)"} />
                )}
              </FormField>
              <FormField name="priority" label={"Priority"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" placeholder={"Priority"} />
                )}
              </FormField>
              <FormField name="minQty" label={"Min. quantity"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" placeholder={"Min. quantity"} />
                )}
              </FormField>
              <FormField name="validFrom" label={"Valid from"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="validTo" label={"Valid to"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
