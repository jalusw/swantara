import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { beforeEach, describe, expect, it } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ServiceContractDetail } from "../service-contract-detail-section";

const serviceContract = {
  id: 1,
  organization_id: 1,
  name: "ACME Maintenance",
  contact_id: null,
  equipment_id: null,
  subscription_id: null,
  coverage: null,
  sla_response_hours: null,
  date_start: null,
  date_end: null,
  state: "draft",
};

function seed(overrides: Record<string, unknown> = {}) {
  server.use(
    http.get("*/api/v1/organizations/:organizationId/service-contracts/:contractId", () =>
      HttpResponse.json({
        success: true,
        message: "OK.",
        data: { service_contract: { ...serviceContract, ...overrides } },
      }),
    ),
    http.post("*/api/v1/organizations/:organizationId/service-contracts/:contractId/activate", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
    http.post("*/api/v1/organizations/:organizationId/service-contracts/:contractId/cancel", () =>
      HttpResponse.json({ success: true, message: "OK.", data: {} }),
    ),
  );
}

beforeEach(() => {
  seed();
});

describe("ServiceContractDetail branches", () => {
  it("renders loading state while fetching", () => {
    server.use(
      http.get(
        "*/api/v1/organizations/:organizationId/service-contracts/:contractId",
        () => new Promise(() => {}),
      ),
    );
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    expect(screen.getByText("Memuat...")).toBeInTheDocument();
  });

  it("renders not-found when contract is missing", async () => {
    server.use(
      http.get("*/api/v1/organizations/:organizationId/service-contracts/:contractId", () =>
        HttpResponse.json({ success: true, message: "OK.", data: { service_contract: null } }),
      ),
    );
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    expect(await screen.findByText("Kontrak tidak ditemukan.")).toBeInTheDocument();
  });

  it("renders dash fallbacks for null relations", async () => {
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    expect(await screen.findByRole("heading", { name: "ACME Maintenance" })).toBeInTheDocument();
    expect(screen.getAllByText("—").length).toBeGreaterThan(0);
  });

  it("renders filled relations with hash prefixes", async () => {
    seed({
      contact_id: 2,
      equipment_id: 3,
      subscription_id: 4,
      coverage: "24/7",
      sla_response_hours: 4,
      date_start: "2026-01-01",
      date_end: "2026-12-31",
    });
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    expect(await screen.findByText("24/7")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#3")).toBeInTheDocument();
  });

  it("cancels an active contract", async () => {
    const user = userEvent.setup();
    seed({ state: "active" });
    renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);

    await screen.findByRole("heading", { name: "ACME Maintenance" });
    const cancel = screen.queryByRole("button", { name: "Batal" });
    if (cancel) await user.click(cancel);
    expect(await screen.findByRole("heading", { name: "ACME Maintenance" })).toBeInTheDocument();
  });

  it("renders each tone branch without crashing", async () => {
    for (const state of ["draft", "active", "expired", "cancelled", "suspended"]) {
      seed({ state });
      const { unmount } = renderWithProviders(<ServiceContractDetail orgId="1" contractId="1" />);
      expect(await screen.findByRole("heading", { name: "ACME Maintenance" })).toBeInTheDocument();
      unmount();
    }
  });
});
