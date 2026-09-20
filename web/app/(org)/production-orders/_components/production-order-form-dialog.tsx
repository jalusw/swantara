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
import { Textarea } from "@/components/textarea";
import type { Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type MoFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  products: Item[];
  onSave: () => void;
};

function useMoFormSchema() {
  return z.object({
    itemId: z.string().min(1, "Select a item."),
    qty: z
      .string()
      .min(1, "Enter a quantity.")
      .refine((v) => Number(v) >= 1, "Enter a quantity."),
    note: z.string(),
  });
}

export function MoFormDialog({ open, onOpenChange, orgId, products, onSave }: MoFormDialogProps) {
  const schema = useMoFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      itemId: "",
      qty: "1",
      note: "",
    },
  });

  function handleSubmit(values: Values) {
    void getSwantaraService()
      .productionOrders.create(Number(orgId), {
        itemId: Number(values.itemId),
        qtyToProduce: Number(values.qty) || 1,
        bomId: 0,
        srcLocationId: 0,
        dstLocationId: 0,
      })
      .then(() => {
        toast.success("Manufacturing order created");
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Create Manufacturing Order"}</DialogTitle>
          <DialogDescription>{"Create a new manufacturing order"}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="itemId" label={"Item"}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={"Item"}>
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
            )}
          </FormField>
          <FormField name="qty" label={"Quantity to Produce"}>
            {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
          </FormField>
          <FormField name="note" label={"Note"}>
            {({ field, id }) => <Textarea {...field} id={id} rows={3} />}
          </FormField>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Create Order"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
