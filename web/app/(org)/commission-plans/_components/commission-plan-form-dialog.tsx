"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { getSwantaraService } from "@/lib/services/swantara";

function useCommissionPlanFormSchema() {
  return z.object({
    name: z.string().min(1, "Name is required"),
    basis: z.string().min(1, "Basis is required"),
  });
}
type CommissionPlanFormValues = z.infer<ReturnType<typeof useCommissionPlanFormSchema>>;

export function CommissionPlanFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: (id: string) => void;
}) {
  const schema = useCommissionPlanFormSchema();

  const form = useForm<CommissionPlanFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      basis: "",
    },
  });

  function handleSubmit(values: CommissionPlanFormValues) {
    void toast.promise(
      getSwantaraService().commissionPlans.create(Number(orgId), {
        name: values.name,
        basis: values.basis as "revenue" | "margin" | "collected",
      }),
      {
        loading: "Saving…",
        success: (result) => {
          onSave(String(result.commissionPlan.id));
          return "Plan created";
        },
        error: "Failed to create plan",
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{"Create commission plan"}</DialogTitle>
          <DialogDescription>{"Assign a salesperson to this commission plan."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="basis" label={"Basis"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Basis"}>
                      <SelectValue placeholder={"Select Basis"} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="revenue">{"Revenue"}</SelectItem>
                      <SelectItem value="margin">{"Margin"}</SelectItem>
                      <SelectItem value="collected">{"Collected"}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
            </div>
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
