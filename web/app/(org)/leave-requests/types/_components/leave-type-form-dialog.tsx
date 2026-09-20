"use client";

import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Switch } from "@/components/switch";
import type { LeaveType } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useLeaveTypeFormSchema() {
  return z.object({
    name: z.string().min(1, "Enter a name."),
    paid: z.boolean(),
    allocationDays: z.coerce.number().nullable(),
  });
}

export function LeaveTypeFormDialog({
  open,
  onOpenChange,
  orgId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  initial?: LeaveType | null;
  onSave: () => void;
}) {
  const isEdit = Boolean(initial);

  const schema = useLeaveTypeFormSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: initial
      ? {
          name: initial.name,
          paid: initial.paid,
          allocationDays: initial.allocationDays,
        }
      : {
          name: "",
          paid: true,
          allocationDays: null,
        },
  });

  function handleSubmit(values: Values) {
    const request = {
      name: values.name,
      paid: values.paid,
      allocationDays: values.allocationDays,
      organizationId: Number(orgId),
    };

    if (isEdit && initial) {
      void getSwantaraService()
        .leaveTypes.update(Number(orgId), initial.id, request)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    } else {
      void getSwantaraService()
        .leaveTypes.create(Number(orgId), request)
        .then(() => onSave())
        .catch(() => toast.error("Something went wrong."));
    }
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={isEdit ? "Edit leave type" : "New leave type"}
      description={"Define a leave type with optional allocation."}
      form={form}
      onSubmit={handleSubmit}
      className="sm:max-w-lg"
    >
      <div className="flex flex-col gap-4">
        <FormField name="name" label={"Name"}>
          {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
        </FormField>
        <FormField name="paid" label={"Paid"}>
          {({ field }) => (
            <Switch
              checked={field.value}
              onCheckedChange={(checked) => field.onChange(Boolean(checked))}
              aria-label={"Paid"}
            />
          )}
        </FormField>
        <FormField name="allocationDays" label={"Allocation days"}>
          {({ field, id }) => (
            <Input
              {...field}
              id={id}
              type="number"
              min="0"
              value={field.value ?? ""}
              onChange={(event) => {
                const value = event.target.value;
                field.onChange(value === "" ? null : Number(value));
              }}
            />
          )}
        </FormField>
      </div>
    </EntityFormDialog>
  );
}
