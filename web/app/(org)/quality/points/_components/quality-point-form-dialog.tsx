"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, QualityPoint } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

const qualityPointSchema = z.object({
  itemId: z.coerce.number().nullable(),
  operation: z.string().nullable(),
  testType: z.enum(["pass_fail", "measure", "instruction"]),
  normMin: z.coerce.number().nullable(),
  normMax: z.coerce.number().nullable(),
  unitId: z.coerce.number().nullable(),
});
type QualityPointValues = z.infer<typeof qualityPointSchema>;

export function QualityPointFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: QualityPoint | null;
  onSave: () => void;
}) {
  const queryClient = useQueryClient();
  const isEdit = Boolean(initial);

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId, _params) => getSwantaraService().products.list(organizationId),
  );

  const unitsQuery = useQuery({
    queryKey: ["units"],
    queryFn: () => getSwantaraService().units.list(),
  });

  const productOptions = (productsQuery.data?.products ?? []).map((p) => ({
    id: String(p.id),
    name: p.name,
  }));

  const uomOptions = (unitsQuery.data?.units ?? []).map((u) => ({
    id: String(u.id),
    name: u.name,
  }));

  const form = useForm<QualityPointValues>({
    resolver: zodResolver(qualityPointSchema),
    defaultValues: {
      itemId: initial?.itemId ?? null,
      operation: initial?.operation ?? null,
      testType: initial?.testType ?? "pass_fail",
      normMin: initial?.normMin ?? null,
      normMax: initial?.normMax ?? null,
      unitId: initial?.unitId ?? null,
    },
  });

  const createMutation = useMutation({
    mutationFn: (
      payload: Omit<QualityPointValues, "organizationId"> & { organizationId: number },
    ) => getSwantaraService().qualityPoints.create(Number(orgId), payload),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["qualityPoints", Number(orgId)] });
      onSave();
      toast.success("Quality point saved.");
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  function handleSubmit(values: QualityPointValues) {
    const payload = {
      organizationId: Number(orgId),
      itemId: values.itemId,
      operation: values.operation,
      testType: values.testType,
      normMin: values.normMin,
      normMax: values.normMax,
      unitId: values.unitId,
    };

    createMutation.mutate(payload);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit quality point" : "New quality point"}
      description={"Define where and how quality is checked."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="itemId" label={"Item"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Item"}>
                <SelectValue placeholder={"Select item"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {productOptions.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="operation" label={"Operation"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value || null)}
            />
          )}
        </FormField>
        <FormField name="testType" label={"Test type"}>
          {({ field, id }) => (
            <Select
              value={field.value ?? ""}
              onValueChange={(value) =>
                field.onChange(value as "pass_fail" | "measure" | "instruction")
              }
            >
              <SelectTrigger id={id} aria-label={"Test type"}>
                <SelectValue placeholder={"Test type"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="pass_fail">{"Pass / Fail"}</SelectItem>
                <SelectItem value="measure">{"Measurement"}</SelectItem>
                <SelectItem value="instruction">{"Instruction"}</SelectItem>
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="normMin" label={"Min value"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="number"
              step="0.01"
              inputMode="decimal"
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value ? Number(e.target.value) : null)}
            />
          )}
        </FormField>
        <FormField name="normMax" label={"Max value"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="number"
              step="0.01"
              inputMode="decimal"
              value={field.value ?? ""}
              onChange={(e) => field.onChange(e.target.value ? Number(e.target.value) : null)}
            />
          )}
        </FormField>
        <FormField name="unitId" label={"Unit of Measure"}>
          {({ field, id }) => (
            <Select
              value={field.value ? String(field.value) : ""}
              onValueChange={(value) => field.onChange(value ? Number(value) : null)}
            >
              <SelectTrigger id={id} aria-label={"Unit of Measure"}>
                <SelectValue placeholder={"Select UoM"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">—</SelectItem>
                {uomOptions.map((u) => (
                  <SelectItem key={u.id} value={u.id}>
                    {u.name}
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
