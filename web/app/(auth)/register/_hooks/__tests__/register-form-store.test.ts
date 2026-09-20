import { describe, expect, it } from "vitest";
import { useRegisterFormStore } from "../register-form-store";

describe("register form store", () => {
  it("starts on the name step", () => {
    const state = useRegisterFormStore.getState();
    expect(state.step).toBe("name");
    expect(state.isFirstStep).toBe(true);
    expect(state.isLastStep).toBe(false);
  });

  it("setStep updates flags for first and last steps", () => {
    useRegisterFormStore.getState().setStep("name");
    expect(useRegisterFormStore.getState().isFirstStep).toBe(true);
    expect(useRegisterFormStore.getState().isLastStep).toBe(false);

    useRegisterFormStore.getState().setStep("password");
    expect(useRegisterFormStore.getState().isFirstStep).toBe(false);
    expect(useRegisterFormStore.getState().isLastStep).toBe(true);
  });

  it("flags no step as first or last for the middle steps", () => {
    useRegisterFormStore.getState().setStep("email");

    expect(useRegisterFormStore.getState().isFirstStep).toBe(false);
    expect(useRegisterFormStore.getState().isLastStep).toBe(false);
  });

  it("tracks pending and email state", () => {
    useRegisterFormStore.getState().setPending(true);
    useRegisterFormStore.getState().setEmail("a@b.co");

    const state = useRegisterFormStore.getState();
    expect(state.isPending).toBe(true);
    expect(state.email).toBe("a@b.co");
  });

  it("has no-op navigation that can be swapped by the form hook", async () => {
    const { goToNextStep, goToPrevStep } = useRegisterFormStore.getState();

    await expect(goToNextStep()).resolves.toBeUndefined();
    expect(goToPrevStep()).toBeUndefined();
  });
});
