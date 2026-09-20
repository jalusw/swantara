import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/card";
import { renderWithProviders } from "@/lib/tests";

describe("Card", () => {
  it("renders all the sub-components together", () => {
    renderWithProviders(
      <Card size="sm">
        <CardHeader>
          <CardTitle>Dimensions</CardTitle>
          <CardDescription>Live metrics</CardDescription>
          <CardAction>
            <button type="button">More</button>
          </CardAction>
        </CardHeader>
        <CardContent>Body</CardContent>
        <CardFooter>Footer</CardFooter>
      </Card>,
    );

    expect(screen.getByText("Dimensions")).toBeInTheDocument();
    expect(screen.getByText("Live metrics")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "More" })).toBeInTheDocument();
    expect(screen.getByText("Body")).toBeInTheDocument();
    expect(screen.getByText("Footer")).toBeInTheDocument();
  });

  it("marks the card slot and size", () => {
    const { container } = renderWithProviders(<Card size="sm" />);

    const card = container.querySelector('[data-slot="card"]');
    expect(card).toHaveAttribute("data-size", "sm");
  });

  it("defaults to the default size", () => {
    const { container } = renderWithProviders(<Card />);

    expect(container.querySelector('[data-slot="card"]')).toHaveAttribute("data-size", "default");
  });
});
