import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { RecordLayout } from "@/components/record-layout";
import { renderWithProviders } from "@/lib/tests";

describe("RecordLayout", () => {
  it("renders header and switches tabs", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <RecordLayout
        breadcrumbItems={[{ label: "Invoicing", href: "/invoices" }, { label: "Invoice #1001" }]}
        title="Invoice #1001"
        status={<span>Paid</span>}
        tabs={[
          { id: "items", label: "Items", content: <p>Line items</p> },
          { id: "history", label: "History", content: <p>Audit history</p> },
        ]}
      />,
    );
    expect(screen.getByRole("heading", { name: "Invoice #1001" })).toBeInTheDocument();
    expect(screen.getByText("Paid")).toBeInTheDocument();
    expect(screen.getByText("Line items")).toBeInTheDocument();

    await user.click(screen.getByRole("tab", { name: "History" }));
    expect(screen.getByText("Audit history")).toBeInTheDocument();
  });
});
