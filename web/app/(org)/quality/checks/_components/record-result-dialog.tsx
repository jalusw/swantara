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
import type { QualityCheck } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { zodResolver } from "@/lib/utils/zod-resolver";

const recordResultSchema = z.object({
  pass: z.enum(["pass", "fail"]),
  measuredValue: z.coerce.number().nullable(),
});
type RecordResultValues = z.infer<typeof recordResultSchema>;

export function RecordResultDialog({
  open,
  onOpenChange,
  orgId,
  check,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  check: QualityCheck;
  onSave: () => void;
}) {
  const form = useForm<RecordResultValues>({
    resolver: zodResolver(recordResultSchema),
    defaultValues: {
      pass: "pass",
      measuredValue: null,
    },
  });

  function handleSubmit(values: RecordResultValues) {
    void getSwantaraService()
      .qualityChecks.result(Number(orgId), check.id, {
        pass: values.pass === "pass",
        measuredValue: values.measuredValue,
        checkedBy: null,
      })
      .then(() => onSave())
      .catch(() => toast.error("Could not disable the organization."));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{"Record Result"}</DialogTitle>
          <DialogDescription>{`Record result for check #${check.id}`}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="pass" label={"Result"}>
              {({ field, id }) => (
                <div className="flex gap-2" id={id}>
                  <Button
                    type="button"
                    variant={field.value === "pass" ? "default" : "outline"}
                    onClick={() => field.onChange("pass")}
                  >
                    {"Pass"}
                  </Button>
                  <Button
                    type="button"
                    variant={field.value === "fail" ? "destructive" : "outline"}
                    onClick={() => field.onChange("fail")}
                  >
                    {"Fail"}
                  </Button>
                </div>
              )}
            </FormField>
            <FormField name="measuredValue" label={"Measured value"}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="number"
                  step="0.01"
                  inputMode="decimal"
                  value={field.value ?? ""}
                  onChange={(e) => field.onChange(e.target.value ? Number(e.target.value) : null)}
                  placeholder={"Enter value"}
                />
              )}
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
