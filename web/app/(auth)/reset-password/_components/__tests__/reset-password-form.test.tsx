import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import ResetPasswordForm from "../reset-password-form";

describe("ResetPasswordForm", () => {
  it("renders both password fields and the submit button", () => {
    renderWithProviders(<ResetPasswordForm token="abc" />);

    expect(screen.getByLabelText(/^kata sandi baru$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^konfirmasi kata sandi baru$/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /atur ulang kata sandi/i })).toBeInTheDocument();
  });

  it("resets the password and navigates to /login", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ResetPasswordForm token="abc" />);

    await user.type(screen.getByLabelText(/^kata sandi baru$/i), "password1");
    await user.type(screen.getByLabelText(/^konfirmasi kata sandi baru$/i), "password1");
    await user.click(screen.getByRole("button", { name: /atur ulang kata sandi/i }));

    await waitFor(() => expect(navigationMock.push).toHaveBeenCalledWith("/login"), {
      timeout: 2000,
    });
  });

  it("shows a mismatch error when the passwords differ", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ResetPasswordForm token="abc" />);

    await user.type(screen.getByLabelText(/^kata sandi baru$/i), "password1");
    await user.type(screen.getByLabelText(/^konfirmasi kata sandi baru$/i), "password2");
    await user.click(screen.getByRole("button", { name: /atur ulang kata sandi/i }));

    expect(await screen.findByText(/tidak cocok/i)).toBeInTheDocument();
  });
});
