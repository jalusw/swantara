"use client";

import { useLocale, useTranslations } from "next-intl";
import { useFormContext } from "react-hook-form";
import type { OnboardingFormSchema } from "../_hooks/use-onboarding-form";
import { countryOptions } from "./company-info-step";

export function ConfirmationStep() {
  const { getValues } = useFormContext<OnboardingFormSchema>();
  const t = useTranslations("Onboarding");
  const tx = t as unknown as (key: string) => string;
  const locale = useLocale();

  const values = getValues();
  const fallbackCountry = countryOptions.find((c) => c.code === values.countryCode);
  const countryDisplay = values.countryCode
    ? (() => {
        const fallbackName = fallbackCountry
          ? `${fallbackCountry.flag} ${fallbackCountry.name} (${fallbackCountry.code})`
          : values.countryCode;
        try {
          const display = new Intl.DisplayNames([locale], { type: "region" });
          const localized = display.of(values.countryCode);
          if (localized && fallbackCountry) {
            return `${fallbackCountry.flag} ${localized} (${fallbackCountry.code})`;
          }
          return localized ?? fallbackName;
        } catch {
          return fallbackName;
        }
      })()
    : "";

  const rows: { id: string; label: string; value: string }[] = [
    { id: "name", label: tx("companyName"), value: values.name },
    { id: "country", label: tx("country"), value: countryDisplay },
  ];

  return (
    <div className="animate-fade-up flex flex-col gap-y-4">
      <div className="rounded-lg border border-divider divide-y divide-divider">
        {rows.map(
          (row) =>
            row.value && (
              <div key={row.id} className="flex items-center justify-between px-4 py-3 text-sm">
                <span className="text-muted-foreground">{row.label}</span>
                <span className=" text-right ml-4 break-all">{row.value}</span>
              </div>
            ),
        )}
      </div>
      {!rows.some((r) => r.value) && (
        <p className="text-sm text-muted-foreground text-center py-4">{tx("reviewEmpty")}</p>
      )}
    </div>
  );
}
