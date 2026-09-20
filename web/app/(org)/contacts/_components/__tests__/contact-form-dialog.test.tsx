import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderWithProviders } from "@/lib/tests";
import { ContactFormDialog } from "../contact-form-dialog";

describe("ContactFormDialog", () => {
  it("renders create form when no initial contact", () => {
    renderWithProviders(
      <ContactFormDialog open={true} onOpenChange={vi.fn()} orgId="1" onSave={vi.fn()} />,
    );

    expect(screen.getByText("New contact")).toBeInTheDocument();
  });

  it("renders edit form when initial contact provided", () => {
    renderWithProviders(
      <ContactFormDialog
        open={true}
        onOpenChange={vi.fn()}
        orgId="1"
        initial={
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
        onSave={vi.fn()}
      />,
    );

    expect(screen.getByText("Edit contact")).toBeInTheDocument();
  });
});
