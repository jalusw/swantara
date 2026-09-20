import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import OrgProfilePage from "@/app/(org)/profile/page";
import CompanySettingsPage from "@/app/(org)/settings/page";
import GlobalErrorPreviewPage from "@/app/global-error";
import { renderWithProviders } from "@/lib/tests";

describe("placeholder pages", () => {
  it("renders the org settings page", async () => {
    const element = await CompanySettingsPage();
    renderWithProviders(element);
  });

  it("renders the profile settings page", () => {
    renderWithProviders(<OrgProfilePage />);
  });

  it("renders the global error preview with a reset action", () => {
    renderWithProviders(
      <GlobalErrorPreviewPage error={new Error("Preview error")} reset={() => {}} />,
    );

    expect(screen.getByRole("button", { name: /try again/i })).toBeInTheDocument();
  });
});
