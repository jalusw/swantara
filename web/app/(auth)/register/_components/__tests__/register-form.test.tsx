import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { useRegisterFormStore } from "../../_hooks/register-form-store";
import RegisterForm from "../register-form";

async function fillNameStep(user: ReturnType<typeof userEvent.setup>) {
  await user.type(screen.getByLabelText(/first name/i), "Jane");
  await user.type(screen.getByLabelText(/last name/i), "Doe");
}

describe("RegisterForm", () => {
  it("registers, logs in and navigates to onboarding", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    expect(screen.getByLabelText(/first name/i)).toBeInTheDocument();

    await fillNameStep(user);
    await user.click(screen.getByRole("button", { name: /next/i }));

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/email/i), "new@example.com");
    await user.click(screen.getByRole("button", { name: /next/i }));

    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();

    await user.type(screen.getByLabelText(/^password$/i), "password1");
    await user.type(screen.getByLabelText(/password confirmation/i), "password1");
    await user.click(screen.getByRole("button", { name: /register/i }));

    await waitFor(() => {
      expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
    });
  });

  it("blocks the first step with an inline error", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    await user.click(screen.getByRole("button", { name: /next/i }));

    expect(screen.getByText(/first name is required/i)).toBeInTheDocument();
    expect(useRegisterFormStore.getState().step).toBe("name");
  });

  it("walks backwards with the back button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<RegisterForm />);

    await fillNameStep(user);
    await user.click(screen.getByRole("button", { name: /next/i }));

    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /back/i }));

    expect(screen.getByLabelText(/first name/i)).toBeInTheDocument();
    expect(useRegisterFormStore.getState().step).toBe("name");
  });
});
