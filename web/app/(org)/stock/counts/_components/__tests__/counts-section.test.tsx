import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CountsSection } from "../counts-section";

const STAMP = "2026-01-01T00:00:00Z";

const counts = [
  {
    id: 11,
    organization_id: 1,
    name: "IC-0001",
    location_id: 9,
    state: "draft",
    count_date: "2026-02-01",
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 12,
    organization_id: 1,
    name: "IC-0002",
    location_id: 9,
    state: "posted",
    count_date: "2026-02-02",
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/stock-counts", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { counts } }),
    ),
  );
});

describe("CountsSection", () => {
  it("renders seeded inventory counts", async () => {
    renderWithProviders(<CountsSection />);

    expect(await screen.findByText("IC-0001")).toBeInTheDocument();
    expect(screen.getByText("IC-0002")).toBeInTheDocument();
  });

  it("filters rows through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CountsSection />);

    await screen.findByText("IC-0001");
    await user.type(screen.getByPlaceholderText("Cari opname…"), "IC-0002");

    expect((await screen.findAllByText("IC-0002")).length).toBeGreaterThan(0);
    expect(screen.queryByText("IC-0001")).not.toBeInTheDocument();
  });
});
