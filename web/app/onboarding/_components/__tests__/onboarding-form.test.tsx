import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { useOnboardingFormStore } from "../../_hooks/onboarding-form-store";
import OnboardingForm from "../onboarding-form";

describe("OnboardingForm", () => {
  beforeEach(() => {
    useOnboardingFormStore.setState({ step: "companyInfo", isPending: false });
  });

  it("renders the company info step first", () => {
    renderWithProviders(<OnboardingForm />);

    expect(screen.getByLabelText(/Nama perusahaan/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Negara/i)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Buat Perusahaan/i })).toBeInTheDocument();
  });

  it("creates the company with the derived standard", async () => {
    const user = userEvent.setup();
    let body: unknown;
    server.use(
      http.post("*/api/v1/organizations/quick", async ({ request }) => {
        body = await request.json();
        return HttpResponse.json(
          { success: true, data: { organization: { id: 1, name: "Acme Inc" } } },
          { status: 201 },
        );
      }),
    );
    renderWithProviders(<OnboardingForm />);

    await user.type(screen.getByLabelText(/Nama perusahaan/i), "Acme Inc");
    await user.click(screen.getByLabelText(/Negara/i));
    const listbox = await screen.findByRole("listbox");
    await user.click(within(listbox).getByText("Indonesia"));
    await user.click(screen.getByRole("button", { name: /Buat Perusahaan/i }));

    expect(body).toMatchObject({
      name: "Acme Inc",
      country_code: "ID",
      standard_code: "PSAK_EMKM",
    });
  });
});
