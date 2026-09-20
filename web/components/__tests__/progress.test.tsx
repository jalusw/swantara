import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Progress } from "@/components/progress";
import { renderWithProviders } from "@/lib/tests";

describe("Progress", () => {
  it("renders with data-slot", () => {
    renderWithProviders(<Progress value={50} />);

    expect(screen.getByRole("progressbar")).toHaveAttribute("data-slot", "progress");
  });

  it("renders track and indicator", () => {
    const { container } = renderWithProviders(<Progress value={50} />);

    expect(container.querySelector("[data-slot='progress-track']")).toBeInTheDocument();
    expect(container.querySelector("[data-slot='progress-indicator']")).toBeInTheDocument();
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Progress value={50} className="my-progress" />);

    expect(screen.getByRole("progressbar")).toHaveClass("my-progress");
  });
});
