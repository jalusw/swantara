import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/accordion";
import { renderWithProviders } from "@/lib/tests";

describe("Accordion", () => {
  it("renders the accordion container with the data-slot", () => {
    const { container } = renderWithProviders(
      <Accordion>
        <AccordionItem value="item-1">
          <AccordionTrigger>Section 1</AccordionTrigger>
          <AccordionContent>Content 1</AccordionContent>
        </AccordionItem>
      </Accordion>,
    );

    expect(container.querySelector('[data-slot="accordion"]')).toBeInTheDocument();
  });

  it("renders accordion items with triggers and content", () => {
    renderWithProviders(
      <Accordion>
        <AccordionItem value="a">
          <AccordionTrigger>First</AccordionTrigger>
          <AccordionContent>Body A</AccordionContent>
        </AccordionItem>
        <AccordionItem value="b">
          <AccordionTrigger>Second</AccordionTrigger>
          <AccordionContent>Body B</AccordionContent>
        </AccordionItem>
      </Accordion>,
    );

    expect(screen.getByText("First")).toBeInTheDocument();
    expect(screen.getByText("Second")).toBeInTheDocument();
  });

  it("applies a custom className to the accordion", () => {
    const { container } = renderWithProviders(
      <Accordion className="my-accordion">
        <AccordionItem value="x">
          <AccordionTrigger>Title</AccordionTrigger>
          <AccordionContent>Details</AccordionContent>
        </AccordionItem>
      </Accordion>,
    );

    expect(container.querySelector('[data-slot="accordion"]')).toHaveClass("my-accordion");
  });
});
