"use client";

import Link from "next/link";
import { useId } from "react";
import { Controller, FormProvider } from "react-hook-form";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Label } from "@/components/label";
import { Password } from "@/components/password";

import { useLoginForm } from "../_hooks/use-login-form";

export default function LoginForm() {
  const { form, handleSubmit, isPending } = useLoginForm({});
  const emailId = useId();
  const passwordId = useId();
  const emailErrorId = `${emailId}-error`;
  const passwordErrorId = `${passwordId}-error`;

  return (
    <FormProvider {...form}>
      <form onSubmit={handleSubmit} noValidate aria-busy={isPending}>
        <div className="flex flex-col gap-y-4">
          <Controller
            name="email"
            control={form.control}
            render={({ field, fieldState }) => {
              const errorId = fieldState.error ? emailErrorId : undefined;
              return (
                <Field className="flex-1">
                  <Label htmlFor={emailId}>{"Email"}</Label>
                  <Input
                    id={emailId}
                    type="email"
                    autoComplete="email"
                    inputMode="email"
                    placeholder={"johndoe@mail.com"}
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
            name="password"
            control={form.control}
            render={({ field, fieldState }) => {
              const errorId = fieldState.error ? passwordErrorId : undefined;
              return (
                <Field className="flex-1">
                  <Label htmlFor={passwordId}>{"Password"}</Label>
                  <Password
                    id={passwordId}
                    autoComplete="current-password"
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
        <div className="mt-2 flex justify-end">
          <Link
            className="inline-flex min-h-12 items-center rounded-md px-2 text-sm text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            href="/forgot-password"
          >
            {"Forgot your password?"}
          </Link>
        </div>
        <div className="mt-6 sm:mt-8">
          <SubmitButton className="w-full" loading={isPending} variant="default" size="lg">
            {"Login"}
          </SubmitButton>
        </div>
      </form>
    </FormProvider>
  );
}
