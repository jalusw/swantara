"use client";

import { useTranslations } from "next-intl";
import { useId } from "react";
import { Controller, useFormContext } from "react-hook-form";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { Input } from "@/components/input";
import { Label } from "@/components/label";

import { type RegisterFormSchema, useRegisterFormStore } from "../_hooks/use-register-form";

export function NameStep() {
  const { control } = useFormContext<RegisterFormSchema>();
  const step = useRegisterFormStore((s) => s.step);
  const t = useTranslations("Auth");
  const tx = t as unknown as (key: string) => string;
  const firstNameId = useId();
  const lastNameId = useId();

  if (step !== "name") {
    return null;
  }

  return (
    <div className="animate-fade-up flex flex-col gap-y-4">
      <Controller
        name="firstName"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${firstNameId}-error` : undefined;
          return (
            <Field>
              <Label htmlFor={firstNameId}>{tx("firstName")}</Label>
              <Input
                id={firstNameId}
                type="text"
                autoComplete="given-name"
                autoFocus
                placeholder={"John"}
                aria-invalid={Boolean(fieldState.error)}
                aria-describedby={errorId}
                {...field}
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
        name="lastName"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${lastNameId}-error` : undefined;
          return (
            <Field>
              <Label htmlFor={lastNameId}>{tx("lastName")}</Label>
              <Input
                id={lastNameId}
                type="text"
                autoComplete="family-name"
                placeholder={"Doe"}
                aria-invalid={Boolean(fieldState.error)}
                aria-describedby={errorId}
                {...field}
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
    </div>
  );
}
