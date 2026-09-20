import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Breadcrumb } from "@/components/breadcrumb";
import { renderWithProviders } from "@/lib/tests";

describe("Breadcrumb", () => {
  it("renders links and marks the current page", () => {
    renderWithProviders(
      <Breadcrumb
        items={[
          { label: "Home", href: "/" },
          { label: "Invoicing", href: "/invoices" },
          { label: "Invoice #1042" },
        ]}
      />,
    );

    expect(screen.getByRole("link", { name: /Home/ })).toHaveAttribute("href", "/");
    expect(screen.getByRole("link", { name: /Invoicing/ })).toHaveAttribute("href", "/invoices");
    expect(screen.getByText("Invoice #1042")).toBeInTheDocument();
    expect(screen.getByText("Invoice #1042")).toHaveAttribute("aria-current", "page");
  });
});
