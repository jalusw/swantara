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
import type { SubscriptionPlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const intervals = ["daily", "weekly", "monthly", "quarterly", "yearly"] as const;

function useSubscriptionPlanFormSchema() {
  return z.object({
    name: z.string().min(1, "Name is required"),
    recurringInterval: z.enum(intervals),
    recurringCount: z.string().min(1, "Multiplier is required"),
  });
}
type SubscriptionPlanFormValues = z.infer<ReturnType<typeof useSubscriptionPlanFormSchema>>;

export function SubscriptionPlanFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: SubscriptionPlan | null;
  onSave: (id: string) => void;
}) {
  const queryClient = useQueryClient();
  const isEdit = Boolean(initial);

  const schema = useSubscriptionPlanFormSchema();

  const form = useForm<SubscriptionPlanFormValues>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          recurringInterval:
            initial.recurringInterval as SubscriptionPlanFormValues["recurringInterval"],
          recurringCount: String(initial.recurringCount),
        }
      : {
          name: "",
          recurringInterval: "monthly",
          recurringCount: "1",
        },
  });

  const createMutation = useMutation({
    mutationFn: (body: { name: string; recurringInterval: string; recurringCount: number }) =>
      getSwantaraService().subscriptionPlans.create(Number(orgId), body),
    onSuccess: (result) => {
      void queryClient.invalidateQueries({ queryKey: ["subscriptionPlans", Number(orgId)] });
      onSave(String(result.plan.id));
      toast.success("Plan created");
    },
    onError: () => {
      toast.error("Failed to save plan");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      planId,
      body,
    }: {
      planId: number;
      body: { name: string; recurringInterval: string; recurringCount: number };
    }) => getSwantaraService().subscriptionPlans.update(Number(orgId), planId, body),
    onSuccess: (_data, variables) => {
      void queryClient.invalidateQueries({ queryKey: ["subscriptionPlans", Number(orgId)] });
      onSave(String(variables.planId));
      toast.success("Plan updated");
    },
    onError: () => {
      toast.error("Failed to save plan");
    },
  });

  function handleSubmit(values: SubscriptionPlanFormValues) {
    const body = {
      name: values.name,
      recurringInterval: values.recurringInterval,
      recurringCount: Number(values.recurringCount),
    };

    if (isEdit && initial) {
      updateMutation.mutate({ planId: initial.id, body });
    } else {
      createMutation.mutate(body);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? "Edit subscription plan" : "Create subscription plan"}
          </DialogTitle>
          <DialogDescription>{"Define the recurring interval and multiplier."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="name" label={"Name"}>
              {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
            </FormField>
            <FormField name="recurringInterval" label={"Recurring interval"}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={"Recurring interval"}>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {intervals.map((iv) => (
                      <SelectItem key={iv} value={iv}>
                        {String(iv)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="recurringCount" label={"Multiplier"}>
              {({ field, id }) => <Input {...field} id={id} type="number" min="1" />}
            </FormField>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {"Cancel"}
            </Button>
            <SubmitButton>{"Save"}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
