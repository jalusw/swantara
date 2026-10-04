import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it } from "vitest";
import { navigationMock, renderWithProviders } from "@/lib/tests";
import { CrmSubNav } from "../crm-subnav";

beforeEach(() => {});

describe("CrmSubNav", () => {
  it("renders links to every CRM tab", () => {
    renderWithProviders(<CrmSubNav />);

    expect(screen.getByRole("link", { name: "Pipa Penjualan" })).toHaveAttribute("href", "/crm");
    expect(screen.getByRole("link", { name: "Prospek" })).toHaveAttribute("href", "/crm/leads");
    expect(screen.getByRole("link", { name: "Peluang" })).toHaveAttribute(
      "href",
      "/crm/opportunities",
    );
    expect(screen.getByRole("link", { name: "Aktivitas" })).toHaveAttribute(
      "href",
      "/crm/activities",
    );
  });

  it("marks the current tab as active", () => {
    navigationMock.setPathname("/crm/leads");
    renderWithProviders(<CrmSubNav />);

    expect(screen.getByRole("link", { name: "Prospek" })).toHaveAttribute("aria-current", "page");
    expect(screen.getByRole("link", { name: "Pipa Penjualan" })).not.toHaveAttribute(
      "aria-current",
    );
  });

  it("focuses the tab when clicked", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CrmSubNav />);

    await user.click(screen.getByRole("link", { name: "Aktivitas" }));

    expect(screen.getByRole("link", { name: "Aktivitas" })).toHaveFocus();
  });
});
