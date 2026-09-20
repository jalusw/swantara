import { act } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderHookWithProviders, server } from "@/lib/tests";
import { useRegisterFormStore } from "../register-form-store";
import { useRegisterForm } from "../use-register-form";

async function fillValidForm(form: ReturnType<typeof useRegisterForm>["form"]) {
  await act(async () => {
    form.setValue("firstName", "Jane");
    form.setValue("lastName", "Doe");
    form.setValue("email", "new@example.com");
    form.setValue("password", "password1");
    form.setValue("passwordConfirmation", "password1");
  });
}

describe("useRegisterForm email availability", () => {
  it("blocks advancing when the email is already taken", async () => {
    server.use(
      http.post("*/api/v1/auth/email/check", () =>
        HttpResponse.json({
          success: true,
          message: "Taken.",
          data: { available: false },
        }),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("email");
    await act(async () => {
      result.current.form.setValue("email", "taken@example.com");
    });

    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("email");
    expect(result.current.form.getFieldState("email").error).toBeDefined();
  });

  it("advances when the email check service is unreachable", async () => {
    server.use(http.post("*/api/v1/auth/email/check", () => HttpResponse.error()));
    const warningSpy = vi.spyOn(toast, "warning");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("email");
    await act(async () => {
      result.current.form.setValue("email", "new@example.com");
    });

    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(warningSpy).toHaveBeenCalled();
    expect(useRegisterFormStore.getState().step).toBe("password");
  });
});

describe("useRegisterForm submission errors", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("warns about rate limiting on a 429 response", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false }, { status: 429 }),
      ),
    );
    const warningSpy = vi.spyOn(toast, "warning");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(warningSpy).toHaveBeenCalled();
    expect(navigationMock.push).not.toHaveBeenCalled();
  });

  it("maps field errors onto the form", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json(
          {
            success: false,
            message: "Validation failed.",
            field_errors: [{ field: "firstName", message: "Too short." }],
          },
          { status: 422 },
        ),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("firstName").error).toBeDefined();
  });

  it("flags a taken email on an already-registered response", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json(
          { success: false, message: "Email already registered." },
          { status: 422 },
        ),
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
});
