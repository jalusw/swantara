import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Badge } from "@/components/badge";
import { Field } from "@/components/field";
import { FieldFeedback } from "@/components/field-feedback";
import { Input } from "@/components/input";
import { Label } from "@/components/label";
import { Separator } from "@/components/separator";
import { Skeleton } from "@/components/skeleton";
import { Textarea } from "@/components/textarea";
import { renderWithProviders } from "@/lib/tests";

describe("Badge", () => {
  it("renders its content with the default variant", () => {
    renderWithProviders(<Badge>New</Badge>);

    expect(screen.getByText("New")).toBeInTheDocument();
  });

  it("applies a custom variant and className", () => {
    const { container } = renderWithProviders(
      <Badge variant="destructive" className="custom-badge">
        Danger
      </Badge>,
    );

    expect(container.querySelector(".custom-badge")).toHaveTextContent("Danger");
  });

  it("supports rendering as another tag via the render prop", () => {
    const { container } = renderWithProviders(
      // biome-ignore lint/a11y/useAnchorContent: Badge injects its children into the render element at runtime; the anchor is intentionally empty here
      <Badge render={<a href="#x" />}>Link</Badge>,
    );

    expect(container.querySelector('a[href="#x"]')).toHaveTextContent("Link");
  });
});

describe("Skeleton", () => {
  it("renders a loading div with the given className", () => {
    const { container } = renderWithProviders(<Skeleton className="h-10" />);

    const skeleton = container.querySelector('[data-slot="skeleton"]');
    expect(skeleton).toHaveClass("motion-safe:animate-pulse", "h-10");
  });
});

describe("Separator", () => {
  it("renders a horizontal separator by default", () => {
    renderWithProviders(<Separator />);

    expect(screen.getByRole("separator")).toBeInTheDocument();
  });
});

describe("Textarea", () => {
  it("renders a textarea and forwards props", () => {
    renderWithProviders(<Textarea placeholder="Describe" disabled />);

    const textarea = screen.getByPlaceholderText("Describe");
    expect(textarea).toBeDisabled();
    expect(textarea).toHaveAttribute("data-slot", "textarea");
  });
});

describe("Input", () => {
  it("renders an input and forwards props", () => {
    renderWithProviders(<Input type="email" placeholder="you@example.com" aria-invalid />);

    const input = screen.getByPlaceholderText("you@example.com");
    expect(input).toHaveAttribute("type", "email");
    expect(input).toHaveAttribute("data-slot", "input");
  });
});

describe("Label", () => {
  it("renders a label bound to its field", () => {
    const { container } = renderWithProviders(<Label htmlFor="input-email">Email</Label>);

    const label = container.querySelector("label");
    expect(label).toHaveAttribute("for", "input-email");
    expect(label).toHaveTextContent("Email");
  });
});

describe("Field", () => {
  it("wraps its children in a vertical stack", () => {
    const { container } = renderWithProviders(
      <Field className="mt-4">
        <span>child</span>
      </Field>,
    );

    expect(container.firstElementChild).toHaveTextContent("child");
    expect(container.firstElementChild).toHaveClass("flex", "flex-col");
    expect(container.firstElementChild).toHaveClass("mt-4");
  });
});

describe("FieldFeedback", () => {
  it("renders the message when visible", () => {
    renderWithProviders(
      <FieldFeedback visible intent="danger">
        Required field
      </FieldFeedback>,
    );

    expect(screen.getByText("Required field")).toBeInTheDocument();
  });

  it("renders nothing when invisible", () => {
    renderWithProviders(<FieldFeedback visible={false}>Hidden</FieldFeedback>);

    expect(screen.queryByText("Hidden")).not.toBeInTheDocument();
  });

  it("renders nothing when there is no message", () => {
    renderWithProviders(<FieldFeedback visible />);

    expect(screen.queryByRole("paragraph")).not.toBeInTheDocument();
  });

  it("applies the intent and size variants", () => {
    renderWithProviders(
      <FieldFeedback visible intent="success" size="lg">
        Saved
      </FieldFeedback>,
    );

    expect(screen.getByText("Saved")).toHaveClass("text-success", "text-lg");
  });
});
