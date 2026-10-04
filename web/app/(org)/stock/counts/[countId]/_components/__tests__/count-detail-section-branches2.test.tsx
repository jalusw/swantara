import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { CountDetail } from "../count-detail-section";

const STAMP = "2026-01-01T00:00:00Z";

function seedCount(count: unknown, lines: unknown[]) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/stock-counts/:id", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { count } }),
    ),
    http.get("*/api/v1/organizations/:organizationId/stock-counts/:id/lines", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { lines } }),
    ),
  );
}

const baseCount = {
  id: 11,
  organization_id: 1,
  name: "IC-0011",
  location_id: 9,
  state: "draft",
  count_date: "2026-02-01",
  created_at: STAMP,
  updated_at: STAMP,
};

const baseLines = [
  {
    id: 71,
    stock_count_id: 11,
    item_id: 5,
    batch_id: null,
    theoretical_qty: 10,
    counted_qty: 12,
    diff_qty: 2,
  },
  {
    id: 72,
    stock_count_id: 11,
    item_id: 6,
    batch_id: null,
    theoretical_qty: 5,
    counted_qty: 5,
    diff_qty: 0,
  },
];

beforeEach(() => {});

describe("CountDetail branches2", () => {
  it("renders the not-found branch for unknown counts", async () => {
    seedCount(null, []);
    renderWithProviders(<CountDetail orgId="1" countId="99" />);

    expect(await screen.findByText("Opname tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders zero-diff and non-zero diff branches", async () => {
    seedCount(baseCount, baseLines);
    renderWithProviders(<CountDetail orgId="1" countId="11" />);

    await screen.findAllByText("IC-0011");
    expect(screen.getByText("Pratinjau selisih")).toBeInTheDocument();
  });

  it("posts the count through the confirm branch", async () => {
    seedCount(baseCount, baseLines);
    let postCalls = 0;
    server.use(
      http.post("*/api/v1/organizations/:organizationId/stock-counts/:id/post", () => {
        postCalls += 1;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const user = userEvent.setup();
    renderWithProviders(<CountDetail orgId="1" countId="11" />);

    await screen.findAllByText("IC-0011");
    await user.click(screen.getByRole("button", { name: "Posting opname" }));
    const confirmButtons = await screen.findAllByRole("button", { name: "Posting opname" });
    const last = confirmButtons[confirmButtons.length - 1];
    if (last) {
      await user.click(last);
    }

    await waitFor(() => expect(postCalls).toBe(1));
  });

  it("hides post for non-draft counts", async () => {
    seedCount({ ...baseCount, state: "posted" }, []);
    renderWithProviders(<CountDetail orgId="1" countId="11" />);

    await screen.findAllByText("IC-0011");
    expect(screen.queryByRole("button", { name: "Posting opname" })).not.toBeInTheDocument();
  });
});
