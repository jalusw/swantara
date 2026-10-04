"use client";

import { useTranslations } from "next-intl";
import { useId } from "react";
import { Controller, useFormContext } from "react-hook-form";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { Label } from "@/components/label";
import { Password } from "@/components/password";

import { type RegisterFormSchema, useRegisterFormStore } from "../_hooks/use-register-form";

export function PasswordStep() {
  const { control } = useFormContext<RegisterFormSchema>();
  const step = useRegisterFormStore((s) => s.step);
  const t = useTranslations("Auth");
  const tx = t as unknown as (key: string) => string;
  const passwordId = useId();
  const confirmId = useId();

  if (step !== "password") {
    return null;
  }

  return (
    <div className="animate-fade-up flex flex-col gap-y-4">
      <Controller
        name="password"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${passwordId}-error` : undefined;
          return (
            <Field>
              <Label htmlFor={passwordId}>{t("password")}</Label>
              <Password
                id={passwordId}
                autoComplete="new-password"
                autoFocus
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
              <p className="text-sm text-muted-foreground">{tx("passwordMinLengthHint")}</p>
            </Field>
          );
        }}
      />
      <Controller
        name="passwordConfirmation"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${confirmId}-error` : undefined;
          return (
            <Field>
              <Label htmlFor={confirmId}>{t("confirmPassword")}</Label>
              <Password
                id={confirmId}
                autoComplete="new-password"
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
