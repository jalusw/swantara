import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";
import { renderWithProviders } from "@/lib/tests";

describe("Tabs", () => {
  it("renders tabs with data-slot", () => {
    renderWithProviders(
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">Content 1</TabsContent>
      </Tabs>,
    );

    expect(screen.getByText("Tab 1")).toHaveAttribute("data-slot", "tabs-trigger");
    expect(screen.getByText("Content 1")).toHaveAttribute("data-slot", "tabs-content");
  });

  it("renders list with data-slot", () => {
    const { container } = renderWithProviders(
      <Tabs defaultValue="tab1">
        <TabsList data-slot="tabs-list">
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>,
    );

    expect(container.querySelector("[data-slot='tabs-list']")).toBeInTheDocument();
  });

  it("passes through a custom className on tabs", () => {
    renderWithProviders(
      <Tabs defaultValue="tab1" className="my-tabs">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
        </TabsList>
      </Tabs>,
    );

    expect(screen.getByRole("tablist").parentElement).toHaveClass("my-tabs");
  });
});
