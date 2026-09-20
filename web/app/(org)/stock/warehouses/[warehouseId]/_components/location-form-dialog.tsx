"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { getSwantaraService } from "@/lib/services/swantara";

const usageOptions = [
  "internal",
  "supplier",
  "customer",
  "transit",
  "production",
  "inventory",
  "scrap",
  "view",
];

function useLocationFormSchema() {
  return z.object({
    name: z.string().min(1, "Location name is required."),
    code: z.string(),
    usage: z.string().min(1, "Usage type is required."),
    barcode: z.string(),
  });
}

export function LocationFormDialog({
  open,
  onOpenChange,
  orgId,
  warehouseId,
  initial,
  parentId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  warehouseId: string;
  initial?: {
    id: string;
    name: string;
    code: string | null;
    usage: string;
  } | null;
  parentId?: string | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const schema = useLocationFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          code: initial.code || "",
          usage: initial.usage,
          barcode: "",
        }
      : {
          name: "",
          code: "",
          usage: "inventory",
          barcode: "",
        },
  });

  function resolveParentId(): number | null | undefined {
    if (initial) return undefined;
    if (parentId) return Number(parentId);
    return null;
  }

  function handleSubmit(values: Values) {
    const request = {
      organizationId: Number(orgId),
      warehouseId: Number(warehouseId),
      name: values.name,
      code: values.code || null,
      parentId: resolveParentId(),
      usage: values.usage,
      barcode: values.barcode || null,
    };

    if (isEdit && initial) {
      void getSwantaraService()
        .inventory.updateStockLocation(Number(orgId), Number(initial.id), {
          ...request,
          parentId: initial.id ? Number(initial.id) : null,
        })
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .inventory.createStockLocation(Number(orgId), {
          ...request,
          parentId: parentId ? Number(parentId) : null,
        })
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit location" : "New location"}</DialogTitle>
          <DialogDescription>{"Define a stock location within this warehouse."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="name" label={"Name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
            </FormField>
            <FormField name="code" label={"Code"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Code"} />}
            </FormField>
            <FormField name="usage" label={"Usage type"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Usage type"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {usageOptions.map((usage) => (
                      <SelectItem key={usage} value={usage}>
                        {usage}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="barcode" label={"Barcode"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Barcode"} />}
            </FormField>
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
