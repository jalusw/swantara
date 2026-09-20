import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import LoginForm from "../login-form";

describe("LoginForm", () => {
  it("renders email, password and submit button", () => {
    renderWithProviders(<LoginForm />);

    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^login$/i })).toBeInTheDocument();
  });

  it("submits valid credentials and navigates to /onboarding", async () => {
    const user = userEvent.setup();
    renderWithProviders(<LoginForm />);

    await user.type(screen.getByLabelText(/^email$/i), "user@mail.com");
    await user.type(screen.getByLabelText(/^password$/i), "password1");
    await user.click(screen.getByRole("button", { name: /^login$/i }));

    expect(navigationMock.push).toHaveBeenCalledWith("/onboarding");
  });
});
