"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useFieldArray, useForm } from "react-hook-form";
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
import type { MOComponent } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type ConsumeDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  productionOrderId: string;
  components: MOComponent[];
  onSave: () => void;
};

function useConsumeFormSchema() {
  return z.object({
    lines: z.array(
      z.object({
        componentId: z.number(),
        qty: z.string(),
      }),
    ),
  });
}

export function ConsumeDialog({
  open,
  onOpenChange,
  orgId,
  productionOrderId,
  components,
  onSave,
}: ConsumeDialogProps) {
  const t = useTranslations("ProductionOrders");
  const tCommon = useTranslations("Common");
  const schema = useConsumeFormSchema();
  type Values = z.infer<typeof schema>;

  const remainingComponents = components.filter((c) => c.qtyConsumed < c.qtyPlanned);

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      lines: remainingComponents.map((c) => ({
        componentId: c.id,
        qty: String(c.qtyPlanned - c.qtyConsumed),
      })),
    },
  });

  const { fields } = useFieldArray({
    control: form.control,
    name: "lines",
    keyName: "fieldId",
  });

  const linesWatch = form.watch("lines");

  function handleSubmit(values: Values) {
    const consumeLines = values.lines
      .filter((l) => Number(l.qty) > 0)
      .map((l) => ({
        componentId: l.componentId,
        qty: Number(l.qty),
      }));

    if (consumeLines.length === 0) return;

    return Promise.all(
      consumeLines.map((line) =>
        getSwantaraService().productionOrders.consume(Number(orgId), Number(productionOrderId), {
          ...line,
          journalId: 1,
          wipAccountId: 1,
        }),
      ),
    )
      .then(() => {
        toast.success(t("materialsConsumed"));
        onSave();
      })
      .catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("consumeTitle")}</DialogTitle>
          <DialogDescription>{t("consumeDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-3">
            {fields.length === 0 ? (
              <div className="text-muted-foreground text-sm">{t("noComponentsToConsume")}</div>
            ) : (
              fields.map((field, index) => {
                const component = components.find((c) => c.id === field.componentId);
                return (
                  <div key={field.fieldId} className="grid items-center gap-2 sm:grid-cols-12">
                    <div className="sm:col-span-6">
                      <span className="text-sm">
                        {`#${component?.itemId ?? field.componentId}`}
                      </span>
                    </div>
                    <div className="sm:col-span-4">
                      <FormField name={`lines.${index}.qty`}>
                        {({ field: lineField, id }) => (
                          <Input {...lineField} id={id} type="number" min="0" />
                        )}
                      </FormField>
                    </div>
                    <div className="sm:col-span-2 text-right text-sm text-muted-foreground tabular-nums">
                      / {component?.qtyPlanned ?? 0}
                    </div>
                  </div>
                );
              })
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton disabled={linesWatch.every((l) => Number(l.qty) === 0)}>
              {t("consumeAction")}
            </SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
