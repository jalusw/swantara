import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { IntegrationEventsSection } from "../integration-events-section";

const STAMP = "2026-01-05T00:00:00Z";

const integrationEvents = [
  {
    id: 1,
    topic: "sale.order.created",
    payload: {},
    status: "sent",
    retries: 0,
    created_at: STAMP,
    updated_at: STAMP,
  },
  {
    id: 2,
    topic: "invoice.paid",
    payload: {},
    status: "failed",
    retries: 3,
    created_at: STAMP,
    updated_at: STAMP,
  },
];

beforeEach(() => {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/integration-events", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { integration_events: integrationEvents },
      }),
    ),
  );
});

describe("IntegrationEventsSection", () => {
  it("renders seeded integration events", async () => {
    renderWithProviders(<IntegrationEventsSection orgId="1" />);

    expect(await screen.findByText("sale.order.created")).toBeInTheDocument();
    expect(screen.getByText("invoice.paid")).toBeInTheDocument();
  });

  it("filters events by search", async () => {
    const user = userEvent.setup();
    renderWithProviders(<IntegrationEventsSection orgId="1" />);

    await screen.findByText("sale.order.created");
    await user.type(screen.getByPlaceholderText("Cari event…"), "invoice");

    expect((await screen.findAllByText("invoice.paid")).length).toBeGreaterThan(0);
    await waitFor(() => expect(screen.queryByText("sale.order.created")).not.toBeInTheDocument());
  });
});
