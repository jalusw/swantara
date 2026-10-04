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
import type { ProjectMilestone } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { getLocalDateString } from "@/lib/utils";
import { zodResolver } from "@/lib/utils/zod-resolver";

function useMilestoneFormSchema() {
  const t = useTranslations("Projects");
  return z.object({
    name: z.string().min(1, t("validation_milestoneNameRequired")),
    deadline: z.string().nullable(),
    saleLineId: z.coerce.number().nullable(),
  });
}
type MilestoneFormValues = z.infer<ReturnType<typeof useMilestoneFormSchema>>;

export function MilestoneFormDialog({
  open,
  onOpenChange,
  orgId,
  projectId,
  initial,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  projectId: number;
  initial?: ProjectMilestone | null;
  onSave: () => void;
}) {
  const t = useTranslations("Projects");
  const tCommon = useTranslations("Common");
  const isEdit = Boolean(initial);

  const schema = useMilestoneFormSchema();

  const form = useForm<MilestoneFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: initial?.name ?? "",
      deadline: initial?.deadline ? getLocalDateString(new Date(initial.deadline)) : null,
      saleLineId: initial?.saleLineId ?? null,
    },
  });

  function handleSubmit(values: MilestoneFormValues) {
    const payload = {
      name: values.name,
      deadline: values.deadline ? new Date(values.deadline) : null,
      saleLineId: values.saleLineId || null,
    };

    return getSwantaraService()
      .projects.milestones.create(Number(orgId), projectId, payload)
      .then(() => onSave())
      .catch(() => void toast.error(t("saveFailed")));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{isEdit ? t("editMilestone") : t("newMilestone")}</DialogTitle>
          <DialogDescription>{t("milestoneFormDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <FormField name="name" label={t("milestoneName")}>
              {({ field, id }) => <Input {...field} id={id} />}
            </FormField>
            <FormField name="deadline" label={t("deadline")}>
              {({ field, id }) => (
                <Input
                  {...field}
                  id={id}
                  type="date"
                  value={field.value ?? ""}
                  onChange={(e) => field.onChange(e.target.value || null)}
                />
              )}
            </FormField>
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
