"use client";

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
import { Switch } from "@/components/switch";
import type { ReorderRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useReorderRuleFormSchema() {
  return z.object({
    itemId: z.coerce.number().min(1, "Item ID is required."),
    minQty: z.coerce.number().min(0),
    maxQty: z.coerce.number().min(0),
    qtyMultiple: z.coerce.number().min(1),
    leadTimeDays: z.coerce.number().nullable(),
    active: z.boolean(),
  });
}

export function ReorderRuleFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: ReorderRule | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const schema = useReorderRuleFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          itemId: initial.itemId,
          minQty: initial.minQty,
          maxQty: initial.maxQty,
          qtyMultiple: initial.qtyMultiple,
          leadTimeDays: initial.leadTimeDays,
          active: initial.active,
        }
      : {
          itemId: 0,
          minQty: 0,
          maxQty: 0,
          qtyMultiple: 1,
          leadTimeDays: null,
          active: true,
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      itemId: values.itemId,
      warehouseId: initial?.warehouseId ?? null,
      locationId: initial?.locationId ?? null,
      minQty: values.minQty,
      maxQty: values.maxQty,
      qtyMultiple: values.qtyMultiple,
      leadTimeDays: values.leadTimeDays,
      active: values.active,
    };

    if (isEdit && initial) {
      void getSwantaraService()
        .inventory.updateReorderRule(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .inventory.createReorderRule(Number(orgId), request)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit reorder rule" : "New reorder rule"}</DialogTitle>
          <DialogDescription>
            {"Define minimum and maximum stock levels for automatic replenishment."}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="itemId" label={"Item"}>
              {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
            </FormField>
            <div className="grid gap-4 sm:grid-cols-3">
              <FormField name="minQty" label={"Min qty"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
              </FormField>
              <FormField name="maxQty" label={"Max qty"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
              </FormField>
              <FormField name="qtyMultiple" label={"Qty multiple"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
              </FormField>
            </div>
            <FormField name="leadTimeDays" label={"Lead time (days)"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  min="0"
                  value={field.value ?? ""}
                  onChange={(event) => {
                    const value = event.target.value;
                    field.onChange(value === "" ? null : Number(value));
                  }}
                />
              )}
            </FormField>
            <FormField name="active" label={"Active"}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={"Active"}
                />
              )}
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
