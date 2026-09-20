import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { FieldFeedback } from "@/components/field-feedback";
import { renderWithProviders } from "@/lib/tests";

describe("FieldFeedback", () => {
  it("renders nothing when not visible", () => {
    const { container } = renderWithProviders(
      <FieldFeedback visible={false}>Error message</FieldFeedback>,
    );

    expect(container.querySelector("[data-slot='field-feedback']")).not.toBeInTheDocument();
  });

  it("renders visible feedback with intent", () => {
    renderWithProviders(
      <FieldFeedback visible intent="danger">
        Required field
      </FieldFeedback>,
    );

    expect(screen.getByText("Required field")).toHaveAttribute("data-slot", "field-feedback");
    expect(screen.getByText("Required field")).toHaveClass("text-destructive");
  });

  it("applies success intent class", () => {
    renderWithProviders(
      <FieldFeedback visible intent="success">
        Looks good
      </FieldFeedback>,
    );

    expect(screen.getByText("Looks good")).toHaveClass("text-success");
  });

  describe("regression: danger intent uses semantic destructive color", () => {
    it("should not use non-existent text-danger class", () => {
      renderWithProviders(
        <FieldFeedback visible intent="danger">
          Required field
        </FieldFeedback>,
      );

      const feedback = screen.getByText("Required field");
      expect(feedback).toHaveClass("text-destructive");
      expect(feedback).not.toHaveClass("text-danger");
    });

    it("should map md size to existing text-base utility", () => {
      renderWithProviders(
        <FieldFeedback visible intent="danger" size="md">
          Required field
        </FieldFeedback>,
      );

      expect(screen.getByText("Required field")).toHaveClass("text-base");
    });
  });
});
