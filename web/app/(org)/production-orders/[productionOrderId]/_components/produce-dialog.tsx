"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("ProductionOrders");
  const tCommon = useTranslations("Common");
  const remaining = productionOrder.qtyToProduce - productionOrder.qtyProduced;

  const schema = z.object({
    qty: z
      .string()
      .min(1, t("validation_qtyRequired"))
      .refine((v) => Number(v) > 0, t("validation_qtyRequired"))
      .refine((v) => Number(v) <= remaining, t("validation_qtyExceeds")),
  });
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      qty: String(Math.max(0, remaining)),
    },
  });

  function handleSubmit(values: Values) {
    return getSwantaraService()
      .productionOrders.produce(Number(orgId), Number(productionOrderId), {
        qty: Number(values.qty),
        journalId: 1,
        wipAccountId: 1,
      })
      .then(() => {
        toast.success(t("goodsProduced"));
        onSave();
      })
      .catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("produceTitle")}</DialogTitle>
          <DialogDescription>{t("produceDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="qty" label={t("colQtyToProduce")}>
            {({ field, id }) => <Input {...field} id={id} type="number" min="0" max={remaining} />}
          </FormField>
          <span className="text-xs text-muted-foreground">
            {t("remainingCount", { remaining })}
          </span>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{t("produceAction")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
