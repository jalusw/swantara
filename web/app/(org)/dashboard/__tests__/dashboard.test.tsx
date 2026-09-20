import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DashboardOverview } from "../_components/dashboard-overview-section";

describe("DashboardOverview", () => {
  it("renders stats, revenue overview and greeting for the active org", async () => {
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("IDR 84,320.00")).length).toBeGreaterThan(0);
    expect(await screen.findByText("1,284")).toBeInTheDocument();
    expect((await screen.findAllByText("96")).length).toBeGreaterThan(0);
    expect((await screen.findAllByText("48")).length).toBeGreaterThan(0);
    expect(await screen.findByText("Good to see you, Alex!")).toBeInTheDocument();
    expect(screen.getByText("Revenue overview")).toBeInTheDocument();
  });

  it("opens the new record menu and shows the export button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("IDR 84,320.00")).length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: /new record/i }));
    expect(await screen.findByRole("menuitem", { name: /new customer/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new invoice/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /export report/i })).toBeInTheDocument();
  });

  it("renders service coverage grid with all business flows", async () => {
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect(await screen.findByText("Explore business flows")).toBeInTheDocument();
    expect(await screen.findByText("CRM")).toBeInTheDocument();
    expect(
      screen.getByText("Every service your organization runs — one tap to its module."),
    ).toBeInTheDocument();
    expect(screen.getByText("Point of Sale")).toBeInTheDocument();
    expect(screen.getByText("Fixed Assets")).toBeInTheDocument();
  });

  it("covers all quick-create actions for every service", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardOverview orgId="1" />);

    expect((await screen.findAllByText("IDR 84,320.00")).length).toBeGreaterThan(0);
    await user.click(screen.getByRole("button", { name: /new record/i }));
    expect(await screen.findByRole("menuitem", { name: /new supplier/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new item/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new sale order/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new purchase order/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new manufacturing order/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new project/i })).toBeInTheDocument();
    expect(screen.getByRole("menuitem", { name: /new employee/i })).toBeInTheDocument();
  });
});
