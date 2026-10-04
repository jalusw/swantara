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
import type { ExpenseCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

type TFn = (key: string, values?: Record<string, string | number>) => string;

function useExpenseCategoryFormSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Expenses");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
  });
}

export function ExpenseCategoryFormDialog({
  open,
  onOpenChange,
  orgId,
  category,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  category: ExpenseCategory | null;
  onSave: () => void;
}) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Expenses");
  const tCommon = useTranslations("Common");
  const queryClient = useQueryClient();

  const schema = useExpenseCategoryFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: category?.name ?? "",
    },
  });

  const createMutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().expenseCategories.create(Number(orgId), {
        organizationId: Number(orgId),
        name: values.name,
        expenseAccountId: null,
        defaultTaxIds: [],
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["expenseCategories", Number(orgId)] });
      onSave();
      toast.success(t("toastCategoryCreated"));
    },
    onError: () => {
      toast.error(t("toastCategoryFailed"));
    },
  });

  const updateMutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().expenseCategories.update(Number(orgId), category!.id, {
        name: values.name,
        expenseAccountId: category!.expenseAccountId,
        defaultTaxIds: category!.defaultTaxIds,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["expenseCategories", Number(orgId)] });
      onSave();
      toast.success(t("toastCategoryUpdated"));
    },
    onError: () => {
      toast.error(t("toastCategoryFailed"));
    },
  });

  function handleSubmit(values: Values) {
    if (category) {
      updateMutation.mutate(values);
    } else {
      createMutation.mutate(values);
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={category ? t("editCategory") : t("newCategory")}
      description={t("categoriesDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={createMutation.isPending || updateMutation.isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="sm:max-w-lg"
    >
      <FormField name="name" label={t("fieldName")}>
        {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
      </FormField>
    </EntityFormDialog>
  );
}
