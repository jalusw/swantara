import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import ForgotPasswordForm from "../forgot-password-form";

describe("ForgotPasswordForm", () => {
  it("renders a single smart input for email or phone", () => {
    renderWithProviders(<ForgotPasswordForm />);

    expect(screen.getByLabelText(/email atau nomor telepon/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/johndoe@mail.com/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /kirim tautan reset/i })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /kembali ke masuk/i })).toBeInTheDocument();
  });

  it("shows the success screen after a successful email request", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email atau nomor telepon/i), "user@mail.com");
    await user.click(screen.getByRole("button", { name: /kirim tautan reset/i }));

    expect(await screen.findByText(/tautan reset terkirim/i)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: /kembali ke masuk/i })).toBeInTheDocument();
  });

  it("shows an inline error for an invalid identifier", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email atau nomor telepon/i), "abc");
    await user.click(screen.getByRole("button", { name: /kirim tautan reset/i }));

    expect(await screen.findByText(/nomor telepon yang valid/i)).toBeInTheDocument();
  });

  it("accepts a valid phone number", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ForgotPasswordForm />);

    await user.type(screen.getByLabelText(/email atau nomor telepon/i), "+62 812-3456-7890");
    await user.click(screen.getByRole("button", { name: /kirim tautan reset/i }));

    expect(await screen.findByText(/tautan reset terkirim/i)).toBeInTheDocument();
  });
});
