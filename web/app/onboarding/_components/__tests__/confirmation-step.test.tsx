import { screen } from "@testing-library/react";
import { FormProvider, useForm } from "react-hook-form";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import type { OnboardingFormSchema } from "../../_hooks/use-onboarding-form";
import { ConfirmationStep } from "../confirmation-step";

function renderConfirmation(defaultValues: Partial<OnboardingFormSchema> = {}) {
  const FormFrame = () => {
    const form = useForm<OnboardingFormSchema>({
      defaultValues: {
        name: "",
        countryCode: "",
        ...defaultValues,
      },
    });
    return (
      <FormProvider {...form}>
        <form>
          <ConfirmationStep />
        </form>
      </FormProvider>
    );
  };
  return renderWithProviders(<FormFrame />);
}

describe("ConfirmationStep", () => {
  it("shows a review prompt when no values are present", () => {
    renderConfirmation();

    expect(screen.getByText(/Review/i)).toBeInTheDocument();
  });

  it("lists the entered values", async () => {
    renderConfirmation({
      name: "Acme Inc",
      countryCode: "ID",
    });

    expect(screen.getByText("Acme Inc")).toBeInTheDocument();
    expect(screen.getByText(/Indonesia/)).toBeInTheDocument();
  });
});
