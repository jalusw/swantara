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
import { Textarea } from "@/components/textarea";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type {
  Contact,
  CreateCrmActivityRequest,
  CrmActivity,
  CrmLead,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

const activityTypes = ["call", "meeting", "email", "note", "task"] as const;

function useActivityFormSchema() {
  return z.object({
    prospectId: z.string(),
    contactId: z.string(),
    type: z.enum(activityTypes),
    summary: z.string().min(1, "Summary"),
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

  const updateMutation = useMutation({
    mutationFn: (payload: CreateCrmActivityRequest) =>
      getSwantaraService().crmActivities.update(Number(orgId), initial!.id, payload),
    onSuccess: () => {
      toast.success("Updated.");
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
      onSave();
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const createMutation = useMutation({
    mutationFn: (payload: CreateCrmActivityRequest) =>
      getSwantaraService().crmActivities.create(Number(orgId), payload),
    onSuccess: () => {
      toast.success("Created.");
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
      onSave();
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

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

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit activity" : "Add activity"}
      description={"Calls, meetings, emails and tasks tied to a lead or contact."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="prospectId" label={"Lead / opportunity"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Lead / opportunity"}>
                <SelectValue placeholder={"No contact"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No contact"}</SelectItem>
                {leads.map((lead) => (
                  <SelectItem key={lead.id} value={String(lead.id)}>
                    {lead.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="contactId" label={"Contact"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Contact"}>
                <SelectValue placeholder={"No contact"} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">{"No contact"}</SelectItem>
                {contacts.map((contact) => (
                  <SelectItem key={contact.id} value={String(contact.id)}>
                    {contact.displayName || contact.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="type" label={"Type"}>
          {({ field, id }) => (
            <Select value={field.value} onValueChange={field.onChange}>
              <SelectTrigger id={id} aria-label={"Type"}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {activityTypes.map((type) => (
                  <SelectItem key={type} value={type}>
                    {humanizeKey(String(type))}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </FormField>
        <FormField name="summary" label={"Summary"}>
          {({ field, id }) => <Input {...field} id={id} />}
        </FormField>
        <FormField name="note" label={"Note"}>
          {({ field, id }) => <Textarea {...field} id={id} />}
        </FormField>
        <FormField name="dueDate" label={"Due date"}>
          {({ field, id }) => <Input {...field} id={id} type="date" />}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
