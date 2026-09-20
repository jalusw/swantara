import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { KpiWall } from "@/components/kpi-wall";
import { renderWithProviders } from "@/lib/tests";

const SLIDES = [
  {
    id: "revenue",
    title: "Revenue",
    subtitle: "This month",
    content: <p>Slide revenue body</p>,
  },
  {
    id: "orders",
    title: "Orders",
    content: <p>Slide orders body</p>,
  },
];

describe("KpiWall", () => {
  it("renders the first slide and its meta", () => {
    renderWithProviders(<KpiWall slides={SLIDES} />);

    expect(screen.getByRole("heading", { name: "Revenue" })).toBeInTheDocument();
    expect(screen.getByText("Slide revenue body")).toBeInTheDocument();
  });

  it("moves to the next slide with the next button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<KpiWall slides={SLIDES} />);

    await user.click(screen.getByRole("button", { name: "Next slide" }));

    expect(screen.getByRole("heading", { name: "Orders" })).toBeInTheDocument();
    expect(screen.getByText("Slide orders body")).toBeInTheDocument();
  });

  it("navigates directly from the slide dots", async () => {
    const user = userEvent.setup();
    renderWithProviders(<KpiWall slides={SLIDES} />);

    await user.click(screen.getByRole("tab", { name: "Go to slide 2: Orders" }));

    expect(screen.getByRole("heading", { name: "Orders" })).toBeInTheDocument();
  });
});
