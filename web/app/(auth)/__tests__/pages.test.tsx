import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import "@/lib/tests/component-mocks";
import ForgotPasswordPage from "../forgot-password/page";
import LoginPage from "../login/page";
import RegisterPage from "../register/page";
import ResetPasswordPage from "../reset-password/page";

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
    expect(screen.getByRole("link", { name: /register a new account/i })).toHaveAttribute(
      "href",
      "/register",
    );
  });

  it("renders the register page with the register form", async () => {
    const element = await RegisterPage();
    render(element);

    expect(screen.getByText("register-form")).toBeInTheDocument();
    expect(screen.getByText(/have an account already/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /log in/i })).toHaveAttribute("href", "/login");
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
