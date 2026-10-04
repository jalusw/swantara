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
import type { Account, AssetCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const methods = ["linear", "declining", "declining_then_linear"] as const;
const periods = ["month", "year"] as const;

function useAssetCategorySchema() {
  const t = useTranslations("FixedAssets");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    method: z.enum(methods),
    methodNumber: z.string(),
    methodPeriod: z.enum(periods),
    assetAccountId: z.string(),
    depreciationAccountId: z.string(),
    expenseAccountId: z.string(),
    gainAccountId: z.string(),
    lossAccountId: z.string(),
  });
}
type AssetCategoryValues = z.infer<ReturnType<typeof useAssetCategorySchema>>;

export function AssetCategoryFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: AssetCategory | null;
  onSave: (id: string) => void;
}) {
  const t = useTranslations("FixedAssets");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();
  const isEdit = Boolean(initial);

  const accountsQuery = useOrgListQuery<{ accounts: Account[] }, Record<string, never>>(
    "accounts",
    (organizationId) => getSwantaraService().accounts.list(organizationId),
  );
  const accounts = accountsQuery.data?.accounts ?? [];

  const schema = useAssetCategorySchema();

  const form = useForm<AssetCategoryValues>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          method: (initial.method ?? "linear") as AssetCategoryValues["method"],
          methodNumber: initial.methodNumber != null ? String(initial.methodNumber) : "",
          methodPeriod: (initial.methodPeriod ?? "year") as AssetCategoryValues["methodPeriod"],
          assetAccountId: initial.assetAccountId != null ? String(initial.assetAccountId) : "",
          depreciationAccountId:
            initial.depreciationAccountId != null ? String(initial.depreciationAccountId) : "",
          expenseAccountId:
            initial.expenseAccountId != null ? String(initial.expenseAccountId) : "",
          gainAccountId: initial.gainAccountId != null ? String(initial.gainAccountId) : "",
          lossAccountId: initial.lossAccountId != null ? String(initial.lossAccountId) : "",
        }
      : {
          name: "",
          method: "linear",
          methodNumber: "",
          methodPeriod: "year",
          assetAccountId: "",
          depreciationAccountId: "",
          expenseAccountId: "",
          gainAccountId: "",
          lossAccountId: "",
        },
  });

  const createMutation = useMutation({
    mutationFn: (body: {
      organizationId: number;
      name: string;
      method: "linear" | "declining" | "declining_then_linear";
      methodNumber: number | null;
      methodPeriod: "month" | "year";
      assetAccountId: number | null;
      depreciationAccountId: number | null;
      expenseAccountId: number | null;
      gainAccountId: number | null;
      lossAccountId: number | null;
    }) => getSwantaraService().assetCategories.create(Number(orgId), body),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["assetCategories", Number(orgId)] });
      onSave(String(result.assetCategory.id));
      toast.success(t("toastCategoryCreated"));
    },
    onError: () => {
      toast.error(t("toastCategoryFailed"));
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      body,
    }: {
      categoryId: number;
      body: {
        organizationId: number;
        name: string;
        method: "linear" | "declining" | "declining_then_linear";
        methodNumber: number | null;
        methodPeriod: "month" | "year";
        assetAccountId: number | null;
        depreciationAccountId: number | null;
        expenseAccountId: number | null;
        gainAccountId: number | null;
        lossAccountId: number | null;
      };
    }) => getSwantaraService().assetCategories.create(Number(orgId), body),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: ["assetCategories", Number(orgId)] });
      onSave(String(variables.categoryId));
      toast.success(t("toastCategoryUpdated"));
    },
    onError: () => {
      toast.error(t("toastCategoryFailed"));
    },
  });

  function handleSubmit(values: AssetCategoryValues) {
    const body = {
      organizationId: Number(orgId),
      name: values.name,
      method: values.method,
      methodNumber: values.methodNumber ? Number(values.methodNumber) : null,
      methodPeriod: values.methodPeriod,
      assetAccountId: values.assetAccountId ? Number(values.assetAccountId) : null,
      depreciationAccountId: values.depreciationAccountId
        ? Number(values.depreciationAccountId)
        : null,
      expenseAccountId: values.expenseAccountId ? Number(values.expenseAccountId) : null,
      gainAccountId: values.gainAccountId ? Number(values.gainAccountId) : null,
      lossAccountId: values.lossAccountId ? Number(values.lossAccountId) : null,
    } as const;

    if (isEdit && initial) {
      updateMutation.mutate({ categoryId: initial.id, body });
    } else {
      createMutation.mutate(body);
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editCategoryTitle") : t("createCategoryTitle")}
      description={t("categoryFormDescription")}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending || updateMutation.isPending}
      className="max-h-[85vh] overflow-y-auto sm:max-w-lg"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={t("fieldName")}>
          {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
        </FormField>
        <FormField name="method" label={t("fieldDepreciationMethod")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldDepreciationMethod")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {methods.map((m) => (
                  <SelectItem key={m} value={m}>
                    {(t as unknown as (k: string) => string)(`categoryMethod_${m}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="methodNumber" label={t("fieldMethodNumber")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
        </FormField>
        <FormField name="methodPeriod" label={t("fieldPeriod")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldPeriod")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {periods.map((p) => (
                  <SelectItem key={p} value={p}>
                    {(t as unknown as (k: string) => string)(`methodPeriod_${p}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="assetAccountId" label={t("fieldAssetAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldAssetAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="depreciationAccountId" label={t("fieldDepreciationAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldDepreciationAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="expenseAccountId" label={t("fieldExpenseAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldExpenseAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="gainAccountId" label={t("fieldGainAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldGainAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="lossAccountId" label={t("fieldLossAccount")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldLossAccount")}>
                <SelectValue placeholder={t("selectAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
