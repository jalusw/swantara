import { describe, expect, it } from "vitest";
import { Container } from "@/components/container";
import { renderWithProviders } from "@/lib/tests";

describe("Container", () => {
  it("renders children inside a div with data-slot", () => {
    const { container } = renderWithProviders(
      <Container>
        <span>Content</span>
      </Container>,
    );

    const el = container.querySelector('[data-slot="container"]');
    expect(el).toBeInTheDocument();
    expect(el).toHaveTextContent("Content");
  });

  it("applies the default layout classes", () => {
    const { container } = renderWithProviders(<Container />);

    expect(container.firstElementChild).toHaveClass("mx-auto", "w-full", "max-w-screen-xl");
  });

  it("passes through a custom className", () => {
    const { container } = renderWithProviders(<Container className="my-container" />);

    expect(container.firstElementChild).toHaveClass("my-container");
  });
});
