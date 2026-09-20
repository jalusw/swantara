"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import type { ExpenseCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function useExpenseCategoryFormSchema() {
  return z.object({
    name: z.string().min(1, "Name is required"),
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
      toast.success("Category created");
    },
    onError: () => {
      toast.error("Failed to save category");
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
      toast.success("Category updated");
    },
    onError: () => {
      toast.error("Failed to save category");
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
      title={category ? "Edit expense category" : "Create expense category"}
      description={"Manage expense categories and default accounts."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <FormField name="name" label={"Name"}>
        {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
      </FormField>
    </EntityFormDialog>
  );
}
