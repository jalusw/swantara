"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { CheckSquare, Mail, Phone, StickyNote, Users } from "lucide-react";
import { useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Textarea } from "@/components/textarea";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  CreateCrmActivityRequest,
  CrmActivity,
  CrmLead,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { cn, getLocalDateString } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { CrmFormDivider, CrmFormSection } from "./crm-form-section";

const activityTypes = ["call", "meeting", "email", "note", "task"] as const;

const activityMeta: Record<(typeof activityTypes)[number], { icon: typeof Phone; hint: string }> = {
  call: { icon: Phone, hint: "Log a call" },
  meeting: { icon: Users, hint: "Log a meeting" },
  email: { icon: Mail, hint: "Log an email" },
  note: { icon: StickyNote, hint: "Quick note" },
  task: { icon: CheckSquare, hint: "Follow-up task" },
};

function useActivityFormSchema() {
  const t = useTranslations("Crm");
  return z.object({
    prospectId: z.string(),
    contactId: z.string(),
    type: z.enum(activityTypes),
    summary: z.string().min(1, t("validationSummaryRequired")),
    note: z.string(),
    dueDate: z.string(),
  });
}
type ActivityFormValues = z.infer<ReturnType<typeof useActivityFormSchema>>;

export function ActivityFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: CrmActivity | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);
  const t = useTranslations("Crm");
  const queryClient = useQueryClient();

  const leadsQuery = useOrgListQuery<{ leads: CrmLead[] }, Record<string, never>>(
    "crmLeads",
    (organizationId) => getSwantaraService().crmLeads.list(organizationId),
  );
  const opportunitiesQuery = useOrgListQuery<{ opportunities: CrmLead[] }, Record<string, never>>(
    "crmOpportunities",
    (organizationId) => getSwantaraService().crmOpportunities.list(organizationId),
  );
  const contactsQuery = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const leads = [
    ...(leadsQuery.data?.leads ?? []),
    ...(opportunitiesQuery.data?.opportunities ?? []),
  ];
  const contacts = contactsQuery.data?.contacts ?? [];

  const schema = useActivityFormSchema();

  const defaultValues: ActivityFormValues = initial
    ? {
        prospectId: initial.prospectId != null ? String(initial.prospectId) : "",
        contactId: initial.contactId != null ? String(initial.contactId) : "",
        type: initial.type,
        summary: initial.summary,
        note: initial.note ?? "",
        dueDate: initial.dueDate ? getLocalDateString(new Date(initial.dueDate)) : "",
      }
    : {
        prospectId: "",
        contactId: "",
        type: "note",
        summary: "",
        note: "",
        dueDate: "",
      };

  const form = useForm<ActivityFormValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const selectedType = form.watch("type");

  const updateMutation = useMutation({
    mutationFn: (payload: CreateCrmActivityRequest) =>
      getSwantaraService().crmActivities.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success(t("updated"));
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const createMutation = useMutation({
    mutationFn: (payload: CreateCrmActivityRequest) =>
      getSwantaraService().crmActivities.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success(t("created"));
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
      onSave();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  const isPending = updateMutation.isPending || createMutation.isPending;

  function handleSubmit(values: ActivityFormValues) {
    const payload = {
      prospectId: values.prospectId ? Number(values.prospectId) : null,
      contactId: values.contactId ? Number(values.contactId) : null,
      type: values.type,
      summary: values.summary,
      note: values.note || null,
      dueDate: values.dueDate || null,
      done: initial?.done ?? null,
    };
    if (isEdit && initial) {
      updateMutation.mutate(payload);
    } else {
      createMutation.mutate(payload);
    }
  }

  function activityTypeLabel(type: string): string {
    try {
      return (t as unknown as (k: string) => string)(`activityType.${type}`);
    } catch {
      return humanizeKey(String(type));
    }
  }

  function activityTypeHint(type: (typeof activityTypes)[number]): string {
    try {
      return (t as unknown as (k: string) => string)(`activityHint.${type}`);
    } catch {
      return activityMeta[type].hint;
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? t("editActivity") : t("newActivity")}
      description={t("activitiesSubtitle")}
      badge={<Badge variant="secondary">{activityTypeLabel(String(selectedType))}</Badge>}
      form={form}
      onSubmit={handleSubmit}
      isPending={isPending}
      submitLabel={isEdit ? t("saveChanges") : t("addActivity")}
      className="sm:max-w-xl"
    >
      <CrmFormSection title={t("sectionActivity")} description={t("sectionActivityDescription")}>
        <fieldset className="sm:col-span-2">
          <legend className="mb-2 text-sm font-medium">{t("tableType")}</legend>
          <div className="grid grid-cols-5 gap-1.5">
            {activityTypes.map((type) => {
              const Icon = activityMeta[type].icon;
              const active = selectedType === type;
              return (
                <button
                  key={type}
                  type="button"
                  aria-pressed={active}
                  aria-label={activityTypeLabel(String(type))}
                  title={activityTypeHint(type)}
                  onClick={() =>
                    form.setValue("type", type, { shouldDirty: true, shouldValidate: true })
                  }
                  className={cn(
                    "flex flex-col items-center gap-1 rounded-lg border px-1 py-2.5 text-xs transition-colors",
                    active
                      ? "border-primary bg-primary/5 text-foreground ring-1 ring-primary/30"
                      : "border-border bg-background text-muted-foreground hover:border-foreground/20 hover:text-foreground",
                  )}
                >
                  <Icon className="size-4" aria-hidden />
                  <span className="leading-none">{activityTypeLabel(String(type))}</span>
                </button>
              );
            })}
          </div>
          {/* Keep the native select in sync for assistive tech + existing tests */}
          <FormField name="type" className="sr-only">
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange}>
                <SelectTrigger id={id} aria-label={t("tableType")} className="sr-only">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {activityTypes.map((type) => (
                    <SelectItem key={type} value={type}>
                      {activityTypeLabel(String(type))}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
        </fieldset>
        <FormField
          name="summary"
          label={t("tableSummary")}
          description={t("summaryHint")}
          className="sm:col-span-2"
        >
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              autoFocus
              autoComplete="off"
              placeholder={t("summaryPlaceholder")}
            />
          )}
        </FormField>
        <FormField name="dueDate" label={t("tableDueDate")} className="sm:col-span-2">
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection title={t("sectionRelated")} description={t("relatedHint")}>
        <FormField name="prospectId" label={t("fieldLeadOpportunity")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("fieldLeadOpportunity")}>
                <SelectValue placeholder={t("noLink")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("noLink")}</SelectItem>
                {leads.map((lead) => (
                  <SelectItem key={lead.id} value={String(lead.id)}>
                    {lead.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="contactId" label={t("tableContact")}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={t("tableContact")}>
                <SelectValue placeholder={t("noContact")} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{t("noContact")}</SelectItem>
                {contacts.map((contact) => (
                  <SelectItem key={contact.id} value={String(contact.id)}>
                    {contact.displayName || contact.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
      </CrmFormSection>

      <CrmFormDivider />

      <CrmFormSection title={t("sectionNotes")} description={t("notesHint")}>
        <FormField name="note" label={t("fieldNote")} className="sm:col-span-2">
          {({ field, id }) => (
            <Textarea {...field} id={id} rows={4} placeholder={t("notePlaceholder")} />
          )}
        </FormField>
      </CrmFormSection>
    </EntityFormDialog>
  );
}
