import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";

import { renderWithProviders } from "@/lib/tests";
import { TextClamp } from "../text-clamp";

function renderOverflowing() {
  const view = renderWithProviders(<TextClamp lines={2}>short</TextClamp>);
  const element = view.container.querySelector('[data-slot="text-clamp-content"]') as HTMLElement;
  Object.defineProperty(element, "scrollHeight", {
    configurable: true,
    value: 200,
  });
  Object.defineProperty(element, "clientHeight", {
    configurable: true,
    value: 60,
  });
  view.rerender(<TextClamp lines={2}>{"a long description ".repeat(50)}</TextClamp>);
  return view;
}

describe("TextClamp", () => {
  it("shows a show-more toggle when the content overflows", () => {
    renderOverflowing();

    expect(screen.getByRole("button", { name: "Show more" })).toBeInTheDocument();
  });

  it("expands and collapses the text", async () => {
    const user = userEvent.setup();
    const view = renderOverflowing();
    const element = view.container.querySelector('[data-slot="text-clamp-content"]') as HTMLElement;

    await user.click(screen.getByRole("button", { name: "Show more" }));

    expect(screen.getByRole("button", { name: "Show less" })).toBeInTheDocument();
    expect(element.style.webkitLineClamp).toBe("");

    await user.click(screen.getByRole("button", { name: "Show less" }));

    expect(screen.getByRole("button", { name: "Show more" })).toBeInTheDocument();
  });

  it("hides the toggle when the content fits", () => {
    const view = renderWithProviders(<TextClamp lines={3}>short</TextClamp>);
    const element = view.container.querySelector('[data-slot="text-clamp-content"]') as HTMLElement;
    Object.defineProperty(element, "scrollHeight", {
      configurable: true,
      value: 40,
    });
    Object.defineProperty(element, "clientHeight", {
      configurable: true,
      value: 60,
    });
    view.rerender(<TextClamp lines={3}>still short</TextClamp>);

    expect(screen.queryByRole("button", { name: /show more/i })).not.toBeInTheDocument();
  });

  it("supports custom toggle labels", () => {
    const view = renderWithProviders(<TextClamp lines={1}>x</TextClamp>);
    const element = view.container.querySelector('[data-slot="text-clamp-content"]') as HTMLElement;
    Object.defineProperty(element, "scrollHeight", {
      configurable: true,
      value: 100,
    });
    Object.defineProperty(element, "clientHeight", {
      configurable: true,
      value: 20,
    });
    view.rerender(
      <TextClamp lines={1} showMoreLabel="Read more" showLessLabel="Read less">
        long text
      </TextClamp>,
    );

    expect(screen.getByRole("button", { name: "Read more" })).toBeInTheDocument();
  });
});
