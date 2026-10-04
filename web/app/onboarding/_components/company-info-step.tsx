"use client";

import { useLocale, useTranslations } from "next-intl";
import { useId, useMemo } from "react";
import { Controller, useFormContext } from "react-hook-form";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/combobox";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { Label } from "@/components/label";
import {
  QuestionnaireDescription,
  QuestionnaireInput,
  QuestionnaireItem,
  QuestionnaireTitle,
} from "@/components/questionnaire";

import type { OnboardingFormSchema } from "../_hooks/use-onboarding-form";

function countryCodeToFlag(code: string): string {
  return code
    .toUpperCase()
    .replace(/./g, (char) => String.fromCodePoint(char.charCodeAt(0) + 127397));
}

export const countryOptions = [
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapura" },
  { code: "US", name: "Amerika Serikat" },
  { code: "GB", name: "Britania Raya" },
  { code: "DE", name: "Jerman" },
  { code: "MY", name: "Malaysia" },
  { code: "TH", name: "Thailand" },
  { code: "PH", name: "Filipina" },
  { code: "VN", name: "Vietnam" },
  { code: "AU", name: "Australia" },
  { code: "JP", name: "Jepang" },
  { code: "KR", name: "Korea Selatan" },
].map((c) => ({ ...c, flag: countryCodeToFlag(c.code) }));

function useLocalizedCountryOptions() {
  const locale = useLocale();
  return useMemo(
    () =>
      countryOptions.map((country) => {
        try {
          const name = new Intl.DisplayNames([locale], { type: "region" }).of(country.code);
          return { ...country, name: name ?? country.name };
        } catch {
          return country;
        }
      }),
    [locale],
  );
}

export function CompanyInfoStep() {
  const { control } = useFormContext<OnboardingFormSchema>();
  const t = useTranslations("Onboarding");
  const tx = t as unknown as (key: string) => string;
  const localizedCountries = useLocalizedCountryOptions();
  const nameId = useId();
  const countryId = useId();

  return (
    <QuestionnaireItem name="companyInfo" required>
      <QuestionnaireTitle>{tx("companyInfoTitle")}</QuestionnaireTitle>
      <QuestionnaireDescription>{tx("companyInfoDescription")}</QuestionnaireDescription>
      <Controller
        name="name"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${nameId}-error` : undefined;
          const { value, onChange, onBlur, ref } = field;
          return (
            <Field>
              <Label htmlFor={nameId}>{tx("companyName")}</Label>
              <QuestionnaireInput
                id={nameId}
                type="text"
                autoComplete="organization"
                autoFocus
                placeholder={"Acme Inc."}
                aria-invalid={Boolean(fieldState.error)}
                aria-describedby={errorId}
                ref={ref}
                value={value}
                onChange={(event) => onChange(event.target.value)}
                onBlur={onBlur}
              />
              <FieldFeedback
                id={errorId}
                visible={Boolean(fieldState.error)}
                intent="danger"
                role={fieldState.error ? "alert" : undefined}
              >
                {fieldState.error?.message}
              </FieldFeedback>
            </Field>
          );
        }}
      />
      <Controller
        name="countryCode"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${countryId}-error` : undefined;
          return (
            <Field>
              <Label htmlFor={countryId}>{tx("country")}</Label>
              <Combobox
                items={localizedCountries.map((country) => country.code)}
                itemToStringLabel={(value) =>
                  localizedCountries.find((country) => country.code === value)?.name ?? ""
                }
                value={field.value}
                onValueChange={(value) => field.onChange(value ?? "")}
              >
                <ComboboxInput
                  id={countryId}
                  placeholder={tx("selectCountry")}
                  aria-invalid={Boolean(fieldState.error)}
                  aria-describedby={errorId}
                />
                <ComboboxContent>
                  <ComboboxList>
                    {(code: string) => {
                      const country = localizedCountries.find((c) => c.code === code);
                      if (!country) return null;
                      return (
                        <ComboboxItem key={country.code} value={country.code}>
                          <span className="flex items-center gap-2">
                            <span aria-hidden="true" className="text-base leading-none">
                              {country.flag}
                            </span>
                            <span>{country.name}</span>
                            <span className="text-muted-foreground text-xs">{country.code}</span>
                          </span>
                        </ComboboxItem>
                      );
                    }}
                  </ComboboxList>
                  <ComboboxEmpty>{tx("noResults")}</ComboboxEmpty>
                </ComboboxContent>
              </Combobox>
              <QuestionnaireInput
                value={field.value}
                readOnly
                tabIndex={-1}
                aria-hidden="true"
                className="hidden"
              />
              <FieldFeedback
                id={errorId}
                visible={Boolean(fieldState.error)}
                intent="danger"
                role={fieldState.error ? "alert" : undefined}
              >
                {fieldState.error?.message}
              </FieldFeedback>
            </Field>
          );
        }}
      />
    </QuestionnaireItem>
  );
}
