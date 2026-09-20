import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactDefaults } from "../contact-defaults";

describe("ContactDefaults", () => {
  it("renders customer and supplier panels", () => {
    renderWithProviders(
      <ContactDefaults
        orgId="1"
        contactId="1"
        contact={
          {
            id: 1,
            name: "Test Contact",
            displayName: "Test",
            isOrganization: true,
            email: null,
            phone: null,
            mobile: null,
            website: null,
            taxId: null,
            industry: null,
            currencyCode: null,
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

    expect(screen.getByText("Customer")).toBeInTheDocument();
    expect(screen.getByText("Supplier")).toBeInTheDocument();
  });
});
