import { act, waitFor } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderHookWithProviders, server } from "@/lib/tests";
import { useResetPasswordForm, useResetPasswordFormSchema } from "../use-reset-password-form";

describe("useResetPasswordFormSchema", () => {
  it("rejects short passwords", async () => {
    const { result } = renderHookWithProviders(() => useResetPasswordFormSchema());
    const parsed = await result.current.safeParse({
      password: "short",
      passwordConfirmation: "short",
    });

    expect(parsed.success).toBe(false);
  });

  it("rejects a mismatched confirmation", async () => {
    const { result } = renderHookWithProviders(() => useResetPasswordFormSchema());
    const parsed = await result.current.safeParse({
      password: "password1",
      passwordConfirmation: "password2",
    });

    expect(parsed.success).toBe(false);
  });

  it("accepts a matching password pair", async () => {
    const { result } = renderHookWithProviders(() => useResetPasswordFormSchema());

    expect(
      (
        await result.current.safeParse({
          password: "password1",
          passwordConfirmation: "password1",
        })
      ).success,
    ).toBe(true);
  });
});

describe("useResetPasswordForm submission", () => {
  beforeEach(() => {
    vi.spyOn(toast, "error");
  });

  it("navigates to login on success with the token included", async () => {
    const { result } = renderHookWithProviders(() => useResetPasswordForm({ token: "abc123" }));
    await act(async () => {
      result.current.form.setValue("password", "password1");
      result.current.form.setValue("passwordConfirmation", "password1");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    await waitFor(() => expect(navigationMock.push).toHaveBeenCalledWith("/login"), {
      timeout: 2000,
    });
  });

  it("shows a server error toast on a 5xx response", async () => {
    server.use(
      http.post("*/api/v1/auth/password-reset", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useResetPasswordForm({}));
    await act(async () => {
      result.current.form.setValue("password", "password1");
      result.current.form.setValue("passwordConfirmation", "password1");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
    expect(navigationMock.push).not.toHaveBeenCalled();
  });

  it("shows a reset failed toast on a network error", async () => {
    server.use(http.post("*/api/v1/auth/password-reset", () => HttpResponse.error()));
    const { result } = renderHookWithProviders(() => useResetPasswordForm({}));
    await act(async () => {
      result.current.form.setValue("password", "password1");
      result.current.form.setValue("passwordConfirmation", "password1");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });
});
