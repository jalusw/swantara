import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CountDetail } from "../count-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

const count = {
  id: 11,
  organization_id: 1,
  name: "IC-0011",
  location_id: 9,
  state: "draft",
  count_date: "2026-02-01",
  created_at: STAMP,
  updated_at: STAMP,
};

const lines = [
  {
    id: 71,
    stock_count_id: 11,
    item_id: 5,
    batch_id: null,
    theoretical_qty: 10,
    counted_qty: 12,
    diff_qty: 2,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/stock-counts/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { count } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-counts/:id/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
  );
});

describe("CountDetail", () => {
  it("renders the count with its diff preview", async () => {
    renderWithProviders(<CountDetail orgId="1" countId="11" />);

    expect((await screen.findAllByText("IC-0011")).length).toBeGreaterThan(0);
    expect(screen.getByText("Pratinjau selisih")).toBeInTheDocument();
  });

  it("opens the posting confirmation dialog", async () => {
    const user = userEvent.setup();
    renderWithProviders(<CountDetail orgId="1" countId="11" />);

    await screen.findAllByText("IC-0011");
    await user.click(screen.getByRole("button", { name: "Posting opname" }));

    expect(await screen.findByText(/Yakin ingin memposting/)).toBeInTheDocument();
  });
});
