"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Textarea } from "@/components/textarea";
import { getSwantaraService } from "@/lib/services/swantara";

function useLostReasonFormSchema() {
  const t = useTranslations("Crm");
  return z.object({
    lostReason: z.string().min(1, t("validationLostReasonRequired")),
  });
}
type LostReasonValues = z.infer<ReturnType<typeof useLostReasonFormSchema>>;

export function LostReasonDialog({
  open,
  onOpenChange,
  orgId,
  opportunityId,
  onDone,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  opportunityId: string;
  onDone: () => void;
}) {
  const t = useTranslations("Crm");
  const queryClient = useQueryClient();
  const schema = useLostReasonFormSchema();
  const form = useForm<LostReasonValues>({
    resolver: zodResolver(schema),
    defaultValues: { lostReason: "" },
  });

  useEffect(() => {
    if (open) {
      form.reset({ lostReason: "" });
    }
  }, [open, form]);

  const loseMutation = useMutation({
    mutationFn: (lostReason: string) =>
      getSwantaraService().crmOpportunities.lose(Number(orgId), Number(opportunityId), {
        lostReason,
      }),
    onSuccess: () => {
      toast.success(t("markedLost"));
      void queryClient.invalidateQueries({ queryKey: ["crmOpportunities"] });
      onDone();
    },
    onError: () => {
      toast.error(t("saveFailed"));
    },
  });

  function handleSubmit(values: LostReasonValues) {
    loseMutation.mutate(values.lostReason);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={t("markLost")}
      description={t("lostReasonDescription")}
      badge={<Badge variant="destructive">{t("lostBadge")}</Badge>}
      form={form}
      onSubmit={handleSubmit}
      isPending={loseMutation.isPending}
      submitLabel={t("markLost")}
      submitVariant="destructive"
      footerHint={t("lostHint")}
      className="sm:max-w-lg"
    >
      <FormField name="lostReason" label={t("fieldLostReason")} description={t("lostReasonHint")}>
        {({ field, id }) => (
          <Textarea
            {...field}
            id={id}
            autoFocus
            rows={4}
            placeholder={t("lostReasonPlaceholder")}
          />
        )}
      </FormField>
    </EntityFormDialog>
  );
}
