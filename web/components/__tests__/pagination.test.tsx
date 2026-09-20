import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
} from "@/components/pagination";
import { renderWithProviders } from "@/lib/tests";

describe("Pagination", () => {
  it("renders a nav with aria-label", () => {
    renderWithProviders(
      <Pagination>
        <PaginationContent />
      </Pagination>,
    );

    expect(screen.getByRole("navigation", { name: "pagination" })).toHaveAttribute(
      "data-slot",
      "pagination",
    );
  });

  it("renders pagination items", () => {
    renderWithProviders(
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationLink isActive>1</PaginationLink>
          </PaginationItem>
          <PaginationItem>
            <PaginationLink>2</PaginationLink>
          </PaginationItem>
        </PaginationContent>
      </Pagination>,
    );

    expect(screen.getByText("1")).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("2")).not.toHaveAttribute("aria-current");
  });

  it("renders link with data-slot", () => {
    renderWithProviders(
      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationLink>3</PaginationLink>
          </PaginationItem>
        </PaginationContent>
      </Pagination>,
    );

    expect(screen.getByText("3").closest("[data-slot='pagination-link']")).toBeInTheDocument();
  });
});
