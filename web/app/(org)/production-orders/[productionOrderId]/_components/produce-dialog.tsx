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
import type { ProductionOrder } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type ProduceDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  productionOrderId: string;
  productionOrder: ProductionOrder;
  onSave: () => void;
};

export function ProduceDialog({
  open,
  onOpenChange,
  orgId,
  productionOrderId,
  productionOrder,
  onSave,
}: ProduceDialogProps) {
  const remaining = productionOrder.qtyToProduce - productionOrder.qtyProduced;

  const schema = z.object({
    qty: z
      .string()
      .min(1, "Enter a quantity.")
      .refine((v) => Number(v) > 0, "Enter a quantity.")
      .refine((v) => Number(v) <= remaining, "Quantity Exceeds Remaining"),
  });
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      qty: String(Math.max(0, remaining)),
    },
  });

  function handleSubmit(values: Values) {
    void getSwantaraService()
      .productionOrders.produce(Number(orgId), Number(productionOrderId), {
        qty: Number(values.qty),
        journalId: 1,
        wipAccountId: 1,
      })
      .then(() => {
        toast.success("Goods produced");
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{"Produce Goods"}</DialogTitle>
          <DialogDescription>{"Record production output"}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="qty" label={"Quantity to Produce"}>
            {({ field, id }) => <Input {...field} id={id} type="number" min="0" max={remaining} />}
          </FormField>
          <span className="text-xs text-muted-foreground">{`${remaining} remaining`}</span>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Produce"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
