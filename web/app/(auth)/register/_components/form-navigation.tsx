"use client";

import { ChevronLeft, Loader2 } from "lucide-react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { SubmitButton } from "@/components/form";

import { useRegisterFormStore } from "../_hooks/use-register-form";

export function FormNavigation() {
  const step = useRegisterFormStore((s) => s.step);
  const t = useTranslations("Auth");
  const tCommon = useTranslations("Common");
  const txCommon = tCommon as unknown as (key: string) => string;
  const isPending = useRegisterFormStore((s) => s.isPending);
  const isCheckingEmail = useRegisterFormStore((s) => s.isCheckingEmail);
  const goToNextStep = useRegisterFormStore((s) => s.goToNextStep);
  const goToPrevStep = useRegisterFormStore((s) => s.goToPrevStep);
  const isFirstStep = step === "name";
  const isLastStep = step === "password";

  return (
    <div className="flex items-center justify-between mt-8">
      {!isFirstStep ? (
        <Button
          type="button"
          variant="ghost"
          size="lg"
          onClick={goToPrevStep}
          className="flex items-center gap-x-1 px-4"
          aria-label={tCommon("back")}
        >
          <ChevronLeft className="size-4" aria-hidden="true" />
          {tCommon("back")}
        </Button>
      ) : (
        <div />
      )}
      {isLastStep ? (
        <SubmitButton variant="default" size="lg" loading={isPending}>
          {t("registerButton")}
        </SubmitButton>
      ) : (
        <Button
          type="button"
          variant="default"
          size="lg"
          onClick={goToNextStep}
          disabled={isCheckingEmail}
          className="px-8"
        >
          {isCheckingEmail ? (
            <Loader2 className="size-4 animate-spin" aria-hidden="true" />
          ) : (
            txCommon("next")
          )}
        </Button>
      )}
    </div>
  );
}
