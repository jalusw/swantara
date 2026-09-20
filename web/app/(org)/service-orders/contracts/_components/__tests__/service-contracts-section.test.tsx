import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceContractsSection } from "../service-contracts-section";

const serviceContracts = [
  {
    id: 1,
    organization_id: 1,
    name: "ACME Maintenance",
    contact_id: 1,
    equipment_id: null,
    subscription_id: null,
    coverage: "24/7 support",
    sla_response_hours: 4,
    date_start: "2026-01-01",
    date_end: "2026-12-31",
    state: "draft",
  },
  {
    id: 2,
    organization_id: 1,
    name: "Beta Support",
    contact_id: 2,
    equipment_id: null,
    subscription_id: null,
    coverage: "Business hours",
    sla_response_hours: 8,
    date_start: "2026-02-01",
    date_end: "2026-12-31",
    state: "active",
  },
];

function useServiceContractHandlers() {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-contracts", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_contracts: serviceContracts },
      }),
    ),
  );
}

beforeEach(() => {
  useServiceContractHandlers();
});

describe("ServiceContractsSection", () => {
  it("renders contracts with coverage and state badges", async () => {
    renderWithProviders(<ServiceContractsSection orgId="1" />);

    expect(await screen.findByText("ACME Maintenance")).toBeInTheDocument();
    expect(screen.getByText("Beta Support")).toBeInTheDocument();
    expect(screen.getByText("24/7 support")).toBeInTheDocument();
    expect(screen.getByText("Draft")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
  });

  it("filters contracts through the search box", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ServiceContractsSection orgId="1" />);

    await screen.findByText("ACME Maintenance");
    await user.type(screen.getByPlaceholderText(/Search contracts/), "Beta");

    expect(await screen.findByText("Beta Support")).toBeInTheDocument();
    await waitFor(() => expect(screen.queryByText("ACME Maintenance")).not.toBeInTheDocument());
  });
});
