import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ReorderRulesSection } from "../reorder-rules-section";

const rules = [
  {
    id: 1,
    item_id: 101,
    warehouse_id: null,
    location_id: null,
    min_qty: 5,
    max_qty: 50,
    qty_multiple: 1,
    lead_time_days: 7,
    active: true,
  },
  {
    id: 2,
    item_id: 202,
    warehouse_id: null,
    location_id: null,
    min_qty: 10,
    max_qty: 100,
    qty_multiple: 5,
    lead_time_days: null,
    active: false,
  },
];

function useReorderRuleHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/reorder-rules", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { rules } }),
    ),
  );
}

beforeEach(() => {
  useReorderRuleHandlers();
});

describe("ReorderRulesSection", () => {
  it("renders rules with quantities and status badges", async () => {
    renderWithProviders(<ReorderRulesSection orgId="1" />);

    expect(await screen.findByText("101")).toBeInTheDocument();
    expect(screen.getByText("202")).toBeInTheDocument();
    expect(screen.getByText("Inactive")).toBeInTheDocument();
  });

  it("opens the create dialog from the add button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ReorderRulesSection orgId="1" />);

    await screen.findByText("101");
    await user.click(screen.getByRole("button", { name: "Add rule" }));

    expect(await screen.findByText("New reorder rule")).toBeInTheDocument();
  });
});
