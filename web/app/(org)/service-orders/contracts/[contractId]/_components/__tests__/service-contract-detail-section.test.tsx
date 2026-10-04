import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceContractDetail } from "../service-contract-detail-section";

const serviceContract = {
  id: 1,
  organization_id: 1,
  name: "ACME Maintenance",
  contact_id: 2,
  equipment_id: null,
  subscription_id: null,
  coverage: "24/7 support",
  sla_response_hours: 4,
  date_start: "2026-01-01",
  date_end: "2026-12-31",
  state: "draft",
};

let activateCalled = false;

function useServiceContractDetailHandlers() {
  activateCalled = false;
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-contracts/:contractId", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_contract: serviceContract },
      }),
    ),
    http.post(
      "*/api/v1/organizations/:organizationId/service-contracts/:contractId/activate",
      () => {
        activateCalled = true;
        return HttpResponse.json({
          success: true,
          message: "OK.",
          data: { service_contract: { ...serviceContract, state: "active" } },
        });
      },
    ),
  );
}

beforeEach(() => {
  useServiceContractDetailHandlers();
});

describe("ServiceContractDetail", () => {
  it("renders contract name with state badge", async () => {
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    expect(await screen.findByRole("heading", { name: "ACME Maintenance" })).toBeInTheDocument();
    expect(screen.getByText("24/7 support")).toBeInTheDocument();
    expect(screen.getByText("Draf")).toBeInTheDocument();
  });

  it("activates the contract from the action button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    await screen.findByRole("heading", { name: "ACME Maintenance" });
    await user.click(screen.getByRole("button", { name: "Aktifkan" }));

    await waitFor(() => expect(activateCalled).toBe(true));
  });
});
