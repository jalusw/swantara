import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Alert, AlertDescription, AlertTitle } from "@/components/alert";
import { renderWithProviders } from "@/lib/tests";

describe("Alert", () => {
  it("renders an alert with role and data-slot", () => {
    renderWithProviders(
      <Alert>
        <AlertTitle>Heads up</AlertTitle>
        <AlertDescription>You have new mail.</AlertDescription>
      </Alert>,
    );

    expect(screen.getByRole("alert")).toHaveAttribute("data-slot", "alert");
    expect(screen.getByText("Heads up")).toBeInTheDocument();
    expect(screen.getByText("You have new mail.")).toBeInTheDocument();
  });

  it("applies the destructive variant", () => {
    const { container } = renderWithProviders(
      <Alert variant="destructive">
        <AlertTitle>Error</AlertTitle>
      </Alert>,
    );

    expect(container.firstElementChild).toHaveClass("text-destructive");
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(<Alert className="my-alert" />);

    expect(container.firstElementChild).toHaveClass("my-alert");
  });
});
