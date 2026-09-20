"use client";

import { LoaderCircleIcon } from "lucide-react";
import type * as React from "react";
import { useEffect, useId, useRef } from "react";
import {
  type ControllerFieldState,
  type ControllerRenderProps,
  type FieldPath,
  type FieldValues,
  FormProvider,
  type UseFormReturn,
  useController,
  useFormContext,
} from "react-hook-form";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Field } from "./field";
import { FieldFeedback } from "./field-feedback";
import { Label } from "./label";

export type FormProps<TFieldValues extends FieldValues> = Omit<
  React.ComponentProps<"form">,
  "onSubmit"
> & {
  form: UseFormReturn<TFieldValues>;
  onSubmit: (data: TFieldValues) => void | Promise<void>;
};

export function Form<TFieldValues extends FieldValues>({
  form,
  onSubmit,
  className,
  children,
  ...props
}: FormProps<TFieldValues>) {
  return (
    <FormProvider {...form}>
      <form
        data-slot="form"
        className={cn("flex flex-col gap-4", className)}
        onSubmit={form.handleSubmit(onSubmit)}
        noValidate
        {...props}
      >
        {children}
      </form>
    </FormProvider>
  );
}

type FormFieldChildren<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = (props: {
  field: ControllerRenderProps<TFieldValues, TName>;
  fieldState: ControllerFieldState;
  id: string;
  "aria-describedby"?: string;
}) => React.ReactNode;

export type FormFieldProps<
  TFieldValues extends FieldValues,
  TName extends FieldPath<TFieldValues>,
> = {
  name: TName;
  label?: string;
  description?: string;
  className?: string;
  children: FormFieldChildren<TFieldValues, TName>;
};

export function FormField<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>({
  name,
  label,
  description,
  className,
  children,
}: FormFieldProps<TFieldValues, TName>) {
  const { field, fieldState } = useController<TFieldValues, TName>({ name });
  const error = fieldState.error?.message;
  const inputId = useId();
  const feedbackId = `${inputId}-feedback`;

  return (
    <Field className={className}>
      {label ? (
        <Label htmlFor={inputId} className="mb-0 justify-start">
          {label}
        </Label>
      ) : null}
      {children({
        field,
        fieldState,
        id: inputId,
        "aria-describedby": error ? feedbackId : undefined,
      })}
      {description && !error ? (
        <p className="text-sm text-muted-foreground">{description}</p>
      ) : null}
      {error ? (
        <FieldFeedback visible intent="danger" id={feedbackId}>
          {error}
        </FieldFeedback>
      ) : null}
    </Field>
  );
}

export function SubmitButton({
  children,
  loadingLabel = "Saving…",
  loading,
  disabled,
  className,
  ...props
}: React.ComponentProps<typeof Button> & {
  loading?: boolean;
  loadingLabel?: string;
}) {
  const context = useFormContext();
  const isSubmitting = loading ?? context?.formState.isSubmitting ?? false;

  return (
    <Button
      data-slot="form-submit"
      type="submit"
      disabled={disabled || isSubmitting}
      aria-busy={isSubmitting || undefined}
      className={cn(className)}
      {...props}
    >
      {isSubmitting ? (
        <>
          <LoaderCircleIcon className="animate-spin" aria-hidden />
          {loadingLabel}
        </>
      ) : (
        children
      )}
    </Button>
  );
}

export function useUnsavedChanges(dirty: boolean, message?: string) {
  const savedDirty = useRef(dirty);
  savedDirty.current = dirty;

  useEffect(() => {
    if (!dirty) {
      return;
    }

    function beforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault();
      event.returnValue = "";
    }

    window.addEventListener("beforeunload", beforeUnload);
    return () => window.removeEventListener("beforeunload", beforeUnload);
  }, [dirty]);

  return {
    confirmNavigation: (continueFn: () => void) => {
      if (savedDirty.current) {
        const ok = window.confirm(message ?? "You have unsaved changes. Leave this page?");
        if (!ok) {
          return;
        }
      }
      continueFn();
    },
  };
}
