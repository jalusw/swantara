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
import type {
  Contact,
  CreateCrmLeadRequest,
  CrmLead,
  CrmStage,
  SalesGroup,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber, getLocalDateString } from "@/lib/utils";
import { CrmFormDivider, CrmFormSection } from "./crm-form-section";

function useLeadFormSchema(effectiveType: string) {
  const t = useTranslations("Crm");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    contactId: z.string(),
    contactName: z.string(),
    email: z.string().refine((v) => v === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v), {
      message: t("validationEmailInvalid"),
    }),
    phone: z.string(),
    jobPosition: z.string(),
    stageId: z.string().refine(
      (v) => {
        if (effectiveType === "opportunity" && v === "") return false;
        return true;
      },
      { message: t("validationStageRequired") },
    ),
    expectedRevenue: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0, {
      message: t("validationRevenueNegative"),
    }),
    probability: z
      .string()
      .refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0 && Number(v) <= 100, {
        message: t("validationProbabilityRange"),
      }),
    priority: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0, {
      message: t("validationPriorityNegative"),
    }),
    salespersonId: z.string(),
    salesGroupId: z.string(),
    source: z.string(),
    expectedClose: z.string(),
  });
}
type LeadFormValues = z.infer<ReturnType<typeof useLeadFormSchema>>;

