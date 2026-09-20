"use client";

import { useId } from "react";
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
  { code: "SG", name: "Singapore" },
  { code: "US", name: "United States" },
  { code: "GB", name: "United Kingdom" },
  { code: "DE", name: "Germany" },
  { code: "MY", name: "Malaysia" },
  { code: "TH", name: "Thailand" },
  { code: "PH", name: "Philippines" },
  { code: "VN", name: "Vietnam" },
  { code: "AU", name: "Australia" },
  { code: "JP", name: "Japan" },
  { code: "KR", name: "South Korea" },
].map((c) => ({ ...c, flag: countryCodeToFlag(c.code) }));

export function CompanyInfoStep() {
  const { control } = useFormContext<OnboardingFormSchema>();
  const nameId = useId();
  const countryId = useId();

  return (
    <QuestionnaireItem name="companyInfo" required>
      <QuestionnaireTitle>{"Company Info"}</QuestionnaireTitle>
      <QuestionnaireDescription>{"Tell us about your company."}</QuestionnaireDescription>
      <Controller
        name="name"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${nameId}-error` : undefined;
          const { value, onChange, onBlur, ref } = field;
          return (
            <Field>
              <Label htmlFor={nameId}>{"Company Name"}</Label>
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
              <Label htmlFor={countryId}>{"Country"}</Label>
              <Combobox
                items={countryOptions.map((country) => country.code)}
                itemToStringLabel={(value) =>
                  countryOptions.find((country) => country.code === value)?.name ?? ""
                }
                value={field.value}
                onValueChange={(value) => field.onChange(value ?? "")}
              >
                <ComboboxInput
                  id={countryId}
                  placeholder={"Select country"}
                  aria-invalid={Boolean(fieldState.error)}
                  aria-describedby={errorId}
                />
                <ComboboxContent>
                  <ComboboxList>
                    {(code: string) => {
                      const country = countryOptions.find((c) => c.code === code);
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
                  <ComboboxEmpty>{"No results found"}</ComboboxEmpty>
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
