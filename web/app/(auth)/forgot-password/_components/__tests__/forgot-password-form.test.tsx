import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import ForgotPasswordForm from "../forgot-password-form";

describe("ForgotPasswordForm", () => {
  it("renders a single smart input for email or phone", () => {
    renderWithProviders(<ForgotPasswordForm />);

    expect(screen.getByLabelText(/email or phone/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/johndoe@mail.com/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /send reset link/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /back to login/i })).toBeInTheDocument();
  });

  it("shows the success screen after a successful email request", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email or phone/i), "user@mail.com");
    await user.click(screen.getByRole("button", { name: /send reset link/i }));

    expect(await screen.findByText(/reset link sent/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /back to login/i })).toBeInTheDocument();
  });

  it("shows an inline error for an invalid identifier", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email or phone/i), "abc");
    await user.click(screen.getByRole("button", { name: /send reset link/i }));

    expect(await screen.findByText(/valid phone number/i)).toBeInTheDocument();
  });

  it("accepts a valid phone number", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email or phone/i), "+62 812-3456-7890");
    await user.click(screen.getByRole("button", { name: /send reset link/i }));

    expect(await screen.findByText(/reset link sent/i)).toBeInTheDocument();
  });
});
