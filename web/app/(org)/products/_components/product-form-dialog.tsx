"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useState } from "react";
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
import { Switch } from "@/components/switch";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Unit } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

import { AttributeEditor, type AttributeRow } from "./attribute-editor";
import { productTypes, trackingModes } from "./products-data";
import { buildMatrix, generateSku, VariantMatrixPreview } from "./variant-matrix-preview";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function useProductSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  return z.object({
    name: z.string().trim().min(1, t("validationNameRequired")),
    categoryId: z.string().trim(),
    type: z.enum(productTypes),
    tracking: z.enum(trackingModes),
    unitId: z.string().trim(),
    purchaseUnitId: z.string().trim(),
    listPrice: z
      .string()
      .trim()
      .refine((value) => Number(value) >= 0, t("validationNonNegative")),
    standardCost: z
      .string()
      .trim()
      .refine((value) => Number(value) >= 0, t("validationNonNegative")),
    isPurchasable: z.boolean(),
    isSellable: z.boolean(),
    isManufactured: z.boolean(),
  });
}
type Values = z.infer<ReturnType<typeof useProductSchema>>;

export function ProductFormDialog({
  open,
  onOpenChange,
  orgId,
  categories,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  categories: { id: string; name: string }[];
  initial?: {
    id: string;
    name: string;
    categoryId: string | null;
    type: Values["type"];
    tracking: Values["tracking"];
    unitId: string | null;
    purchaseUnitId: string | null;
    listPrice: number;
    standardCost: number;
    isPurchasable: boolean;
    isSellable: boolean;
    isManufactured: boolean;
  };
  onSave: (itemId: string) => void;
}) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const tCommon = useTranslations("Common");
  const schema = useProductSchema();
  const unitsQuery = useOrgListQuery<{ units: Unit[] }, Record<string, never>>("units", () =>
    getSwantaraService().units.list(),
  );
  const uomOptions = (unitsQuery.data?.units ?? []).map((unit) => ({
    id: String(unit.id),
    name: unit.name,
  }));

  const [attributes, setAttributes] = useState<AttributeRow[]>([]);

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: toValues(initial ?? emptyTemplate()),
  });

  const templateName = form.watch("name");
  const matrix = buildMatrix(attributes);
  const generatedSkus = matrix.map((combination) => generateSku(templateName || "", combination));
  const hasDuplicateSkus = new Set(generatedSkus).size !== generatedSkus.length;

  function handleSubmit(values: Values) {
    if (hasDuplicateSkus) {
      toast.error(t("duplicateSkuError"));
      return;
    }

    const request = {
      organizationId: Number(orgId) || 0,
      name: values.name,
      categoryId: values.categoryId === "" ? null : Number(values.categoryId) || null,
      type: values.type,
      unitId: values.unitId === "" ? null : Number(values.unitId) || null,
      purchaseUnitId: values.purchaseUnitId === "" ? null : Number(values.purchaseUnitId) || null,
      listPrice: Number(values.listPrice) || 0,
      standardCost: Number(values.standardCost) || 0,
      isPurchasable: values.isPurchasable,
      isSellable: values.isSellable,
      isManufactured: values.isManufactured,
      tracking: values.tracking,
      weight: 0,
      volume: 0,
      hsCode: null,
      descriptionSale: null,
      descriptionPurchase: null,
      active: true,
      variants: matrix.map((combination, index) => ({
        sku: generatedSkus[index] || null,
        barcode: null,
        attributeJson: combination,
        extraCost: 0,
        active: true,
      })),
      attributeMatrix: attributes.map((attr) => ({
        name: attr.name,
        values: attr.values,
      })),
    };

    return getSwantaraService()
      .products.create(Number(orgId) || 0, request)
      .then(({ item }) => {
        toast.success(t("toastItemSaved"));
        onSave(String(item.id));
      })
      .catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{initial ? t("editItem") : t("newItem")}</DialogTitle>
          <DialogDescription>{t("itemDialogDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={t("fieldName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
            </FormField>
            <FormField name="categoryId" label={t("fieldCategory")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldCategory")}>
                    <SelectValue placeholder={t("noCategory")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">{t("noCategory")}</SelectItem>
                    {categories.map((category) => (
                      <SelectItem key={category.id} value={category.id}>
                        {category.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="type" label={t("fieldType")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldType")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {productTypes.map((type) => (
                      <SelectItem key={type} value={type}>
                        {(t as unknown as (k: string) => string)(`productType_${type}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="tracking" label={t("fieldTracking")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldTracking")}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {trackingModes.map((mode) => (
                      <SelectItem key={mode} value={mode}>
                        {(t as unknown as (k: string) => string)(`trackingMode_${mode}`)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="listPrice" label={t("colListPrice")}>
              {({ field, id }) => (
                <Input {...field} id={id} type="number" step="any" min="0" inputMode="decimal" />
              )}
            </FormField>
            <FormField name="standardCost" label={t("colStandardCost")}>
              {({ field, id }) => (
                <Input {...field} id={id} type="number" step="any" min="0" inputMode="decimal" />
              )}
            </FormField>
            <FormField name="unitId" label={t("fieldUnit")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldUnit")}>
                    <SelectValue placeholder={t("fieldUnit")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">—</SelectItem>
                    {uomOptions.map((unit) => (
                      <SelectItem key={unit.id} value={unit.id}>
                        {unit.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="purchaseUnitId" label={t("fieldPurchaseUnit")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("fieldPurchaseUnit")}>
                    <SelectValue placeholder={t("fieldPurchaseUnit")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">—</SelectItem>
                    {uomOptions.map((unit) => (
                      <SelectItem key={unit.id} value={unit.id}>
                        {unit.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="isSellable" label={t("canBeSold")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("canBeSold")}
                />
              )}
            </FormField>
            <FormField name="isPurchasable" label={t("canBePurchased")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("canBePurchased")}
                />
              )}
            </FormField>
            <FormField name="isManufactured" label={t("isManufactured")}>
              {({ field }) => (
                <Switch
                  checked={field.value}
                  onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                  aria-label={t("isManufactured")}
                />
              )}
            </FormField>
          </div>

          <div className="flex flex-col gap-3 rounded-md border p-4">
            <div className="flex flex-col gap-1">
              <p className="text-sm">{t("attributesVariants")}</p>
              <p className="text-xs text-muted-foreground">{t("attributesVariantsDescription")}</p>
            </div>
            <AttributeEditor value={attributes} onChange={setAttributes} />
            <VariantMatrixPreview templateName={templateName || ""} attributes={attributes} />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{t("saveItem")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

function emptyTemplate() {
  return {
    id: "",
    name: "",
    categoryId: null,
    type: "stockable" as const,
    unitId: null,
    purchaseUnitId: null,
    listPrice: 0,
    standardCost: 0,
    isPurchasable: true,
    isSellable: true,
    isManufactured: false,
    tracking: "none" as const,
  };
}

function toValues(
  template: NonNullable<React.ComponentProps<typeof ProductFormDialog>["initial"]>,
) {
  return {
    name: template.name,
    categoryId: template.categoryId ?? "",
    type: template.type,
    tracking: template.tracking,
    unitId: template.unitId ?? "",
    purchaseUnitId: template.purchaseUnitId ?? "",
    listPrice: String(template.listPrice),
    standardCost: String(template.standardCost),
    isPurchasable: template.isPurchasable,
    isSellable: template.isSellable,
    isManufactured: template.isManufactured,
  };
}
