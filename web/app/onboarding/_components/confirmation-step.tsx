"use client";

import { useFormContext } from "react-hook-form";
import type { OnboardingFormSchema } from "../_hooks/use-onboarding-form";
import { countryOptions } from "./company-info-step";

export function ConfirmationStep() {
  const { getValues } = useFormContext<OnboardingFormSchema>();

  const values = getValues();
  const selectedCountry = countryOptions.find((c) => c.code === values.countryCode);
  const countryDisplay = selectedCountry
    ? `${selectedCountry.flag} ${selectedCountry.name} (${selectedCountry.code})`
    : values.countryCode;

  const rows: { id: string; label: string; value: string }[] = [
    { id: "name", label: "Company Name", value: values.name },
    { id: "country", label: "Country", value: countryDisplay },
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
        <p className="text-sm text-muted-foreground text-center py-4">{"Review"}</p>
      )}
    </div>
  );
}
