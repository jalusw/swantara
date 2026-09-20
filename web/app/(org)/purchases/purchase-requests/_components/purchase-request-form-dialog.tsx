"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useCallback, useState } from "react";
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
import type { Contact, Department, Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { createId } from "@/lib/utils";

type LineRow = {
  id: string;
  itemId: string;
  description: string;
  qty: string;
  neededBy: string;
};

function usePurchaseRequestFormSchema() {
  return z.object({
    requesterId: z.string().min(1, "Select a requester."),
    departmentId: z.string(),
    neededBy: z.string(),
    note: z.string(),
  });
}
type PurchaseRequestFormValues = z.infer<ReturnType<typeof usePurchaseRequestFormSchema>>;

export function PurchaseRequestFormDialog({
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
  const departmentsQuery = useOrgListQuery<{ departments: Department[] }, Record<string, never>>(
    "departments",
    (organizationId) => getSwantaraService().departments.list(organizationId),
  );
  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const departments = departmentsQuery.data?.departments ?? [];
  const products = productsQuery.data?.products ?? [];

  const [lines, setLines] = useState<LineRow[]>([
    { id: createId(), itemId: "", description: "", qty: "1", neededBy: "" },
  ]);

  const createEmptyLine = useCallback(
    () => ({
      id: createId(),
      itemId: "",
      description: "",
      qty: "1",
      neededBy: "",
    }),
    [],
  );

  useLineResetEffect(open, setLines, createEmptyLine);

  const schema = usePurchaseRequestFormSchema();

  const form = useForm<PurchaseRequestFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      requesterId: "",
      departmentId: "",
      neededBy: "",
      note: "",
    },
  });

  function addLine() {
    setLines((prev) => [
      ...prev,
      {
        id: createId(),
        itemId: "",
        description: "",
        qty: "1",
        neededBy: "",
      },
    ]);
  }

  function updateLine(id: string, patch: Partial<LineRow>) {
    setLines((prev) => prev.map((line) => (line.id === id ? { ...line, ...patch } : line)));
  }

  function removeLine(id: string) {
    setLines((prev) => prev.filter((line) => line.id !== id));
  }

  function handleSubmit(values: PurchaseRequestFormValues) {
    if (lines.length === 0) {
      toast.error("Add at least one line.");
      return;
    }
    const invalid = lines.some((line) => !line.itemId || Number(line.qty) <= 0);
    if (invalid) {
      toast.error("Quantity must be greater than zero.");
      return;
    }
    const request = {
      organizationId: Number(orgId),
      requesterId: Number(values.requesterId),
      departmentId: values.departmentId ? Number(values.departmentId) : null,
      neededBy: values.neededBy || null,
      lines: lines.map((line) => ({
        itemId: Number(line.itemId),
        description: line.description || null,
        qty: Number(line.qty),
        unitId: null as number | null,
        neededBy: line.neededBy || null,
      })),
    };
    void getSwantaraService()
      .purchaseRequests.create(Number(orgId), request)
      .then(() => {
        toast.success("Request created.");
        onSave();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"New purchase request"}
      description={"Create a purchase request with item lines and quantities."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-3xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="requesterId" label={"Requester"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Requester"}>
                <SelectValue placeholder={"Select requester"} />
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
        <FormField name="departmentId" label={"Department"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Department"}>
                <SelectValue placeholder={"Select department"} />
              </SelectTrigger>
              <SelectContent>
                {departments.map((d) => (
                  <SelectItem key={d.id} value={String(d.id)}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="neededBy" label={"Needed by"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="note" label={"Note"}>
          {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
        </FormField>
      </div>

      <div className="mt-6 flex flex-col gap-3 rounded-md border p-4">
        <div className="flex items-center justify-between">
          <h3 className="text-sm">{"Request lines"}</h3>
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
              <div className="sm:col-span-5">
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
                  value={line.qty}
                  onChange={(e) => updateLine(line.id, { qty: e.target.value })}
                  type="number"
                  min="0"
                  step="any"
                />
              </div>
              <div className="sm:col-span-2">
                <span className="text-xs text-muted-foreground">{"Needed by"}</span>
                <Input
                  value={line.neededBy}
                  onChange={(e) => updateLine(line.id, { neededBy: e.target.value })}
                  type="date"
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
