"use client";

import { FormProvider } from "react-hook-form";
import {
  Questionnaire,
  QuestionnaireActions,
  QuestionnaireSubmit,
} from "@/components/questionnaire";
import { useOnboardingFormStore } from "../_hooks/onboarding-form-store";
import { useOnboardingForm } from "../_hooks/use-onboarding-form";
import { CompanyInfoStep } from "./company-info-step";
import { ONBOARDING_ITEMS } from "./onboarding-utils";

export default function OnboardingForm() {
  const { form, handleItemChange, handleSubmit } = useOnboardingForm({});
  const step = useOnboardingFormStore((s) => s.step);
  const isPending = useOnboardingFormStore((s) => s.isPending);

  return (
    <FormProvider {...form}>
      <Questionnaire
        items={ONBOARDING_ITEMS}
        item={step}
        onItemChange={(next) => {
          void handleItemChange(next);
        }}
        onSubmit={(event) => {
          void handleSubmit(event);
        }}
        aria-busy={isPending}
      >
        <CompanyInfoStep />
        <QuestionnaireActions>
          <QuestionnaireSubmit disabled={isPending}>{"Create Company"}</QuestionnaireSubmit>
        </QuestionnaireActions>
      </Questionnaire>
    </FormProvider>
  );
}
