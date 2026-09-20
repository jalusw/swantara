"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CrmStage, PromoteLeadRequest } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

function usePromoteFormSchema() {
  return z.object({
    stageId: z.string().min(1, "Select a stage."),
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

  const promoteMutation = useMutation({
    mutationFn: (payload: PromoteLeadRequest) =>
      getSwantaraService().crmLeads.promote(Number(orgId), Number(prospectId), payload),
    onSuccess: () => {
      toast.success("Lead promoted.");
      void queryClient.invalidateQueries({ queryKey: ["crmLeads"] });
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onPromoted();
    },
    onError: () => {
      toast.error("Could not disable the organization.");
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
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Promote lead to opportunity"}</DialogTitle>
          <DialogDescription>
            {"Assign a stage and sales context to qualify this lead."}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="stageId" label={"Stage"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Stage"}>
                    <SelectValue placeholder={"Stage"} />
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
            <FormField name="expectedRevenue" label={"Expected revenue"}>
              {({ field, id }) => <Input {...field} id={id} type="number" step="any" min="0" />}
            </FormField>
            <FormField name="probability" label={"Probability %"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  min="0"
                  max="100"
                  placeholder={"auto from stage"}
                />
              )}
            </FormField>
            <FormField name="priority" label={"Priority"}>
              {({ field, id }) => <Input {...field} id={id} type="number" min="0" />}
            </FormField>
            <FormField name="expectedClose" label={"Expected close"}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
            </FormField>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Promote"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
