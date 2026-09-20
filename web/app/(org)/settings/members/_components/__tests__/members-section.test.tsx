import { screen, within } from "@testing-library/react";
import { HttpResponse, http } from "msw";
import { describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { MembersSection } from "../members-section";

describe("MembersSection", () => {
  it("renders member metrics and the member list from the API", async () => {
    renderWithProviders(<MembersSection />);

    expect(await screen.findByText("Alex Rivera")).toBeInTheDocument();
    expect(screen.getByText("June Park")).toBeInTheDocument();
    expect(screen.getByText("All members")).toBeInTheDocument();

    const grid = screen
      .getByText("Total members")
      .closest('[data-slot="metric-grid"]') as HTMLElement;
    const total = within(grid).getByText("Total members").closest('[data-slot="stat-card"]');
    const active = within(grid).getByText("Active members").closest('[data-slot="stat-card"]');
    const invited = within(grid).getByText("Invited").closest('[data-slot="stat-card"]');
    const admins = within(grid).getByText("Admins").closest('[data-slot="stat-card"]');

    expect(within(total as HTMLElement).getByText("3")).toBeInTheDocument();
    expect(within(active as HTMLElement).getByText("2")).toBeInTheDocument();
    expect(within(invited as HTMLElement).getByText("1")).toBeInTheDocument();
    expect(within(admins as HTMLElement).getByText("2")).toBeInTheDocument();
  });

  it("surfaces a retryable error state when the members request fails", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/members", () =>
        HttpResponse.json({ success: false, message: "Service unavailable." }, { status: 503 }),
      ),
    );

    renderWithProviders(<MembersSection />);

    expect(await screen.findByText("Service unavailable.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Retry" })).toBeInTheDocument();
  });
});
