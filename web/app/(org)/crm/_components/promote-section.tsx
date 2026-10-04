"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Progress } from "@/components/progress";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CrmStage, PromoteLeadRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { CrmFormDivider, CrmFormSection } from "./crm-form-section";

function usePromoteFormSchema() {
  const t = useTranslations("Crm");
  return z.object({
    stageId: z.string().min(1, t("validationStageRequired")),
    expectedRevenue: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0),
    probability: z
      .string()
      .refine((v) => v === "" || (!Number.isNaN(Number(v)) && Number(v) >= 0 && Number(v) <= 100)),
    priority: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0),
    expectedClose: z.string(),
  });
}
type PromoteValues = z.infer<ReturnType<typeof usePromoteFormSchema>>;

export function PromoteDialog({
  open,
  onOpenChange,
  orgId,
  prospectId,
  onPromoted,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  prospectId: string;
  onPromoted: () => void;
}) {
  const t = useTranslations("Crm");
  const queryClient = useQueryClient();
  const stagesQuery = useOrgListQuery<{ stages: CrmStage[] }, Record<string, never>>(
    "crmStages",
    (organizationId) => getSwantaraService().crmStages.list(organizationId),
  );
  const stages = stagesQuery.data?.stages ?? [];

  const schema = usePromoteFormSchema();

  const form = useForm<PromoteValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      stageId: "",
      expectedRevenue: "0",
      probability: "",
      priority: "0",
      expectedClose: "",
    },
  });

  const watchedStageId = form.watch("stageId");
  const watchedRevenue = form.watch("expectedRevenue");
  const watchedProbability = form.watch("probability");
  const selectedStage = stages.find((s) => String(s.id) === watchedStageId);
  const effectiveProbability =
    watchedProbability === "" ? selectedStage?.probability : Number(watchedProbability);
  const weightedValue =
    !Number.isNaN(Number(watchedRevenue)) &&
    effectiveProbability != null &&
    !Number.isNaN(Number(effectiveProbability))
      ? (Number(watchedRevenue) * Number(effectiveProbability)) / 100
      : 0;

  const promoteMutation = useMutation({
    mutationFn: (payload: PromoteLeadRequest) =>
      getSwantaraService().crmLeads.promote(Number(orgId), Number(prospectId), payload),
    onSuccess: () => {
      toast.success(t("leadPromoted"));
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onPromoted();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  function handleSubmit(values: PromoteValues) {
    const payload = {
      stageId: Number(values.stageId),
      expectedRevenue: Number(values.expectedRevenue),
      probability: values.probability ? Number(values.probability) : null,
      priority: Number(values.priority),
      salespersonId: null,
      salesGroupId: null,
      expectedClose: values.expectedClose || null,
    };
    promoteMutation.mutate(payload);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("promoteTitle")}
      description={t("promoteDescription")}
      badge={<Badge variant="default">{t("opportunityBadge")}</Badge>}
      form={form}
      onSubmit={handleSubmit}
      isPending={promoteMutation.isPending}
      submitLabel={t("promoteSubmit")}
      footerHint={
        selectedStage
          ? t("promoteStageHint", { probability: selectedStage.probability })
          : t("promoteEmptyHint")
      }
      className="sm:max-w-xl"
    >
      <CrmFormSection title={t("sectionPlacement")} description={t("placementHint")}>
        <FormField
          name="stageId"
          label={t("tableStage")}
          description={
            selectedStage
              ? t("stageDefaultHint", { probability: selectedStage.probability })
              : t("stageRequiredHint")
          }
          className="sm:col-span-2"
        >
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableStage")}>
                <SelectValue placeholder={t("selectStage")} />
              </SelectTrigger>
              <SelectContent>
                {stages.map((stage) => (
                  <SelectItem key={stage.id} value={String(stage.id)}>
                    {stage.name} — {stage.probability}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection title={t("sectionDealValue")} description={t("dealValueHint")}>
        <FormField name="expectedRevenue" label={t("tableExpectedRevenue")}>
          {({ field, id }) => <Input {...field} id={id} type="number" step="any" min="0" />}
        </FormField>
        <FormField
          name="probability"
          label={t("probabilityLabel")}
          description={t("probabilityEmptyHint")}
        >
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="number"
              min="0"
              max="100"
              placeholder={t("probabilityAutoPlaceholder")}
            />
          )}
        </FormField>
        <div className="rounded-lg border bg-muted/40 p-3 sm:col-span-2">
          <div className="flex items-baseline justify-between gap-2">
            <p className="text-xs font-medium text-muted-foreground">{t("weightedValue")}</p>
            <p className="text-sm font-semibold tabular-nums">{formatNumber(weightedValue)}</p>
          </div>
          <Progress value={Number(effectiveProbability) || 0} className="mt-2" />
        </div>
        <FormField name="priority" label={t("fieldPriority")} description={t("priorityHint")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
        </FormField>
        <FormField name="expectedClose" label={t("tableExpectedClose")}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
      </CrmFormSection>
    </EntityFormDialog>
  );
}
