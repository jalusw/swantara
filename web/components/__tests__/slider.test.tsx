import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Slider } from "@/components/slider";
import { renderWithProviders } from "@/lib/tests";

describe("Slider", () => {
  it("renders with the slider slot", () => {
    const { container } = renderWithProviders(<Slider defaultValue={[30]} />);

    expect(container.querySelector("[data-slot='slider']")).not.toBeNull();
    expect(container.querySelector("[data-slot='slider-thumb']")).not.toBeNull();
  });

  it("renders a thumb per value", () => {
    const { container } = renderWithProviders(<Slider defaultValue={[20, 80]} />);

    expect(container.querySelectorAll("[data-slot='slider-thumb']")).toHaveLength(2);
  });

  it("applies a custom className", () => {
    const { container } = renderWithProviders(<Slider defaultValue={[10]} className="my-slider" />);

    expect(container.querySelector("[data-slot='slider']")).toHaveClass("my-slider");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
