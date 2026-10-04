"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useId } from "react";
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

function useInboundCostFormSchema() {
  const t = useTranslations("Purchases");
  const lineSchema = z.object({
    itemId: z.string().min(1, t("validationItemRequired")),
    description: z.string(),
    amount: z.string().min(1, t("validationAmountRequired")),
    splitMethod: z.enum(["by_quantity", "by_weight", "by_volume", "by_value", "equal"]),
  });
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    date: z.string(),
    targetShipmentIds: z.array(z.string()),
    lines: z.array(lineSchema).min(1, t("addAtLeastOneLine")),
  });
}

export function InboundCostFormDialog({
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
  const t = useTranslations("Purchases");
  const tCommon = useTranslations("Common");
  const linePrefix = useId();

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );
  const products = productsQuery.data?.products ?? [];

  const schema = useInboundCostFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      date: getLocalDateString(),
      targetShipmentIds: [],
      lines: [
        {
          itemId: "",
          description: "",
          amount: "0",
          splitMethod: "by_quantity",
        },
      ],
    },
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "lines",
  });

  async function handleSubmit(values: Values) {
    const request = getSwantaraService().inventory.createInboundCost(Number(orgId), {
      name: values.name,
      date: values.date || null,
      targetShipmentIds: values.targetShipmentIds.map(Number),
      lines: values.lines.map((l) => ({
        itemId: Number(l.itemId),
        description: l.description,
        amount: Number(l.amount),
        vendorBillLineId: null,
        splitMethod: l.splitMethod,
        accountId: null,
      })),
    });
    toast.promise(request, {
      loading: t("saving"),
      success: () => {
        onSave();
        return t("inboundCostCreated");
      },
      error: t("inboundCostCreateFailed"),
    });
    await request.catch(() => {});
  }

  function splitMethodLabel(method: string): string {
    try {
      return (t as unknown as (k: string) => string)(`splitMethod.${method}`);
    } catch {
      return method;
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("newInboundCost")}</DialogTitle>
          <DialogDescription>{t("newInboundCostDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("tableName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("tableName")} />}
              </FormField>
              <FormField name="date" label={t("tableDate")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-sm">{t("costLines")}</span>
              {fields.map((field, index) => (
                <div
                  key={field.id}
                  className="grid grid-cols-[1fr_1fr_120px_120px_40px] items-end gap-2"
                >
                  <div>
                    <label
                      htmlFor={`${linePrefix}-item-${index}`}
                      className="text-muted-foreground text-xs"
                    >
                      {t("fieldItem")}
                    </label>
                    <Select
                      value={form.watch(`lines.${index}.itemId`)}
                      onValueChange={(v) => form.setValue(`lines.${index}.itemId`, v ?? "")}
                    >
                      <SelectTrigger id={`${linePrefix}-item-${index}`} aria-label={t("fieldItem")}>
                        <SelectValue placeholder={t("selectItem")} />
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
                  <div>
                    <label
                      htmlFor={`${linePrefix}-desc-${index}`}
                      className="text-muted-foreground text-xs"
                    >
                      {t("fieldDescription")}
                    </label>
                    <Input
                      id={`${linePrefix}-desc-${index}`}
                      {...form.register(`lines.${index}.description`)}
                    />
                  </div>
                  <div>
                    <label
                      htmlFor={`${linePrefix}-amount-${index}`}
                      className="text-muted-foreground text-xs"
                    >
                      {t("tableAmount")}
                    </label>
                    <Input
                      id={`${linePrefix}-amount-${index}`}
                      type="number"
                      min="0"
                      step="0.01"
                      {...form.register(`lines.${index}.amount`)}
                    />
                  </div>
                  <div>
                    <label
                      htmlFor={`${linePrefix}-split-${index}`}
                      className="text-muted-foreground text-xs"
                    >
                      {t("fieldSplitMethod")}
                    </label>
                    <Select
                      value={form.watch(`lines.${index}.splitMethod`)}
                      onValueChange={(v) =>
                        form.setValue(
                          `lines.${index}.splitMethod`,
                          v as "by_quantity" | "by_weight" | "by_volume" | "by_value" | "equal",
                        )
                      }
                    >
                      <SelectTrigger
                        id={`${linePrefix}-split-${index}`}
                        aria-label={t("fieldSplitMethod")}
                      >
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="by_quantity">
                          {splitMethodLabel("by_quantity")}
                        </SelectItem>
                        <SelectItem value="by_weight">{splitMethodLabel("by_weight")}</SelectItem>
                        <SelectItem value="by_volume">{splitMethodLabel("by_volume")}</SelectItem>
                        <SelectItem value="by_value">{splitMethodLabel("by_value")}</SelectItem>
                        <SelectItem value="equal">{splitMethodLabel("equal")}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => remove(index)}
                    disabled={fields.length <= 1}
                  >
                    ×
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() =>
                  append({
                    itemId: "",
                    description: "",
                    amount: "0",
                    splitMethod: "by_quantity",
                  })
                }
              >
                {t("addLine")}
              </Button>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
