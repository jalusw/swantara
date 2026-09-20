import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  Carousel,
  CarouselContent,
  CarouselItem,
  CarouselNext,
  CarouselPrevious,
  useCarousel,
} from "@/components/carousel";
import { renderWithProviders } from "@/lib/tests";

const scrollPrev = vi.fn();
const scrollNext = vi.fn();
let canPrev = false;
let canNext = false;
const setApiCalls: unknown[] = [];

vi.mock("embla-carousel-react", () => ({
  default: () => [
    vi.fn(),
    {
      canScrollPrev: () => canPrev,
      canScrollNext: () => canNext,
      scrollPrev,
      scrollNext,
      on: vi.fn(),
      off: vi.fn(),
    },
  ],
}));

function renderCarousel(orientation: "horizontal" | "vertical" = "horizontal") {
  return renderWithProviders(
    <Carousel orientation={orientation} setApi={(api) => setApiCalls.push(api)}>
      <CarouselContent>
        <CarouselItem>Slide 1</CarouselItem>
        <CarouselItem>Slide 2</CarouselItem>
      </CarouselContent>
      <CarouselPrevious />
      <CarouselNext />
    </Carousel>,
  );
}

beforeEach(() => {
  scrollPrev.mockClear();
  scrollNext.mockClear();
  setApiCalls.length = 0;
  canPrev = false;
  canNext = false;
});

describe("Carousel extra2", () => {
  it("disables navigation buttons when scrolling is unavailable", () => {
    renderCarousel();

    expect(screen.getByRole("button", { name: "Previous slide" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next slide" })).toBeDisabled();
  });

  it("enables navigation buttons when scrolling is available", () => {
    canPrev = true;
    canNext = true;
    renderCarousel();

    expect(screen.getByRole("button", { name: "Previous slide" })).toBeEnabled();
    expect(screen.getByRole("button", { name: "Next slide" })).toBeEnabled();
  });

  it("scrolls on navigation button clicks", async () => {
    canPrev = true;
    canNext = true;
    const user = userEvent.setup();
    renderCarousel();

    await user.click(screen.getByRole("button", { name: "Previous slide" }));
    await user.click(screen.getByRole("button", { name: "Next slide" }));

    expect(scrollPrev).toHaveBeenCalledTimes(1);
    expect(scrollNext).toHaveBeenCalledTimes(1);
  });

  it("publishes the api through setApi", () => {
    renderCarousel();

    expect(setApiCalls).toHaveLength(1);
  });

  it("scrolls with arrow keys", async () => {
    canPrev = true;
    canNext = true;
    renderCarousel();

    const region = screen.getByRole("region");
    fireEvent.keyDown(region, { key: "ArrowLeft" });
    fireEvent.keyDown(region, { key: "ArrowRight" });

    expect(scrollPrev).toHaveBeenCalledTimes(1);
    expect(scrollNext).toHaveBeenCalledTimes(1);
  });

  it("throws when useCarousel is used outside the provider", () => {
    function Rogue() {
      useCarousel();
      return null;
    }

    expect(() => renderWithProviders(<Rogue />)).toThrow(
      "useCarousel must be used within a <Carousel />",
    );
  });
});
