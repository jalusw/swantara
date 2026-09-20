import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DashboardActions } from "../dashboard-actions";

describe("DashboardActions", () => {
  it("renders export and new-record actions", async () => {
    renderWithProviders(<DashboardActions />);

    expect(await screen.findByText("Export report")).toBeInTheDocument();
    expect(screen.getByText("New record")).toBeInTheDocument();
  });

  it("opens the new-record menu with entity links", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardActions />);

    await user.click(await screen.findByText("New record"));

    const customerItem = await screen.findByText("New customer");
    expect(customerItem).toBeInTheDocument();
    expect(customerItem.closest("a")).toHaveAttribute("href", "/contacts");
  });
});
