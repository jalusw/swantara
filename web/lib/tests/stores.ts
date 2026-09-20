import { useRegisterFormStore } from "@/app/(auth)/register/_hooks/register-form-store";
import { useOnboardingFormStore } from "@/app/onboarding/_hooks/onboarding-form-store";
import { NORMAL, useDensityStore } from "@/stores/density.store";
import { useDigitalFingerprintStore } from "@/stores/digital-fingerprint.store";
import { DARK, useThemeStore } from "@/stores/theme.store";

const themeInitialState = {
  theme: DARK,
  resolvedTheme: DARK,
} as const;

const fingerprintInitialState = {
  deviceName: null,
  fingerprint: null,
  browser: null,
  os: null,
} as const;

const densityInitialState = {
  density: NORMAL,
} as const;

const registerFormInitialState = {
  step: "name",
  isPending: false,
  email: "",
  isFirstStep: true,
  isLastStep: false,
  goToNextStep: async () => {},
  goToPrevStep: () => {},
} as const;

const onboardingFormInitialState = {
  step: "companyInfo",
  isPending: false,
  goToNextStep: async () => {},
  goToPrevStep: () => {},
} as const;

export function resetStores() {
  useThemeStore.setState(themeInitialState);
  useDensityStore.setState(densityInitialState);
  useDigitalFingerprintStore.setState(fingerprintInitialState);
  useRegisterFormStore.setState(registerFormInitialState);
  useOnboardingFormStore.setState(onboardingFormInitialState);
}
