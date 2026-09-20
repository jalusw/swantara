import { act } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHookWithProviders, server } from "@/lib/tests";
import { useForgotPasswordForm, useForgotPasswordFormSchema } from "../use-forgot-password-form";

describe("useForgotPasswordFormSchema", () => {
  it("rejects an empty identifier", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordFormSchema());
    const parsed = await result.current.safeParse({
      identifier: "",
    });

    expect(parsed.success).toBe(false);
  });

  it("rejects a malformed email when it contains @", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordFormSchema());
    const parsed = await result.current.safeParse({
      identifier: "not-an-email",
    });

    expect(parsed.success).toBe(false);
  });

  it("accepts a valid email", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordFormSchema());

    expect(
      (
        await result.current.safeParse({
          identifier: "demo@swantara.local",
        })
      ).success,
    ).toBe(true);
  });

  it("rejects a short phone number", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordFormSchema());
    const parsed = await result.current.safeParse({
      identifier: "123",
    });

    expect(parsed.success).toBe(false);
  });

  it("accepts a valid phone number with formatting", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordFormSchema());

    expect(
      (
        await result.current.safeParse({
          identifier: "+62 812-3456-7890",
        })
      ).success,
    ).toBe(true);
  });
});

describe("useForgotPasswordForm submission", () => {
  beforeEach(() => {
    vi.spyOn(toast, "error");
  });

  it("marks the form as submitted after a successful request", async () => {
    const { result } = renderHookWithProviders(() => useForgotPasswordForm());
    await act(async () => {
      result.current.form.setValue("identifier", "demo@swantara.local");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.isSubmitted).toBe(true);
  });

  it("shows a server error toast on a 5xx response", async () => {
    server.use(
      http.post("*/api/v1/auth/password-reset/request", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useForgotPasswordForm());
    await act(async () => {
      result.current.form.setValue("identifier", "demo@swantara.local");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
    expect(result.current.isSubmitted).toBe(false);
  });

  it("shows a request failed toast on a network error", async () => {
    server.use(http.post("*/api/v1/auth/password-reset/request", () => HttpResponse.error()));
    const { result } = renderHookWithProviders(() => useForgotPasswordForm());
    await act(async () => {
      result.current.form.setValue("identifier", "demo@swantara.local");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });

  it("shows a request failed toast for other error statuses", async () => {
    server.use(
      http.post("*/api/v1/auth/password-reset/request", () =>
        HttpResponse.json({ success: false }, { status: 422 }),
      ),
    );
    const { result } = renderHookWithProviders(() => useForgotPasswordForm());
    await act(async () => {
      result.current.form.setValue("identifier", "demo@swantara.local");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });
});
