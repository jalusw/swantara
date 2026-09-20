import { act } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { navigationMock, renderHookWithProviders, server } from "@/lib/tests";
import { useLoginForm, useLoginFormSchema } from "../use-login-form";

function renderLogin() {
  return renderHookWithProviders(() => useLoginForm({}));
}

async function submitValid(handleSubmit: () => Promise<void> | void) {
  await act(async () => {
    await handleSubmit();
  });
}

describe("useLoginFormSchema", () => {
  it("rejects an empty email and password", async () => {
    const { result } = renderHookWithProviders(() => useLoginFormSchema());
    const parsed = await result.current.safeParse({ email: "", password: "" });

    expect(parsed.success).toBe(false);
    if (!parsed.success) {
      expect(parsed.error.issues.map((issue) => issue.path[0])).toEqual(["email", "password"]);
    }
  });

  it("accepts a valid email and password", async () => {
    const { result } = renderHookWithProviders(() => useLoginFormSchema());

    expect(
      (
        await result.current.safeParse({
          email: "demo@swantara.local",
          password: "secret",
        })
      ).success,
    ).toBe(true);
  });
});

describe("useLoginForm submission", () => {
  beforeEach(() => {
    vi.spyOn(toast, "error");
  });

  it("navigates to onboarding on a successful login", async () => {
    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
  });

  it("shows invalid credentials on a 401", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({ success: false }, { status: 401 }),
      ),
    );
    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "wrong");
    });

    await submitValid(result.current.handleSubmit);

    expect(toast.error).toHaveBeenCalled();
    expect(navigationMock.push).not.toHaveBeenCalled();
  });

  it("shows a server error toast on a 5xx response", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(toast.error).toHaveBeenCalled();
  });

  it("shows a network error toast when the request fails offline", async () => {
    server.use(http.post("*/api/v1/auth/login", () => HttpResponse.error()));
    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(toast.error).toHaveBeenCalled();
  });

  it("shows a generic failure toast for other error statuses", async () => {
    server.use(
      http.post("*/api/v1/auth/login", () =>
        HttpResponse.json({ success: false }, { status: 422 }),
      ),
    );
    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(toast.error).toHaveBeenCalled();
  });

  it("does not navigate when login validation fails", async () => {
    const { result } = renderLogin();

    await submitValid(result.current.handleSubmit);

    expect(navigationMock.push).not.toHaveBeenCalled();
  });
});

describe("useLoginForm with credential management", () => {
  beforeEach(() => {
    vi.spyOn(toast, "error");
  });

  it("pre-fills the form from a stored password credential", async () => {
    class FakeCredential {
      id = "saved@example.com";
      password = "saved-pass";
    }
    vi.stubGlobal("PasswordCredential", FakeCredential);
    Object.defineProperty(navigator, "credentials", {
      configurable: true,
      value: {
        get: vi.fn().mockResolvedValue(new FakeCredential()),
        store: vi.fn(),
      },
    });

    const { result } = renderLogin();

    await act(async () => {
      await Promise.resolve();
    });

    expect(result.current.form.getValues("email")).toBe("saved@example.com");
    expect(result.current.form.getValues("password")).toBe("saved-pass");
  });

  it("stores the credentials after a successful login", async () => {
    class FakeCredential {
      constructor(
        public id: string,
        public password: string,
      ) {}
    }
    const storeMock = vi.fn().mockResolvedValue(undefined);
    vi.stubGlobal("PasswordCredential", FakeCredential);
    Object.defineProperty(navigator, "credentials", {
      configurable: true,
      value: { get: vi.fn().mockResolvedValue(null), store: storeMock },
    });

    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(storeMock).toHaveBeenCalled();
    expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
  });

  it("ignores a failing credential fetch", async () => {
    vi.stubGlobal("PasswordCredential", class {});
    Object.defineProperty(navigator, "credentials", {
      configurable: true,
      value: { get: vi.fn().mockRejectedValue(new Error("denied")) },
    });

    const { result } = renderLogin();

    await act(async () => {
      await Promise.resolve();
    });

    expect(result.current.form.getValues("email")).toBe("");
  });

  it("survives a failing credential store without blocking navigation", async () => {
    class FakeCredential {
      constructor(
        public id: string,
        public password: string,
      ) {}
    }
    vi.stubGlobal("PasswordCredential", FakeCredential);
    Object.defineProperty(navigator, "credentials", {
      configurable: true,
      value: {
        get: vi.fn().mockResolvedValue(null),
        store: vi.fn().mockRejectedValue(new Error("denied")),
      },
    });

    const { result } = renderLogin();
    await act(async () => {
      result.current.form.setValue("email", "demo@swantara.local");
      result.current.form.setValue("password", "secret");
    });

    await submitValid(result.current.handleSubmit);

    expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
  });
});
