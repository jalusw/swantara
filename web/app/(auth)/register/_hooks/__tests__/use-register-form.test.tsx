import { act } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderHookWithProviders, server } from "@/lib/tests";
import { useRegisterFormStore } from "../register-form-store";
import { useRegisterForm } from "../use-register-form";

describe("useRegisterForm navigation", () => {
  it("resets the store to the name step on mount", () => {
    useRegisterFormStore.getState().setStep("password");

    renderHookWithProviders(() => useRegisterForm({}));

    expect(useRegisterFormStore.getState().step).toBe("name");
  });

  it("advances from the name step when the fields are valid", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));

    await act(async () => {
      result.current.form.setValue("firstName", "Jane");
      result.current.form.setValue("lastName", "Doe");
    });
    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("email");
  });

  it("blocks advancing when required fields are empty", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));

    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("name");
    expect(result.current.form.getFieldState("firstName").error).toBeDefined();
  });

  it("advances from the email step when the email is present", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("email");

    await act(async () => {
      result.current.form.setValue("email", "new@example.com");
    });
    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("password");
  });

  it("walks backwards through the steps", async () => {
    renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("email");

    await act(async () => {
      useRegisterFormStore.getState().goToPrevStep();
    });
    expect(useRegisterFormStore.getState().step).toBe("name");

    useRegisterFormStore.getState().setStep("password");
    await act(async () => {
      useRegisterFormStore.getState().goToPrevStep();
    });
    expect(useRegisterFormStore.getState().step).toBe("email");
  });
});

describe("useRegisterForm submission", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  async function fillValidForm(form: ReturnType<typeof useRegisterForm>["form"]) {
    await act(async () => {
      form.setValue("firstName", "Jane");
      form.setValue("lastName", "Doe");
      form.setValue("email", "new@example.com");
      form.setValue("password", "password1");
      form.setValue("passwordConfirmation", "password1");
    });
  }

  it("registers, logs in and navigates to onboarding", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
  });

  it("falls back to the login page when auto-login fails", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({ success: false }, { status: 401 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(navigationMock.push).toHaveBeenCalledWith("/login");
  });

  it("flags a taken email on a conflict response", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false }, { status: 409 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("password");
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("email").error).toBeDefined();
    expect(navigationMock.push).not.toHaveBeenCalled();
  });

  it("surfaces a server error toast for a 5xx response", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const errorSpy = vi.spyOn(toast, "error");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(errorSpy).toHaveBeenCalled();
  });

  it("surfaces a network error toast when the request fails offline", async () => {
    server.use(http.post("*/api/v1/auth/register", () => HttpResponse.error()));
    const errorSpy = vi.spyOn(toast, "error");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(errorSpy).toHaveBeenCalled();
  });

  it("always clears the pending flag after submission", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(useRegisterFormStore.getState().isPending).toBe(false);
  });
});
