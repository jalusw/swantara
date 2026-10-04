"use client";

import { useTranslations } from "next-intl";
import { FormProvider } from "react-hook-form";
import { cn } from "@/lib/utils/style";
import { useRegisterForm, useRegisterFormStore } from "../_hooks/use-register-form";
import { EmailStep } from "./email-step";
import { FormNavigation } from "./form-navigation";
import { NameStep } from "./name-step";
import { PasswordStep } from "./password-step";

const stepOrder = ["name", "email", "password"] as const;

function StepProgress({ step }: { step: string }) {
  const t = useTranslations("Auth");
  const tx = t as unknown as (key: string, values?: Record<string, string | number>) => string;
  const current = stepOrder.indexOf(step as (typeof stepOrder)[number]) + 1;
  const total = stepOrder.length;
  const stepLabel =
    step === "name" ? tx("firstName") : step === "email" ? t("email") : t("password");
  return (
    <div className="mb-6" aria-live="polite" aria-atomic="true">
      <div className="flex items-center gap-2">
        {stepOrder.map((s, idx) => (
          <div
            key={s}
            className={cn(
              "h-1.5 flex-1 rounded-full transition-colors",
              idx < current ? "bg-primary" : "bg-border",
            )}
            aria-hidden="true"
          />
        ))}
      </div>
      <p className="mt-2 text-sm text-muted-foreground">
        {tx("stepCounter", { current, total })} — {stepLabel}
      </p>
    </div>
  );
}

export default function RegisterForm() {
  const { form, handleSubmit } = useRegisterForm({});
  const step = useRegisterFormStore((s) => s.step);

  return (
    <FormProvider {...form}>
      <form
        noValidate
        aria-busy={useRegisterFormStore((s) => s.isPending)}
        onSubmit={(e) => {
          const currentStep = useRegisterFormStore.getState().step;
          const isLastStep = currentStep === "password";
          if (!isLastStep) {
            e.preventDefault();
            void useRegisterFormStore.getState().goToNextStep();
            return;
          }
          void handleSubmit(e);
        }}
        onKeyDown={(e) => {
          const currentStep = useRegisterFormStore.getState().step;
          const isLastStep = currentStep === "password";
          if (e.key === "Enter" && !isLastStep) {
            const target = e.target as HTMLElement;
            if (target.tagName === "TEXTAREA" || target.isContentEditable) return;
            e.preventDefault();
            void useRegisterFormStore.getState().goToNextStep();
          }
        }}
        className="w-full"
      >
        <StepProgress step={step} />
        <NameStep />
        <EmailStep />
        <PasswordStep />
        <FormNavigation />
      </form>
    </FormProvider>
  );
}
