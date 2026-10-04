"use client";

import { useTranslations } from "next-intl";
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
  const t = useTranslations("Commissions");
  return z.object({
    salespersonId: z.coerce.number().min(1, t("validationSalespersonRequired")),
    dateStart: z.string().min(1, t("validationStartDateRequired")),
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
  const t = useTranslations("Commissions");
  const tCommon = useTranslations("Common");
  const schema = useCommissionAssignmentFormSchema();

  const form = useForm<CommissionAssignmentFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      salespersonId: 0,
      dateStart: "",
      dateEnd: "",
    },
  });

  async function handleSubmit(values: CommissionAssignmentFormValues) {
    const request = getSwantaraService().commissionPlans.assignments.create(
      Number(orgId),
      Number(planId),
      {
        salespersonId: values.salespersonId,
        dateStart: values.dateStart,
        dateEnd: values.dateEnd || null,
      },
    );
    toast.promise(request, {
      loading: t("saving"),
      success: () => {
        onSave();
        return t("assignmentCreated");
      },
      error: t("createAssignmentFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("addAssignment")}</DialogTitle>
          <DialogDescription>{t("newAssignmentDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="salespersonId" label={t("fieldSalespersonId")}>
              {({ field, id }) => <Input {...field} id={id} type="number" min={1} />}
            </FormField>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="dateStart" label={t("fieldDateStart")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={t("fieldDateEnd")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
