import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { TaxYearsSection } from "../tax-years-section";

function useLocalYears() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/tax-years", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: {
          tax_years: [
            {
              id: 1,
              organization_id: 1,
              name: "FY 2025",
              date_start: "2025-01-01",
              date_end: "2025-12-31",
              state: "open",
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
            {
              id: 2,
              organization_id: 1,
              name: "FY 2024",
              date_start: "2024-01-01",
              date_end: "2024-12-31",
              state: "done",
              created_at: "2026-01-01T00:00:00Z",
              updated_at: "2026-01-01T00:00:00Z",
            },
          ],
        },
      }),
    ),
  );
}

beforeEach(() => {
  useLocalYears();
});

describe("TaxYearsSection", () => {
  it("renders seeded tax years", async () => {
    renderWithProviders(<TaxYearsSection orgId="1" />);

    expect(await screen.findByText("FY 2025")).toBeInTheDocument();
    expect(screen.getByText("FY 2024")).toBeInTheDocument();
  });

  it("filters tax years by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TaxYearsSection orgId="1" />);

    await screen.findByText("FY 2025");
    await user.type(screen.getByPlaceholderText("Cari tahun pajak..."), "2024");

    expect(await screen.findByText("FY 2024")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<TaxYearsSection orgId="1" />);

    await screen.findByText("FY 2025");
    await user.click(screen.getByRole("button", { name: "Tambah tahun pajak" }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Buat tahun pajak")).toBeInTheDocument();
  });
});
