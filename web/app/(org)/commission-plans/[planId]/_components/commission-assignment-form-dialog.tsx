"use client";

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
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useCommissionAssignmentFormSchema() {
  return z.object({
    salespersonId: z.coerce.number().min(1, "Salesperson is required"),
    dateStart: z.string().min(1, "Start date is required"),
    dateEnd: z.string().optional(),
  });
}
type CommissionAssignmentFormValues = z.infer<ReturnType<typeof useCommissionAssignmentFormSchema>>;

export function CommissionAssignmentFormDialog({
  open,
  onOpenChange,
  orgId,
  planId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  planId: string;
  onSave: () => void;
}) {
  const schema = useCommissionAssignmentFormSchema();

  const form = useForm<CommissionAssignmentFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      salespersonId: 0,
      dateStart: "",
      dateEnd: "",
    },
  });

  function handleSubmit(values: CommissionAssignmentFormValues) {
    void toast.promise(
      getSwantaraService().commissionPlans.assignments.create(Number(orgId), Number(planId), {
        salespersonId: values.salespersonId,
        dateStart: values.dateStart,
        dateEnd: values.dateEnd || null,
      }),
      {
        loading: "Saving…",
        success: () => {
          onSave();
          return "Assignment added";
        },
        error: "Failed to add assignment",
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Add assignment"}</DialogTitle>
          <DialogDescription>{"Assign a salesperson to this commission plan."}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="salespersonId" label={"Salesperson ID"}>
              {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
            </FormField>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="dateStart" label={"Start date"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={"End date"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
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
