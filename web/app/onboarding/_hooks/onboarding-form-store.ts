import { create } from "zustand";

export type OnboardingStep = "companyInfo";

export type OnboardingFormStore = {
  step: OnboardingStep;
  isPending: boolean;

  setStep: (step: OnboardingStep) => void;
  setPending: (isPending: boolean) => void;
};

export const useOnboardingFormStore = create<OnboardingFormStore>((set) => ({
  step: "companyInfo",
  isPending: false,

  setStep: (step) => set({ step }),
  setPending: (isPending) => set({ isPending }),
}));
