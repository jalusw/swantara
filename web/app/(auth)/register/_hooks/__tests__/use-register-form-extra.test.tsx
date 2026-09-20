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

beforeEach(() => {
  vi.useFakeTimers();
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("useRegisterForm email check", () => {
  it("flags a taken email during the email step", async () => {
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

    expect(result.current.form.getFieldState("email").error).toBeDefined();
    expect(useRegisterFormStore.getState().step).toBe("email");
  });

  it("warns but advances when the email check fails", async () => {
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

  it("stays on the password step when advancing past the end", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("password");
    await fillValidForm(result.current.form);

    await act(async () => {
      await useRegisterFormStore.getState().goToNextStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("password");
  });

  it("stays on the name step when going back from the start", async () => {
    renderHookWithProviders(() => useRegisterForm({}));
    useRegisterFormStore.getState().setStep("name");

    await act(async () => {
      useRegisterFormStore.getState().goToPrevStep();
    });

    expect(useRegisterFormStore.getState().step).toBe("name");
  });

  it("syncs the watched email into the store", async () => {
    const { result } = renderHookWithProviders(() => useRegisterForm({}));

    await act(async () => {
      result.current.form.setValue("email", "sync@example.com");
    });

    expect(useRegisterFormStore.getState().email).toBe("sync@example.com");
  });
});

describe("useRegisterForm submission remainder", () => {
  it("falls back to login when auto-login throws", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new Error("offline"));
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(navigationMock.push).toHaveBeenCalledWith("/login");
  });

  it("flags email taken on a 422 already-registered message", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json(
          { success: false, message: "Email already registered." },
          { status: 422 },
        ),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("email").error).toBeDefined();
    expect(useRegisterFormStore.getState().step).toBe("email");
  });

  it("warns with retry seconds on rate limiting", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json(
          { success: false, message: "Slow down." },
          { status: 429, headers: { "retry-after": "30" } },
        ),
      ),
    );
    const warningSpy = vi.spyOn(toast, "warning");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(warningSpy).toHaveBeenCalled();
  });

  it("warns without retry seconds when the header is missing", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false, message: "Slow down." }, { status: 429 }),
      ),
    );
    const warningSpy = vi.spyOn(toast, "warning");
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(warningSpy).toHaveBeenCalled();
  });

  it("maps field errors onto the form", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json(
          {
            success: false,
            message: "Invalid.",
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

    expect(result.current.form.getFieldState("firstName").error?.message).toBe("Too short.");
  });

  it("flags email taken on a 400 email message", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false, message: "Email invalid." }, { status: 400 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("email").error).toBeDefined();
  });

  it("toasts a generic message on a 400 non-email message", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false, message: "Bad payload." }, { status: 400 }),
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

  it("falls back to a generic email error for unknown failures", async () => {
    server.use(
      http.post("*/api/v1/auth/register", () =>
        HttpResponse.json({ success: false, message: "Weird." }, { status: 418 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useRegisterForm({}));
    await fillValidForm(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("email").error).toBeDefined();
  });
});
