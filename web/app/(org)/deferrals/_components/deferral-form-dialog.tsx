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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useDeferralSchema() {
  return z.object({
    type: z.string().min(1, "Type is required"),
    sourceType: z.string().min(1, "Source type is required"),
    sourceId: z.coerce.number().min(1, "Source ID is required"),
    totalAmount: z.coerce.number().positive("Total amount must be positive"),
    method: z.string().min(1, "Method is required"),
    periods: z.coerce.number().min(1, "Periods is required"),
    dateStart: z.string().min(1, "Start date is required"),
  });
}
type DeferralValues = z.infer<ReturnType<typeof useDeferralSchema>>;

export function DeferralFormDialog({
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
  const schema = useDeferralSchema();

  const form = useForm<DeferralValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      type: "",
      sourceType: "",
      sourceId: 0,
      totalAmount: 0,
      method: "",
      periods: 1,
      dateStart: "",
    },
  });

  function handleSubmit(values: DeferralValues) {
    void toast.promise(
      getSwantaraService().deferrals.create(Number(orgId), {
        type: values.type as "deferred_revenue" | "deferred_expense" | "prepaid",
        sourceType: values.sourceType,
        sourceId: values.sourceId,
        contactId: null,
        itemId: null,
        totalAmount: values.totalAmount,
        balanceSheetAccountId: 1,
        plAccountId: 1,
        dimensionId: null,
        method: values.method as "linear" | "manual" | "milestone",
        dateStart: new Date(values.dateStart),
        dateEnd: null,
        periods: values.periods,
        lines: [],
      }),
      {
        loading: "Saving…",
        success: (result) => {
          onSave(String(result.schedule.id));
          return "Deferral schedule created";
        },
        error: "Failed to create deferral schedule",
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{"Create deferral"}</DialogTitle>
          <DialogDescription>
            {"Create a deferral schedule for deferred revenue, expense, or prepaid items."}
          </DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="type" label={"Type"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Type"}>
                      <SelectValue placeholder={"Select type"} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="deferred_revenue">{"Deferred revenue"}</SelectItem>
                      <SelectItem value="deferred_expense">{"Deferred expense"}</SelectItem>
                      <SelectItem value="prepaid">{"Prepaid"}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="method" label={"Method"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Method"}>
                      <SelectValue placeholder={"Select method"} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="linear">{"Linear"}</SelectItem>
                      <SelectItem value="manual">{"Manual"}</SelectItem>
                      <SelectItem value="milestone">{"Milestone"}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="sourceType" label={"Source type"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Source type"} />}
              </FormField>
              <FormField name="sourceId" label={"Source ID"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
              </FormField>
            </div>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="totalAmount" label={"Total amount"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={0} step="0.01" />}
              </FormField>
              <FormField name="periods" label={"Periods"}>
                {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
              </FormField>
            </div>
            <FormField name="dateStart" label={"Start date"}>
              {({ field, id }) => <Input {...field} id={id} type="date" />}
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
