import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { setActiveOrg } from "@/lib/server/active-org-actions";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { OrgActiveProvider } from "@/providers/org-active-provider";
import { OrgBrand } from "../org-brand";

vi.mock("@/lib/server/active-org-actions", () => ({
  setActiveOrg: vi.fn(async () => {}),
  clearActiveOrg: vi.fn(async () => {}),
}));

describe("OrgBrand", () => {
  it("lists the actor's organizations and switches to the selected one", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <OrgActiveProvider orgId={1}>
        <OrgBrand />
      </OrgActiveProvider>,
    );

    expect(await screen.findByText("Acme Inc")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Acme Inc/ }));
    await user.click(await screen.findByRole("menuitem", { name: /PT Nusantara/ }));

    expect(setActiveOrg).toHaveBeenCalledWith(2);
    expect(navigationMock.push).toHaveBeenCalledWith("/dashboard");
  });

  it("opens the create organization dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <OrgActiveProvider orgId={1}>
        <OrgBrand />
      </OrgActiveProvider>,
    );

    await user.click(await screen.findByRole("button"));
    await user.click(await screen.findByRole("menuitem", { name: /Create organization/ }));

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Create organization")).toBeInTheDocument();
    expect(navigationMock.push).not.toHaveBeenCalledWith("/onboarding");
  });
});
