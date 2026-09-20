import { create } from "zustand";
import type { RegistrationStep } from "./register-form.types";

export type RegisterFormStore = {
  step: RegistrationStep;
  isPending: boolean;
  isCheckingEmail: boolean;
  email: string;
  isFirstStep: boolean;
  isLastStep: boolean;

  setStep: (step: RegistrationStep) => void;
  setPending: (isPending: boolean) => void;
  setCheckingEmail: (isCheckingEmail: boolean) => void;
  setEmail: (email: string) => void;

  goToNextStep: () => Promise<void>;
  goToPrevStep: () => void;
};

export const useRegisterFormStore = create<RegisterFormStore>((set) => ({
  step: "name",
  isPending: false,
  isCheckingEmail: false,
  email: "",
  isFirstStep: true,
  isLastStep: false,

  setStep: (step) =>
    set({
      step,
      isFirstStep: step === "name",
      isLastStep: step === "password",
    }),
  setPending: (isPending) => set({ isPending }),
  setCheckingEmail: (isCheckingEmail) => set({ isCheckingEmail }),
  setEmail: (email) => set({ email }),

  goToNextStep: async () => {},
  goToPrevStep: () => {},
}));

export function isRegisterLastStep(step: RegistrationStep) {
  return step === "password";
}

export function isRegisterFirstStep(step: RegistrationStep) {
  return step === "name";
}
