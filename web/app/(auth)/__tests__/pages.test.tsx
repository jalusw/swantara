import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import "@/lib/tests/component-mocks";
import ForgotPasswordPage from "../forgot-password/page";
import LoginPage from "../login/page";
import RegisterPage from "../register/page";
import ResetPasswordPage from "../reset-password/page";

// Server components call `getTranslations` from `next-intl/server`, which throws
// outside a React Server Components environment (vitest resolves the client
// build). Resolve the same keys from the default test locale (id) instead.
vi.mock("next-intl/server", async () => {
  const loaded = (await import("@/messages/id.json")) as unknown as Record<
    string,
    Record<string, string>
  >;
  const messages: Record<string, Record<string, string>> = (
    loaded as { default?: Record<string, Record<string, string>> }
  ).default ?? loaded;
  return {
    getTranslations: async (namespace: string) => {
      const ns = messages[namespace];
      if (!ns) {
        throw new Error(`Missing messages namespace: ${namespace}`);
      }
      return (key: string, values?: Record<string, string | number>) => {
        const message = ns[key];
        if (message === undefined) {
          throw new Error(`Missing message: ${namespace}.${key}`);
        }
        if (!values) return message;
        return message.replace(/\{(\w+)\}/g, (match, name: string) =>
          values[name] !== undefined ? String(values[name]) : match,
        );
      };
    },
  };
});

vi.mock("../login/_components/login-form", () => ({
  default: () => <div>login-form</div>,
}));

vi.mock("../register/_components/register-form", () => ({
  default: () => <div>register-form</div>,
}));

vi.mock("../forgot-password/_components/forgot-password-form", () => ({
  default: () => <div>forgot-password-form</div>,
}));

vi.mock("../reset-password/_components/reset-password-form", () => ({
  default: () => <div>reset-password-form</div>,
}));

vi.mock("../_components/auth-shell", () => ({
  default: ({ children }: { children: React.ReactNode }) => (
    <div>
      <span>auth-shell</span>
      {children}
    </div>
  ),
}));

describe("auth pages", () => {
  it("renders the login page with links and the login form", async () => {
    const element = await LoginPage();
    render(element);

    expect(screen.getByText("login-form")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /daftarkan akun baru/i })).toHaveAttribute(
      "href",
      "/register",
    );
  });

  it("renders the register page with the register form", async () => {
    const element = await RegisterPage();
    render(element);

    expect(screen.getByText("register-form")).toBeInTheDocument();
    expect(screen.getByText(/sudah punya akun/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /masuk/i })).toHaveAttribute("href", "/login");
  });

  it("renders the forgot password page inside the auth shell", async () => {
    const element = await ForgotPasswordPage();
    render(element);

    expect(screen.getByText("forgot-password-form")).toBeInTheDocument();
    expect(screen.getByText("auth-shell")).toBeInTheDocument();
  });

  it("renders the reset password page with the token from search params", async () => {
    const element = await ResetPasswordPage({
      searchParams: Promise.resolve({ token: "tok-123" }),
    });
    render(element);

    expect(screen.getByText("reset-password-form")).toBeInTheDocument();
  });
});
