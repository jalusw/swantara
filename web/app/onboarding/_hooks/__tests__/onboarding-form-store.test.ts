import { describe, expect, it } from "vitest";
import { useOnboardingFormStore } from "../onboarding-form-store";

describe("onboarding form store", () => {
  it("starts on the company info step", () => {
    expect(useOnboardingFormStore.getState().step).toBe("companyInfo");
    expect(useOnboardingFormStore.getState().isPending).toBe(false);
  });

  it("setPending tracks submission state", () => {
    useOnboardingFormStore.getState().setPending(true);

    expect(useOnboardingFormStore.getState().isPending).toBe(true);
  });

  it("setStep switches the active step", () => {
    useOnboardingFormStore.getState().setStep("companyInfo");

    expect(useOnboardingFormStore.getState().step).toBe("companyInfo");
  });
});
