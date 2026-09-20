import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Carousel, CarouselContent, CarouselItem } from "@/components/carousel";
import { renderWithProviders } from "@/lib/tests";

vi.mock("embla-carousel-react", () => ({
  default: () => [vi.fn(), undefined],
}));

describe("Carousel", () => {
  it("renders with carousel role and roledescription", () => {
    renderWithProviders(
      <Carousel>
        <CarouselContent>
          <CarouselItem>Slide 1</CarouselItem>
          <CarouselItem>Slide 2</CarouselItem>
        </CarouselContent>
      </Carousel>,
    );

    const region = screen.getByRole("region");
    expect(region).toHaveAttribute("aria-roledescription", "carousel");
    expect(region).toHaveAttribute("data-slot", "carousel");
  });

  it("renders slides with group role", () => {
    renderWithProviders(
      <Carousel>
        <CarouselContent>
          <CarouselItem>First</CarouselItem>
          <CarouselItem>Second</CarouselItem>
        </CarouselContent>
      </Carousel>,
    );

    const slides = screen.getAllByRole("group");
    expect(slides).toHaveLength(2);
    for (const slide of slides) {
      expect(slide).toHaveAttribute("aria-roledescription", "slide");
    }
  });

  it("applies a custom className", () => {
    renderWithProviders(
      <Carousel className="my-carousel">
        <CarouselContent>
          <CarouselItem>Slide</CarouselItem>
        </CarouselContent>
      </Carousel>,
    );

    expect(screen.getByRole("region")).toHaveClass("my-carousel");
  });
});
