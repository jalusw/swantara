"use client";

import { useTranslations } from "next-intl";
import { useEffect, useId, useState } from "react";
import { Controller, useFormContext } from "react-hook-form";
import { useShallow } from "zustand/react/shallow";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { Input } from "@/components/input";
import { Label } from "@/components/label";
import { axiosInstance } from "@/lib/client/axios";
import { useDebouncedValue } from "@/lib/hooks/use-debounced-value";

import { type RegisterFormSchema, useRegisterFormStore } from "../_hooks/use-register-form";

function isEmailFormatValid(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function EmailStep() {
  const { control, watch } = useFormContext<RegisterFormSchema>();
  const t = useTranslations("Auth");
  const tx = t as unknown as (key: string) => string;
  const { step } = useRegisterFormStore(
    useShallow((s) => ({
      step: s.step,
    })),
  );
  const emailId = useId();

  const emailValue = watch("email");
  const debouncedEmail = useDebouncedValue(emailValue?.trim() ?? "", 500);
  const [emailAvailable, setEmailAvailable] = useState<boolean | null>(null);

  useEffect(() => {
    if (!debouncedEmail || !isEmailFormatValid(debouncedEmail)) {
      setEmailAvailable(null);
      return;
    }

    let cancelled = false;
    setEmailAvailable(null);

    axiosInstance
      .post<{ data: { available: boolean } }>("/auth/email/check", { email: debouncedEmail })
      .then((res) => {
        if (!cancelled) setEmailAvailable(res.data.data.available);
      })
      .catch(() => {
        if (!cancelled) setEmailAvailable(null);
      });

    return () => {
      cancelled = true;
    };
  }, [debouncedEmail]);

  if (step !== "email") {
    return null;
  }

  const unavailableMessage = emailAvailable === false ? tx("emailTaken") : undefined;

  return (
    <div className="animate-fade-up flex flex-col gap-y-4">
      <Controller
        name="email"
        control={control}
        render={({ field, fieldState }) => {
          const errorId = fieldState.error ? `${emailId}-error` : undefined;
          const feedbackId = errorId ?? (unavailableMessage ? `${emailId}-unavailable` : undefined);
          return (
            <Field>
              <Label htmlFor={emailId}>{t("email")}</Label>
              <Input
                id={emailId}
                type="email"
                autoComplete="email"
                inputMode="email"
                autoFocus
                placeholder={"johndoe@mail.com"}
                aria-invalid={Boolean(fieldState.error) || emailAvailable === false}
                aria-describedby={feedbackId}
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
              <FieldFeedback
                visible={!fieldState.error && Boolean(unavailableMessage)}
                intent="danger"
                role="alert"
              >
                {unavailableMessage}
              </FieldFeedback>
            </Field>
          );
        }}
      />
    </div>
  );
}
