"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { AssetCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useFixedAssetSchema() {
  const t = useTranslations("FixedAssets");
  return z
    .object({
      name: z.string().min(1, t("validationNameRequired")),
      categoryId: z.string().min(1, t("validationCategoryRequired")),
      purchaseValue: z.string().min(1, t("validationPurchaseValueRequired")),
      salvageValue: z.string(),
      acquisitionDate: z.string().min(1, t("validationAcquisitionDateRequired")),
      inServiceDate: z.string().min(1, t("validationInServiceDateRequired")),
      invoiceLineId: z.string(),
    })
    .refine((data) => data.inServiceDate >= data.acquisitionDate, {
      message: t("validationInServiceDateOrder"),
      path: ["inServiceDate"],
    });
}
type FixedAssetValues = z.infer<ReturnType<typeof useFixedAssetSchema>>;

export function FixedAssetFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: (id: string) => void;
}) {
  const t = useTranslations("FixedAssets");
  const queryClient = useQueryClient();

  const categoriesQuery = useOrgListQuery<
    { assetCategories: AssetCategory[] },
    Record<string, never>
  >("assetCategories", (organizationId) =>
    getSwantaraService().assetCategories.list(organizationId),
  );
  const categories = categoriesQuery.data?.assetCategories ?? [];

  const schema = useFixedAssetSchema();

  const form = useForm<FixedAssetValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      categoryId: "",
      purchaseValue: "",
      salvageValue: "0",
      acquisitionDate: "",
      inServiceDate: "",
      invoiceLineId: "",
    },
  });

  const createMutation = useMutation({
    mutationFn: (values: FixedAssetValues) =>
      getSwantaraService().fixedAssets.create(Number(orgId), {
        organizationId: Number(orgId),
        name: values.name,
        categoryId: Number(values.categoryId),
        purchaseValue: Number(values.purchaseValue),
        salvageValue: values.salvageValue ? Number(values.salvageValue) : 0,
        acquisitionDate: values.acquisitionDate,
        inServiceDate: values.inServiceDate,
        invoiceLineId: values.invoiceLineId ? Number(values.invoiceLineId) : 0,
      }),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["fixedAssets", Number(orgId)] });
      onSave(String(result.fixedAsset.id));
      toast.success(t("toastAssetRegistered"));
    },
    onError: () => {
      toast.error(t("toastAssetFailed"));
    },
  });

  function handleSubmit(values: FixedAssetValues) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("registerAssetTitle")}
      description={t("registerAssetDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="categoryId" label={t("fieldCategory")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldCategory")}>
                <SelectValue placeholder={t("selectCategory")} />
              </SelectTrigger>
              <SelectContent>
                {categories.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {c.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="purchaseValue" label={t("fieldPurchaseValue")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="0.01" />}
        </FormField>
        <FormField name="salvageValue" label={t("fieldSalvageValue")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="0.01" />}
        </FormField>
        <FormField name="acquisitionDate" label={t("fieldAcquisitionDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="inServiceDate" label={t("fieldInServiceDate")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="invoiceLineId" label={t("fieldInvoiceLineId")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
