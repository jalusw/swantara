import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { useRegisterFormStore } from "../../_hooks/use-register-form";
import { FormNavigation } from "../form-navigation";

describe("FormNavigation", () => {
  it("shows only Next on the first step", () => {
    useRegisterFormStore.getState().setStep("name");

    renderWithProviders(<FormNavigation />);

    expect(screen.getByRole("button", { name: /lanjut/i })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /kembali/i })).not.toBeInTheDocument();
  });

  it("shows Back and Next on a middle step", () => {
    useRegisterFormStore.getState().setStep("email");

    renderWithProviders(<FormNavigation />);

    expect(screen.getByRole("button", { name: /kembali/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /lanjut/i })).toBeInTheDocument();
  });

  it("navigates backwards and forwards through steps", async () => {
    const user = userEvent.setup();
    useRegisterFormStore.getState().setStep("email");
    useRegisterFormStore.setState({
      goToPrevStep: () => useRegisterFormStore.getState().setStep("name"),
      goToNextStep: async () => useRegisterFormStore.getState().setStep("password"),
    });

    renderWithProviders(<FormNavigation />);

    await user.click(screen.getByRole("button", { name: /kembali/i }));
    expect(useRegisterFormStore.getState().step).toBe("name");

    await user.click(screen.getByRole("button", { name: /lanjut/i }));
    expect(useRegisterFormStore.getState().step).toBe("password");
  });

  it("shows a disabled submit button on the password step", () => {
    useRegisterFormStore.getState().setStep("password");

    renderWithProviders(<FormNavigation />);

    const submit = screen.getByRole("button", { name: /daftar/i });
    expect(submit).toBeInTheDocument();
  });
});
