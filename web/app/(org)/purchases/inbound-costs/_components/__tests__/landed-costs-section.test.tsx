import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { InboundCostsSection } from "../inbound-costs-section";

const costs = [
  { id: 1, name: "Freight January", date: "2026-01-05", state: "draft" },
  { id: 2, name: "Insurance February", date: "2026-02-10", state: "posted" },
];

function useInboundCostHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/inbound-costs", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { costs } }),
    ),
  );
}

beforeEach(() => {
  useInboundCostHandlers();
});

describe("InboundCostsSection", () => {
  it("renders costs with state badges", async () => {
    renderWithProviders(<InboundCostsSection orgId="1" />);

    expect(await screen.findByText("Freight January")).toBeInTheDocument();
    expect(screen.getByText("Insurance February")).toBeInTheDocument();
    expect(screen.getByText("Draf")).toBeInTheDocument();
    expect(screen.getByText("Diposting")).toBeInTheDocument();
  });

  it("filters costs through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<InboundCostsSection orgId="1" />);

    await screen.findByText("Freight January");
    await user.type(screen.getByPlaceholderText(/Cari biaya masuk/), "Insurance");

    expect(await screen.findByText("Insurance February")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("Freight January")).not.toBeInTheDocument());
  });
});
