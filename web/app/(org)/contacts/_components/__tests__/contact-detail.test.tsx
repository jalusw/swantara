import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactDetail } from "../contact-detail-section";

vi.mock("@/lib/hooks/use-contact-query", () => ({
  useContactQuery: vi.fn(() => ({
    data: {
      id: 1,
      name: "Bluebird Trading",
      displayName: "Bluebird",
      isOrganization: true,
      email: "info@bluebird.com",
      phone: "+65 6234 5678",
      mobile: null,
      website: null,
      taxId: "T123456",
      industry: "Trading",
      currencyCode: "SGD",
      lang: "en",
      active: true,
      organizationId: 1,
      parentId: null,
      createdAt: new Date(),
      updatedAt: new Date(),
    },
    isLoading: false,
    refetch: vi.fn(),
  })),
  useContactAddressesQuery: vi.fn(() => ({ data: [], isLoading: false })),
  useContactBankAccountsQuery: vi.fn(() => ({ data: [], isLoading: false })),
}));

describe("ContactDetail", () => {
  it("renders the record layout with all tabs", () => {
    renderWithProviders(<ContactDetail orgId="1" contactId="1" />);

    expect(screen.getAllByText("Bluebird").length).toBeGreaterThan(0);
    expect(screen.getByRole("tab", { name: "Overview" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Addresses" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Bank accounts" })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: "Defaults" })).toBeInTheDocument();
  });
});
