import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import AuthShell from "../_components/auth-shell";

describe("AuthShell", () => {
  it("renders the title, subtitle, and children inside a card", async () => {
    const element = await AuthShell({
      title: "Welcome back",
      subtitle: "Enter your credentials",
      children: <p>inside</p>,
    });
    const { container } = render(element);

    expect(screen.getByRole("heading", { level: 1, name: "Welcome back" })).toBeInTheDocument();
    expect(screen.getByText("Enter your credentials")).toBeInTheDocument();
    expect(screen.getByText("inside")).toBeInTheDocument();
    expect(container.querySelector('[data-slot="auth-shell"]')).toBeInTheDocument();
    expect(container.querySelector('[data-slot="card"]')).toBeInTheDocument();
  });

  it("renders an empty subtitle when not provided", async () => {
    const element = await AuthShell({
      title: "Set a new password",
      children: <p>inside</p>,
    });
    render(element);

    expect(
      screen.getByRole("heading", { level: 1, name: "Set a new password" }),
    ).toBeInTheDocument();
    expect(screen.getByText("inside")).toBeInTheDocument();
  });

  it("renders a single centered column", async () => {
    const element = await AuthShell({
      title: "Welcome back",
      subtitle: "Enter your credentials",
      children: <p>inside</p>,
    });
    const { container } = render(element);

    const wrapper = container.querySelector('[data-slot="auth-shell"]') as HTMLElement;
    expect(wrapper.className).toContain("min-h-screen");
    expect(wrapper.className).toContain("bg-muted/20");
    expect(wrapper.className).not.toContain("lg:grid-cols-2");
  });
});
