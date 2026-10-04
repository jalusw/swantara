import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { DashboardActions } from "../dashboard-actions";

describe("DashboardActions", () => {
  it("renders export and new-record actions", async () => {
    renderWithProviders(<DashboardActions />);

    expect(await screen.findByText("Ekspor Laporan")).toBeInTheDocument();
    expect(screen.getByText("Buat cepat")).toBeInTheDocument();
  });

  it("opens the new-record menu with entity links", async () => {
    const user = userEvent.setup();
    renderWithProviders(<DashboardActions />);

    await user.click(await screen.findByText("Buat cepat"));

    const customerItem = await screen.findByText("Pelanggan baru");
    expect(customerItem).toBeInTheDocument();
    expect(customerItem.closest("a")).toHaveAttribute("href", "/contacts");
  });
});
