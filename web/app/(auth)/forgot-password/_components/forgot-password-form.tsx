"use client";

import { MailCheck } from "lucide-react";
import Link from "next/link";
import { useId } from "react";
import { Controller } from "react-hook-form";
import { Button } from "@/components/button";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Label } from "@/components/label";
import { detectMethod, useForgotPasswordForm } from "../_hooks/use-forgot-password-form";

export default function ForgotPasswordForm() {
  const { form, handleSubmit, isPending, isSubmitted } = useForgotPasswordForm();
  const identifierId = useId();

  if (isSubmitted) {
    const identifier = form.getValues("identifier");
    const method = detectMethod(identifier);
    const methodLabel = method === "email" ? "Email" : "Phone";
    return (
      <div className="animate-fade-up flex flex-col items-center gap-4 text-center">
        <MailCheck className="size-10 text-primary" aria-hidden="true" />
        <h2 className="font-heading text-2xl">{"Reset link sent"}</h2>
        <p className="text-muted-foreground">{`If an account exists for that ${methodLabel}, you'll receive instructions to reset your password shortly.`}</p>
        <div className="flex w-full flex-col gap-3 pt-2">
          <Button
            type="button"
            variant="outline"
            size="lg"
            className="w-full"
            onClick={() => form.reset({ identifier: "" })}
          >
            {"Send to another address"}
          </Button>
          <Button variant="default" size="lg" asChild className="w-full">
            <Link href="/login">{"Back to login"}</Link>
          </Button>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} noValidate aria-busy={isPending}>
      <div className="animate-fade-up flex flex-col gap-y-4">
        <Controller
          name="identifier"
          control={form.control}
          render={({ field, fieldState }) => {
            const errorId = fieldState.error ? `${identifierId}-error` : undefined;
            return (
              <Field>
                <Label htmlFor={identifierId}>{"Email or phone"}</Label>
                <Input
                  id={identifierId}
                  type="text"
                  autoComplete="email"
                  inputMode="email"
                  autoFocus
                  placeholder={`${"johndoe@mail.com"} or ${"+62 812-3456-7890"}`}
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
                <p className="text-sm text-muted-foreground">
                  {"We will send reset instructions if an account exists."}
                </p>
              </Field>
            );
          }}
        />
      </div>
      <div className="mt-6 flex flex-col gap-3 sm:mt-8">
        <SubmitButton variant="default" size="lg" className="w-full" loading={isPending}>
          {"Send reset link"}
        </SubmitButton>
        <Button variant="outline" size="lg" asChild className="w-full">
          <Link href="/login">{"Back to login"}</Link>
        </Button>
      </div>
    </form>
  );
}
