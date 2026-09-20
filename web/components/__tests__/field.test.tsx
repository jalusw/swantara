import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Field } from "@/components/field";
import { renderWithProviders } from "@/lib/tests";

describe("Field", () => {
  it("renders with data-slot", () => {
    const { container } = renderWithProviders(
      <Field>
        <label>Name</label>
      </Field>,
    );

    expect(container.firstElementChild).toHaveAttribute("data-slot", "field");
  });

  it("renders children", () => {
    renderWithProviders(
      <Field>
        <label>Email</label>
      </Field>,
    );

    expect(screen.getByText("Email")).toBeInTheDocument();
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(<Field className="my-field" />);

    expect(container.firstElementChild).toHaveClass("my-field");
  });
});
