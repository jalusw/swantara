import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { useRegisterFormStore } from "../../_hooks/register-form-store";
import RegisterForm from "../register-form";

async function fillNameStep(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/nama depan/i), "Jane");
  await user.type(screen.getByLabelText(/nama belakang/i), "Doe");
}

describe("RegisterForm", () => {
  it("registers, logs in and navigates to onboarding", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    expect(screen.getByLabelText(/nama depan/i)).toBeInTheDocument();

    await fillNameStep(user);
    await user.click(screen.getByRole("button", { name: /lanjut/i }));

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/email/i), "new@example.com");
    await user.click(screen.getByRole("button", { name: /lanjut/i }));

    expect(screen.getByLabelText(/^kata sandi$/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/^kata sandi$/i), "password1");
    await user.type(screen.getByLabelText(/konfirmasi kata sandi/i), "password1");
    await user.click(screen.getByRole("button", { name: /daftar/i }));

    await waitFor(() => {
      expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
    });
  });

  it("blocks the first step with an inline error", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    await user.click(screen.getByRole("button", { name: /lanjut/i }));

    expect(screen.getByText(/nama depan wajib diisi/i)).toBeInTheDocument();
    expect(useRegisterFormStore.getState().step).toBe("name");
  });

  it("walks backwards with the back button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    await fillNameStep(user);
    await user.click(screen.getByRole("button", { name: /lanjut/i }));

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /kembali/i }));

    expect(screen.getByLabelText(/nama depan/i)).toBeInTheDocument();
    expect(useRegisterFormStore.getState().step).toBe("name");
  });
});
