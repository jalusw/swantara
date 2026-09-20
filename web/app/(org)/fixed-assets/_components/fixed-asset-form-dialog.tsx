"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
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
  return z
    .object({
      name: z.string().min(1, "Name is required"),
      categoryId: z.string().min(1, "Category is required"),
      purchaseValue: z.string().min(1, "Purchase value is required"),
      salvageValue: z.string(),
      acquisitionDate: z.string().min(1, "Acquisition date is required"),
      inServiceDate: z.string().min(1, "In service date is required"),
      invoiceLineId: z.string(),
    })
    .refine((data) => data.inServiceDate >= data.acquisitionDate, {
      message: "In-service date must be on or after acquisition date",
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
      toast.success("Asset registered");
    },
    onError: () => {
      toast.error("Failed to register asset");
    },
  });

  function handleSubmit(values: FixedAssetValues) {
    createMutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={"Register fixed asset"}
      description={"Register a new asset from a supplier bill or manually."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="categoryId" label={"Category"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Category"}>
                <SelectValue placeholder={"Select category"} />
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
        <FormField name="purchaseValue" label={"Purchase value"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="0.01" />}
        </FormField>
        <FormField name="salvageValue" label={"Salvage value"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" step="0.01" />}
        </FormField>
        <FormField name="acquisitionDate" label={"Acquisition date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="inServiceDate" label={"In service date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
        <FormField name="invoiceLineId" label={"Invoice line ID"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
