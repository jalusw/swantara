import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { PageHeader } from "@/components/page-header";
import { renderWithProviders } from "@/lib/tests";

describe("PageHeader", () => {
  it("renders the title and description", () => {
    renderWithProviders(<PageHeader title="Dashboard" description="Welcome back to Acme Inc." />);

    expect(screen.getByRole("heading", { name: "Dashboard", level: 1 })).toBeInTheDocument();
    expect(screen.getByText("Welcome back to Acme Inc.")).toBeInTheDocument();
  });

  it("renders the actions slot", () => {
    renderWithProviders(
      <PageHeader title="Dashboard" actions={<button type="button">Export</button>} />,
    );

    expect(screen.getByRole("button", { name: "Export" })).toBeInTheDocument();
  });
});
