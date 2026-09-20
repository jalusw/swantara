import { act } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { toast } from "sonner";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderHookWithProviders, server } from "@/lib/tests";
import { useOnboardingFormStore } from "../onboarding-form-store";
import { useOnboardingForm } from "../use-onboarding-form";

function renderOnboarding() {
  return renderHookWithProviders(() => useOnboardingForm({}));
}

async function fillCompanyInfo(form: ReturnType<typeof useOnboardingForm>["form"]) {
  await act(async () => {
    form.setValue("name", "Acme Inc");
    form.setValue("countryCode", "ID");
  });
}

describe("useOnboardingForm navigation", () => {
  beforeEach(() => {
    useOnboardingFormStore.setState({ step: "companyInfo", isPending: false });
  });

  it("starts on the company info step", () => {
    renderOnboarding();

    expect(useOnboardingFormStore.getState().step).toBe("companyInfo");
  });

  it("blocks submission when the company name is empty", async () => {
    const { result } = renderOnboarding();

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(result.current.form.getFieldState("name").error).toBeDefined();
    expect(useOnboardingFormStore.getState().isPending).toBe(false);
  });
});

describe("useOnboardingForm submission", () => {
  beforeEach(() => {
    useOnboardingFormStore.setState({ step: "companyInfo", isPending: false });
    vi.spyOn(toast, "error");
  });

  it("sends PSAK_EMKM for an Indonesian business", async () => {
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
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(body).toMatchObject({
      name: "Acme Inc",
      country_code: "ID",
      standard_code: "PSAK_EMKM",
    });
    expect(body).not.toHaveProperty("business_size");
  });

  it("sends IFRS for a business outside Indonesia", async () => {
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
    const { result } = renderOnboarding();
    await act(async () => {
      result.current.form.setValue("name", "Acme Inc");
      result.current.form.setValue("countryCode", "US");
    });

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(body).toMatchObject({ standard_code: "IFRS" });
  });

  it("shows a server error toast and flags the name on other failures", async () => {
    server.use(
      http.post("*/api/v1/organizations/quick", () =>
        HttpResponse.json({ success: false }, { status: 422 }),
      ),
    );
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });

  it("surfaces the server message on a 422", async () => {
    server.use(
      http.post("*/api/v1/organizations/quick", () =>
        HttpResponse.json(
          {
            success: false,
            message: "Unknown accounting standard code.",
            error_code: "ERR_UNPROCESSABLE",
            field_errors: [
              { field: "standard_code", code: "UNKNOWN", message: "Unknown standard." },
            ],
          },
          { status: 422 },
        ),
      ),
    );
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalledWith("Unknown accounting standard code.");
  });

  it("shows a server error toast on a 5xx response", async () => {
    server.use(
      http.post("*/api/v1/organizations/quick", () =>
        HttpResponse.json({ success: false }, { status: 500 }),
      ),
    );
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });

  it("shows a network error toast when the request fails offline", async () => {
    server.use(http.post("*/api/v1/organizations/quick", () => HttpResponse.error()));
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(toast.error).toHaveBeenCalled();
  });

  it("clears the pending flag after submission", async () => {
    const { result } = renderOnboarding();
    await fillCompanyInfo(result.current.form);

    await act(async () => {
      await result.current.handleSubmit();
    });

    expect(useOnboardingFormStore.getState().isPending).toBe(false);
  });
});
