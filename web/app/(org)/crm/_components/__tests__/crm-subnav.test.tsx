import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { CrmSubNav } from "../crm-subnav";

beforeEach(() => {});

describe("CrmSubNav", () => {
  it("renders links to every CRM tab", () => {
    renderWithProviders(<CrmSubNav />);

    expect(screen.getByRole("link", { name: "Pipeline" })).toHaveAttribute("href", "/crm");
    expect(screen.getByRole("link", { name: "Leads" })).toHaveAttribute("href", "/crm/leads");
    expect(screen.getByRole("link", { name: "Opportunities" })).toHaveAttribute(
      "href",
      "/crm/opportunities",
    );
    expect(screen.getByRole("link", { name: "Activities" })).toHaveAttribute(
      "href",
      "/crm/activities",
    );
  });

  it("marks the current tab as active", () => {
    navigationMock.setPathname("/crm/leads");
    renderWithProviders(<CrmSubNav />);

    expect(screen.getByRole("link", { name: "Leads" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Pipeline" })).not.toHaveAttribute("aria-current");
  });

  it("focuses the tab when clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CrmSubNav />);

    await user.click(screen.getByRole("link", { name: "Activities" }));

    expect(screen.getByRole("link", { name: "Activities" })).toHaveFocus();
  });
});
