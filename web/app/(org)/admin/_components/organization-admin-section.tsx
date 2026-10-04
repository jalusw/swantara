"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale, useTranslations } from "next-intl";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { BASE_CURRENCIES, ORGANIZATION_TIMEZONES } from "@/lib/constants/organization";
import { useMeOrganizationsQuery } from "@/lib/hooks/use-me-query";
import { useActiveOrg } from "@/lib/hooks/use-org-context";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";
import { toOrganizationUpdateRequest } from "@/lib/utils/organization";

type ParentOption = { id: number; name: string };

const countries = [
  { code: "US", name: "Amerika Serikat" },
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapura" },
  { code: "DE", name: "Jerman" },
  { code: "NL", name: "Netherlands" },
  { code: "GB", name: "Britania Raya" },
];

function useOrganizationFormSchema() {
  const t = useTranslations("Admin");
  return z.object({
    name: z.string().min(1, t("validation_nameRequired")),
    legalName: z.string(),
    parentId: z.string(),
    baseCurrency: z.string().min(1, t("validation_currencyRequired")),
    countryCode: z.string(),
    taxId: z.string(),
    timezone: z.string().min(1, t("validation_timezoneRequired")),
    taxYearStartMonth: z.string().min(1, t("validation_taxYearRequired")),
  });
}
type OrganizationFormValues = z.infer<ReturnType<typeof useOrganizationFormSchema>>;

export function OrganizationAdminSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Admin");
  const tCommon = useTranslations("Common");
  const locale = useLocale();
  const monthName = (month: number) =>
    new Date(2000, month - 1, 1).toLocaleString(locale, { month: "long" });
  const activeOrg = useActiveOrg();
  const orgListQuery = useMeOrganizationsQuery();
  const { data: orgList } = orgListQuery;

  const organization =
    (orgList?.organizations ?? []).find((org) => String(org.id) === orgId) ?? null;

  const parents: ParentOption[] = (orgList?.organizations ?? [])
    .filter((org) => String(org.id) !== orgId)
    .map((org) => ({ id: org.id, name: org.name }));

  const schema = useOrganizationFormSchema();

  const form = useForm<OrganizationFormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      legalName: "",
      parentId: "",
      baseCurrency: "",
      countryCode: "",
      taxId: "",
      timezone: "",
      taxYearStartMonth: "",
    },
  });

  useEffect(() => {
    if (!organization) return;
    form.reset({
      name: organization.name,
      legalName: organization.legalName ?? "",
      parentId: organization.parentId == null ? "" : String(organization.parentId),
      baseCurrency: organization.baseCurrency ?? "",
      countryCode: organization.countryCode ?? "",
      taxId: organization.taxId ?? "",
      timezone: organization.timezone ?? "",
      taxYearStartMonth: String(organization.taxYearStartMonth ?? ""),
    });
  }, [organization, form]);

  const isActiveOrg = (Number(orgId) || 0) === activeOrg?.id;

  function handleSave(values: OrganizationFormValues) {
    if (!organization) return;
    return getSwantaraService()
      .organizations.update(
        organization.id,
        toOrganizationUpdateRequest(organization, {
          name: values.name,
          legalName: values.legalName,
          parentId: values.parentId ? Number(values.parentId) : null,
          baseCurrency: values.baseCurrency,
          countryCode: values.countryCode || null,
          taxId: values.taxId,
          timezone: values.timezone,
          taxYearStartMonth: Number(values.taxYearStartMonth),
        }),
      )
      .then(() => {
        toast.success(t("profileUpdated"));
        void orgListQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to update organization", error);
        toast.error(t("saveFailed"));
      });
  }

  if (orgListQuery.isPending) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }
  if (orgListQuery.isError) {
    return <p className="text-sm text-destructive">{t("loadFailed")}</p>;
  }
  if (!organization) {
    return <p className="text-sm text-muted-foreground">{t("orgNotFound")}</p>;
  }

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <CardTitle>{t("profileTitle")}</CardTitle>
          <CardDescription>{t("profileDescription")}</CardDescription>
        </div>
        {isActiveOrg ? <Badge variant="secondary">{t("currentOrg")}</Badge> : null}
      </CardHeader>
      <CardContent>
        <Form form={form} onSubmit={handleSave}>
          <div className="grid gap-4 sm:grid-cols-2">
            <FormField name="name" label={t("fieldName")} className="sm:col-span-2">
              {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
            </FormField>
            <FormField name="legalName" label={t("legalName")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("legalName")} />}
            </FormField>
            <FormField name="parentId" label={t("parentOrg")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("parentOrg")}>
                    <SelectValue placeholder={t("noParent")} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="">{t("noParent")}</SelectItem>
                    {parents.map((parent) => (
                      <SelectItem key={parent.id} value={String(parent.id)}>
                        {parent.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="baseCurrency" label={t("baseCurrency")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("baseCurrency")}>
                    <SelectValue placeholder={t("baseCurrency")} />
                  </SelectTrigger>
                  <SelectContent>
                    {BASE_CURRENCIES.map((currency) => (
                      <SelectItem key={currency.code} value={currency.code}>
                        {currency.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="countryCode" label={t("country")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("country")}>
                    <SelectValue placeholder={t("country")} />
                  </SelectTrigger>
                  <SelectContent>
                    {countries.map((country) => (
                      <SelectItem key={country.code} value={country.code}>
                        {country.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="taxId" label={t("taxId")}>
              {({ field, id }) => <Input {...field} id={id} placeholder={t("taxId")} />}
            </FormField>
            <FormField name="timezone" label={t("timezone")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("timezone")}>
                    <SelectValue placeholder={t("timezone")} />
                  </SelectTrigger>
                  <SelectContent>
                    {ORGANIZATION_TIMEZONES.map((timezone) => (
                      <SelectItem key={timezone} value={timezone}>
                        {timezone}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
            <FormField name="taxYearStartMonth" label={t("taxYearStart")}>
              {({ field, id }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger id={id} aria-label={t("taxYearStart")}>
                    <SelectValue placeholder={t("taxYearStart")} />
                  </SelectTrigger>
                  <SelectContent>
                    {Array.from({ length: 12 }, (_, index) => index + 1).map((month) => (
                      <SelectItem key={month} value={String(month)}>
                        {monthName(month)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              )}
            </FormField>
          </div>
          <div className="flex justify-end pt-2">
            <SubmitButton>{t("saveChanges")}</SubmitButton>
          </div>
        </Form>
      </CardContent>
    </Card>
  );
}
