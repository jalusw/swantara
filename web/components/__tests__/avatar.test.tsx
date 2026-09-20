import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Avatar, AvatarFallback } from "@/components/avatar";
import { renderWithProviders } from "@/lib/tests";

describe("Avatar", () => {
  it("renders the avatar container with data-slot", () => {
    const { container } = renderWithProviders(
      <Avatar>
        <AvatarFallback>JD</AvatarFallback>
      </Avatar>,
    );

    expect(container.querySelector('[data-slot="avatar"]')).toBeInTheDocument();
  });

  it("renders the fallback text", () => {
    renderWithProviders(
      <Avatar>
        <AvatarFallback>AB</AvatarFallback>
      </Avatar>,
    );

    expect(screen.getByText("AB")).toBeInTheDocument();
  });

  it("applies size variants via data-size", () => {
    const { container } = renderWithProviders(
      <Avatar size="lg">
        <AvatarFallback>XL</AvatarFallback>
      </Avatar>,
    );

    expect(container.querySelector('[data-slot="avatar"]')).toHaveAttribute("data-size", "lg");
  });
});