export function LeadFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  type,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: CrmLead | null;
  type?: "lead" | "opportunity";
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);
  const effectiveType = initial?.type ?? type ?? "lead";
  const isOpportunity = effectiveType === "opportunity";
  const t = useTranslations("Crm");
  const queryClient = useQueryClient();

  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );
  const stagesQuery = useOrgListQuery<{ stages: CrmStage[] }, Record<string, never>>(
    "crmStages",
    (organizationId) => getSwantaraService().crmStages.list(organizationId),
  );
  const teamsQuery = useOrgListQuery<{ teams: SalesGroup[] }, Record<string, never>>(
    "salesGroups",
    (organizationId) => getSwantaraService().salesGroups.list(organizationId),
  );

  const contacts = contactsQuery.data?.contacts ?? [];
  const stages = stagesQuery.data?.stages ?? [];
  const teams = teamsQuery.data?.teams ?? [];

  const schema = useLeadFormSchema(effectiveType);

  const defaultValues: LeadFormValues = initial
    ? {
        name: initial.name,
        contactId: initial.contactId != null ? String(initial.contactId) : "",
        contactName: initial.contactName ?? "",
        email: initial.email ?? "",
        phone: initial.phone ?? "",
        jobPosition: initial.jobPosition ?? "",
        stageId: initial.stageId != null ? String(initial.stageId) : "",
        expectedRevenue: String(initial.expectedRevenue),
        probability: String(initial.probability),
        priority: String(initial.priority),
        salespersonId: initial.salespersonId != null ? String(initial.salespersonId) : "",
        salesGroupId: initial.salesGroupId != null ? String(initial.salesGroupId) : "",
        source: initial.source ?? "",
        expectedClose: initial.expectedClose
          ? getLocalDateString(new Date(initial.expectedClose))
          : "",
      }
    : {
        name: "",
        contactId: "",
        contactName: "",
        email: "",
        phone: "",
        jobPosition: "",
        stageId: "",
        expectedRevenue: "0",
        probability: "10",
        priority: "0",
        salespersonId: "",
        salesGroupId: "",
        source: "",
        expectedClose: "",
      };

  const form = useForm<LeadFormValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const watchedRevenue = form.watch("expectedRevenue");
  const watchedProbability = form.watch("probability");
  const watchedStageId = form.watch("stageId");
  const selectedStage = stages.find((s) => String(s.id) === watchedStageId);
  const weightedValue =
    !Number.isNaN(Number(watchedRevenue)) && !Number.isNaN(Number(watchedProbability))
      ? (Number(watchedRevenue) * Number(watchedProbability)) / 100
      : 0;

  const updateLeadMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmLeads.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success(t("updated"));
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const updateOpportunityMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmOpportunities.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success(t("updated"));
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const createLeadMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmLeads.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success(t("created"));
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const createOpportunityMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmOpportunities.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success(t("created"));
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const isPending =
    updateLeadMutation.isPending ||
    updateOpportunityMutation.isPending ||
    createLeadMutation.isPending ||
    createOpportunityMutation.isPending;

  function handleSubmit(values: LeadFormValues) {
    const payload = {
      organizationId: Number(orgId),
      name: values.name,
      type: effectiveType as "lead" | "opportunity",
      contactId: values.contactId ? Number(values.contactId) : null,
      contactName: values.contactName || null,
      email: values.email || null,
      phone: values.phone || null,
      jobPosition: values.jobPosition || null,
      stageId: values.stageId ? Number(values.stageId) : null,
      expectedRevenue: Number(values.expectedRevenue),
      probability: Number(values.probability),
      priority: Number(values.priority),
      salespersonId: values.salespersonId ? Number(values.salespersonId) : null,
      salesGroupId: values.salesGroupId ? Number(values.salesGroupId) : null,
      source: values.source || null,
      medium: null,
      campaign: null,
      expectedClose: values.expectedClose || null,
    };

    if (isEdit && initial) {
      if (initial.type === "lead") {
        updateLeadMutation.mutate(payload);
      } else {
        updateOpportunityMutation.mutate(payload);
      }
    } else {
      if (effectiveType === "lead") {
        createLeadMutation.mutate(payload);
      } else {
        createOpportunityMutation.mutate(payload);
      }
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={
        isEdit
          ? isOpportunity
            ? t("editOpportunity")
            : t("editLead")
          : isOpportunity
            ? t("newOpportunity")
            : t("newLead")
      }
      description={isOpportunity ? t("opportunityFormDescription") : t("leadFormDescription")}
      badge={
        <Badge variant={isOpportunity ? "default" : "secondary"}>
          {isOpportunity ? t("opportunityBadge") : t("leadBadge")}
        </Badge>
      }
      form={form}
      onSubmit={handleSubmit}
      isPending={isPending}
      submitLabel={
        isEdit ? t("saveChanges") : isOpportunity ? t("createOpportunity") : t("createLead")
      }
      footerHint={isOpportunity ? t("stageForecastHint") : undefined}
      className="max-h-[85vh] sm:max-w-2xl"
    >
      <CrmFormSection
        title={isOpportunity ? t("sectionDeal") : t("sectionLeadDetails")}
        description={isOpportunity ? t("sectionDealDescription") : t("sectionLeadDescription")}
      >
        <FormField
          name="name"
          label={t("fieldName")}
          description={t("fieldNameHint")}
          className="sm:col-span-2"
        >
          {({ field, id }) => (
            <Input {...field} id={id} placeholder={t("fieldName")} autoFocus autoComplete="off" />
          )}
        </FormField>
        <FormField
          name="source"
          label={t("fieldSource")}
          description={isOpportunity ? t("sourceOpportunityHint") : t("sourceLeadHint")}
        >
          {({ field, id }) => (
            <Input {...field} id={id} placeholder={t("sourcePlaceholder")} autoComplete="off" />
          )}
        </FormField>
        <FormField
          name="expectedClose"
          label={t("tableExpectedClose")}
          description={t("expectedCloseHint")}
        >
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection title={t("sectionContact")} description={t("sectionContactDescription")}>
        <FormField name="contactId" label={t("tableContact")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableContact")}>
                <SelectValue placeholder={t("noContact")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("noContact")}</SelectItem>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField
          name="contactName"
          label={t("fieldContactName")}
          description={t("contactNameHint")}
        >
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              placeholder={t("contactNamePlaceholder")}
              autoComplete="off"
            />
          )}
        </FormField>
        <FormField name="email" label={t("fieldEmail")}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="email"
              placeholder={t("emailPlaceholder")}
              autoComplete="email"
            />
          )}
        </FormField>
        <FormField name="phone" label={t("fieldPhone")}>
          {({ field, id }) => (
            <Input {...field} id={id} placeholder={t("phonePlaceholder")} autoComplete="tel" />
          )}
        </FormField>
        <FormField name="jobPosition" label={t("fieldJobPosition")} className="sm:col-span-2">
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              placeholder={t("jobPositionPlaceholder")}
              autoComplete="off"
            />
          )}
        </FormField>
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection
        title={isOpportunity ? t("sectionPipeline") : t("sectionQualification")}
        description={isOpportunity ? t("pipelineHint") : t("qualificationHint")}
      >
        <FormField
          name="stageId"
          label={t("tableStage")}
          description={
            selectedStage
              ? t("stageDefaultHint", { probability: selectedStage.probability })
              : isOpportunity
                ? t("stageRequiredHint")
                : t("stageOptionalHint")
          }
          className="sm:col-span-2"
        >
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableStage")}>
                <SelectValue placeholder={t("noStage")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("noStage")}</SelectItem>
                {stages.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.name} — {s.probability}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="expectedRevenue" label={t("tableExpectedRevenue")}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" step="any" min="0" inputMode="decimal" />
          )}
        </FormField>
        <FormField
          name="probability"
          label={t("probabilityLabel")}
          description={t("probabilityHint")}
        >
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" max="100" />}
        </FormField>
        <div className="rounded-lg border bg-muted/40 p-3 sm:col-span-2">
          <div className="flex items-baseline justify-between gap-2">
            <p className="text-xs font-medium text-muted-foreground">{t("weightedValue")}</p>
            <p className="text-sm font-semibold tabular-nums">{formatNumber(weightedValue)}</p>
          </div>
          <Progress
            value={Number.isFinite(Number(watchedProbability)) ? Number(watchedProbability) : 0}
            className="mt-2"
            aria-label={t("tableProbability")}
          />
          <p className="mt-1.5 text-xs text-muted-foreground">{t("weightedHint")}</p>
        </div>
        <FormField name="priority" label={t("fieldPriority")} description={t("priorityHint")}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
        </FormField>
        <div className="hidden sm:block" aria-hidden />
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection title={t("sectionOwnership")} description={t("ownershipHint")}>
        <FormField name="salespersonId" label={t("fieldSalesperson")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldSalesperson")}>
                <SelectValue placeholder={t("unassigned")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("unassigned")}</SelectItem>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="salesGroupId" label={t("fieldSalesTeam")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldSalesTeam")}>
                <SelectValue placeholder={t("noTeam")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("noTeam")}</SelectItem>
                {teams.map((team) => (
                  <SelectItem key={team.id} value={String(team.id)}>
                    {team.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
      </CrmFormSection>
    </EntityFormDialog>
  );
}
