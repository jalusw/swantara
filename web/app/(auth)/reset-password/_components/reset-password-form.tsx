"use client";

import { AlertCircle, CheckCircle2 } from "lucide-react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import { useEffect, useId } from "react";
import { Controller } from "react-hook-form";
import { Button } from "@/components/button";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { SubmitButton } from "@/components/form";
import { Label } from "@/components/label";
import { Password } from "@/components/password";

import { useResetPasswordForm } from "../_hooks/use-reset-password-form";

export default function ResetPasswordForm({ token }: { token?: string }) {
  const { form, handleSubmit, isPending, isSuccess } = useResetPasswordForm({
    token,
  });
  const t = useTranslations("Auth");
  const tx = t as unknown as (key: string) => string;
  const passwordId = useId();
  const confirmId = useId();

  useEffect(() => {
    if (token && window.location.search.includes("token=")) {
      window.history.replaceState(null, "", window.location.pathname);
    }
  }, [token]);

  if (!token) {
    return (
      <div className="animate-fade-up flex flex-col items-center gap-4 py-4 text-center">
        <AlertCircle className="size-10 text-destructive" aria-hidden="true" />
        <h2 className="text-xl">{tx("resetLinkInvalidTitle")}</h2>
        <p className="max-w-sm text-muted-foreground">{tx("resetLinkInvalidDescription")}</p>
        <Button asChild variant="default" size="lg" className="w-full">
          <Link href="/forgot-password">{tx("requestNewLink")}</Link>
        </Button>
        <Button asChild variant="outline" size="lg" className="w-full">
          <Link href="/login">{t("backToLogin")}</Link>
        </Button>
      </div>
    );
  }

  if (isSuccess) {
    return (
      <div className="animate-fade-up flex flex-col items-center gap-4 py-4 text-center">
        <CheckCircle2 className="size-10 text-primary" aria-hidden="true" />
        <h2 className="text-xl">{tx("passwordUpdatedTitle")}</h2>
        <p className="max-w-sm text-muted-foreground">{tx("passwordUpdatedDescription")}</p>
        <Button asChild variant="default" size="lg" className="w-full">
          <Link href="/login">{tx("goToLogin")}</Link>
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} noValidate aria-busy={isPending}>
      <div className="flex flex-col gap-y-4">
        <Controller
          name="password"
          control={form.control}
          render={({ field, fieldState }) => {
            const errorId = fieldState.error ? `${passwordId}-error` : undefined;
            return (
              <Field>
                <Label htmlFor={passwordId}>{tx("newPassword")}</Label>
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
          control={form.control}
          render={({ field, fieldState }) => {
            const errorId = fieldState.error ? `${confirmId}-error` : undefined;
            return (
              <Field>
                <Label htmlFor={confirmId}>{tx("confirmNewPassword")}</Label>
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
      <div className="mt-6 flex flex-col gap-3 sm:mt-8">
        <SubmitButton className="w-full" loading={isPending} size="lg">
          {t("resetButton")}
        </SubmitButton>
        <Button variant="outline" size="lg" asChild className="w-full">
          <Link href="/login">{t("backToLogin")}</Link>
        </Button>
      </div>
    </form>
  );
}
