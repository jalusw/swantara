import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactOverview } from "../contact-overview";

describe("ContactOverview", () => {
  it("renders contact identity details", () => {
    renderWithProviders(
      <ContactOverview
        orgId="1"
        contactId="1"
        contact={
          {
            id: 1,
            name: "Test Contact",
            displayName: "Test",
            isOrganization: true,
            email: "test@example.com",
            phone: "+1 234 5678",
            mobile: null,
            website: null,
            taxId: "T123",
            industry: "Tech",
            currencyCode: "USD",
            lang: "en",
            active: true,
            organizationId: 1,
            parentId: null,
            createdAt: new Date(),
            updatedAt: new Date(),
          } as never
        }
        onRefetch={() => {}}
      />,
    );

    expect(screen.getByText("Test Contact")).toBeInTheDocument();
    expect(screen.getByText("test@example.com")).toBeInTheDocument();
  });
});
