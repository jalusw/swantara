"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useTranslations } from "next-intl";
import { useEffect, type ReactNode } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { useMeOrganizationsQuery } from "@/lib/hooks/use-me-query";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";
import { toOrganizationUpdateRequest } from "@/lib/utils/organization";

function useProfileFormSchema() {
  const t = useTranslations("Settings");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    legalName: z.string(),
    taxId: z.string(),
  });
}

type ProfileFormValues = z.infer<ReturnType<typeof useProfileFormSchema>>;

export function ProfileCard({ orgId }: { orgId: string }) {
  const t = useTranslations("Settings");
  const tCommon = useTranslations("Common");
  const { has } = usePermissions();
  const canManage = has("organization.update");
  const orgQuery = useMeOrganizationsQuery();
  const org = orgQuery.data?.organizations.find((candidate) => String(candidate.id) === orgId);

  const schema = useProfileFormSchema();

  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", legalName: "", taxId: "" },
  });

  useEffect(() => {
    if (org) {
      form.reset({
        name: org.name,
        legalName: org.legalName ?? "",
        taxId: org.taxId ?? "",
      });
    }
  }, [org, form]);

  function handleSubmit(values: ProfileFormValues) {
    if (!org) return;
    return getSwantaraService()
      .organizations.update(
        org.id,
        toOrganizationUpdateRequest(org, {
          name: values.name,
          legalName: values.legalName,
          taxId: values.taxId,
        }),
      )
      .then(() => {
        toast.success(t("changesSaved"));
        void orgQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to update organization profile", error);
        toast.error(t("saveFailed"));
      });
  }

  let content: ReactNode;
  if (orgQuery.isPending) {
    content = <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  } else if (orgQuery.isError || !org) {
    content = <p className="text-sm text-destructive">{t("orgNotFound")}</p>;
  } else {
    content = (
      <Form form={form} onSubmit={handleSubmit}>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="name" label={t("orgName")} className="sm:col-span-2">
            {({ field, id }) => (
              <Input {...field} id={id} placeholder={t("orgName")} disabled={!canManage} />
            )}
          </FormField>
          <FormField name="legalName" label={t("legalName")}>
            {({ field, id }) => (
              <Input {...field} id={id} placeholder={t("legalName")} disabled={!canManage} />
            )}
          </FormField>
          <FormField name="taxId" label={t("taxId")}>
            {({ field, id }) => (
              <Input {...field} id={id} placeholder={t("taxId")} disabled={!canManage} />
            )}
          </FormField>
        </div>
        <div className="flex justify-end pt-2">
          <SubmitButton disabled={!canManage}>{t("saveChanges")}</SubmitButton>
        </div>
      </Form>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("profileTitle")}</CardTitle>
        <CardDescription>{t("profileDescription")}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">{content}</CardContent>
    </Card>
  );
}
