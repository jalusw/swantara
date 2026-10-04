"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, type ReactNode } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { BASE_CURRENCIES, ORGANIZATION_TIMEZONES } from "@/lib/constants/organization";
import { useMeOrganizationsQuery } from "@/lib/hooks/use-me-query";
import { usePermissions } from "@/lib/hooks/use-permissions";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";
import { toOrganizationUpdateRequest } from "@/lib/utils/organization";

function useFinancialFormSchema() {
  const t = useTranslations("Settings");
  return z.object({
    baseCurrency: z.string().min(1, t("validationCurrencyRequired")),
    timezone: z.string().min(1, t("validationTimezoneRequired")),
    taxYearStartMonth: z.string().min(1, t("validationTaxYearRequired")),
  });
}

type FinancialFormValues = z.infer<ReturnType<typeof useFinancialFormSchema>>;

function currencyOptionsFor(current?: string): { code: string; name: string }[] {
  const options: { code: string; name: string }[] = BASE_CURRENCIES.map((currency) => ({
    code: currency.code,
    name: currency.name,
  }));
  if (current && !options.some((option) => option.code === current)) {
    options.unshift({ code: current, name: current });
  }
  return options;
}

function timezoneOptionsFor(current?: string): string[] {
  const options: string[] = [...ORGANIZATION_TIMEZONES];
  if (current && !options.includes(current)) {
    options.unshift(current);
  }
  return options;
}

export function FinancialCard({ orgId }: { orgId: string }) {
  const t = useTranslations("Settings");
  const tCommon = useTranslations("Common");
  const locale = useLocale();
  const { has } = usePermissions();
  const canManage = has("organization.update");
  const orgQuery = useMeOrganizationsQuery();
  const org = orgQuery.data?.organizations.find((candidate) => String(candidate.id) === orgId);

  const monthName = (month: number) =>
    new Date(2000, month - 1, 1).toLocaleString(locale, { month: "long" });

  const schema = useFinancialFormSchema();

  const form = useForm<FinancialFormValues>({
    resolver: zodResolver(schema),
    defaultValues: { baseCurrency: "", timezone: "", taxYearStartMonth: "" },
  });

  useEffect(() => {
    if (org) {
      form.reset({
        baseCurrency: org.baseCurrency ?? "",
        timezone: org.timezone ?? "",
        taxYearStartMonth: String(org.taxYearStartMonth ?? ""),
      });
    }
  }, [org, form]);

  function handleSubmit(values: FinancialFormValues) {
    if (!org) return;
    return getSwantaraService()
      .organizations.update(
        org.id,
        toOrganizationUpdateRequest(org, {
          baseCurrency: values.baseCurrency,
          timezone: values.timezone,
          taxYearStartMonth: Number(values.taxYearStartMonth),
        }),
      )
      .then(() => {
        toast.success(t("changesSaved"));
        void orgQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to update organization financial settings", error);
        toast.error(t("saveFailed"));
      });
  }

  let content: ReactNode;
  if (orgQuery.isPending) {
    content = <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  } else if (orgQuery.isError || !org) {
    content = <p className="text-sm text-destructive">{t("orgNotFound")}</p>;
  } else {
    const currencyOptions = currencyOptionsFor(org.baseCurrency);
    const timezoneOptions = timezoneOptionsFor(org.timezone);

    content = (
      <Form form={form} onSubmit={handleSubmit}>
        <div className="grid gap-4 sm:grid-cols-2">
          <FormField name="baseCurrency" label={t("defaultCurrency")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange} disabled={!canManage}>
                <SelectTrigger id={id} aria-label={t("defaultCurrency")}>
                  <SelectValue placeholder={t("defaultCurrency")} />
                </SelectTrigger>
                <SelectContent>
                  {currencyOptions.map((currency) => (
                    <SelectItem key={currency.code} value={currency.code}>
                      {currency.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}
          </FormField>
          <FormField name="taxYearStartMonth" label={t("taxYear")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange} disabled={!canManage}>
                <SelectTrigger id={id} aria-label={t("taxYear")}>
                  <SelectValue placeholder={t("taxYear")} />
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
          <FormField name="timezone" label={t("timezone")}>
            {({ field, id }) => (
              <Select value={field.value} onValueChange={field.onChange} disabled={!canManage}>
                <SelectTrigger id={id} aria-label={t("timezone")}>
                  <SelectValue placeholder={t("timezone")} />
                </SelectTrigger>
                <SelectContent>
                  {timezoneOptions.map((timezone) => (
                    <SelectItem key={timezone} value={timezone}>
                      {timezone}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
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
        <CardTitle>{t("financialTitle")}</CardTitle>
        <CardDescription>{t("financialDescription")}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">{content}</CardContent>
    </Card>
  );
}
