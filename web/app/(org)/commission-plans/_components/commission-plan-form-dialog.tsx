"use client";

import { zodResolver } from "@hookform/resolvers/zod";
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { getSwantaraService } from "@/lib/services/swantara";

function useCommissionPlanFormSchema() {
  const t = useTranslations("Commissions");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    basis: z.string().min(1, t("validationBasisRequired")),
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
  const t = useTranslations("Commissions");
  const tCommon = useTranslations("Common");
  const schema = useCommissionPlanFormSchema();

  function basisLabel(value: string): string {
    return (t as unknown as (k: string) => string)(`basis_${value}`);
  }

  const form = useForm<CommissionPlanFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      basis: "",
    },
  });

  async function handleSubmit(values: CommissionPlanFormValues) {
    const request = getSwantaraService().commissionPlans.create(Number(orgId), {
      name: values.name,
      basis: values.basis as "revenue" | "margin" | "collected",
    });
    toast.promise(request, {
      loading: t("saving"),
      success: (result) => {
        onSave(String(result.commissionPlan.id));
        return t("planCreated");
      },
      error: t("createPlanFailed"),
    });
    await request.catch(() => {});
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("newPlan")}</DialogTitle>
          <DialogDescription>{t("newPlanDescription")}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("name")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("name")} />}
              </FormField>
              <FormField name="basis" label={t("basis")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("basis")}>
                      <SelectValue placeholder={t("selectBasis")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="revenue">{basisLabel("revenue")}</SelectItem>
                      <SelectItem value="margin">{basisLabel("margin")}</SelectItem>
                      <SelectItem value="collected">{basisLabel("collected")}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
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
