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
import type {
  Contact,
  CreateCrmLeadRequest,
  CrmLead,
  CrmStage,
  SalesGroup,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";

function useLeadFormSchema(effectiveType: string) {
  return z.object({
    name: z.string().min(1, "Enter a name."),
    contactId: z.string(),
    contactName: z.string(),
    email: z.string().refine((v) => v === "" || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(v), {
      message: "Enter a valid email.",
    }),
    phone: z.string(),
    jobPosition: z.string(),
    stageId: z.string().refine(
      (v) => {
        if (effectiveType === "opportunity" && v === "") return false;
        return true;
      },
      { message: "An opportunity must have a stage." },
    ),
    expectedRevenue: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0, {
      message: "Expected revenue cannot be negative.",
    }),
    probability: z
      .string()
      .refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0 && Number(v) <= 100, {
        message: "Probability must be 0–100.",
      }),
    priority: z.string().refine((v) => !Number.isNaN(Number(v)) && Number(v) >= 0, {
      message: "Priority cannot be negative.",
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

  const updateLeadMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmLeads.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success("Updated.");
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  const updateOpportunityMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmOpportunities.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success("Updated.");
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  const createLeadMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmLeads.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success("Created.");
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

  const createOpportunityMutation = useMutation({
    mutationFn: (payload: CreateCrmLeadRequest) =>
      getSwantaraService().crmOpportunities.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success("Created.");
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onSave();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
    },
  });

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
          ? effectiveType === "opportunity"
            ? "Edit opportunity"
            : "Edit lead"
          : effectiveType === "opportunity"
            ? "New opportunity"
            : "New lead"
      }
      description={"Details that identify this record."}
      form={form}
      onSubmit={handleSubmit}
      className="max-h-[85vh] overflow-y-auto sm:max-w-2xl"
    >
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="contactId" label={"Contact"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Contact"}>
                <SelectValue placeholder={"No contact"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No contact"}</SelectItem>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="contactName" label={"Contact name"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="email" label={"Email"}>
          {({ field, id }) => <Input {...field} id={id} type="email" />}
        </FormField>
        <FormField name="phone" label={"Phone"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="jobPosition" label={"Job position"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="stageId" label={"Stage"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Stage"}>
                <SelectValue placeholder={"No stage"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No stage"}</SelectItem>
                {stages.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    {s.name} — {s.probability}%
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="expectedRevenue" label={"Expected revenue"}>
          {({ field, id }) => (
            <Input {...field} id={id} type="number" step="any" min="0" inputMode="decimal" />
          )}
        </FormField>
        <FormField name="probability" label={"Probability %"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" max="100" />}
        </FormField>
        <FormField name="priority" label={"Priority"}>
          {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
        </FormField>
        <FormField name="salespersonId" label={"Salesperson"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Salesperson"}>
                <SelectValue placeholder={"No contact"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No contact"}</SelectItem>
                {contacts.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.displayName || p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="salesGroupId" label={"Sales team"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Sales team"}>
                <SelectValue placeholder={"No contact"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No contact"}</SelectItem>
                {teams.map((team) => (
                  <SelectItem key={team.id} value={String(team.id)}>
                    {team.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="source" label={"Source"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="expectedClose" label={"Expected close"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
