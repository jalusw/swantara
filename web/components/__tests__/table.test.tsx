import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/table";
import { renderWithProviders } from "@/lib/tests";

describe("Table", () => {
  it("renders table with data-slot", () => {
    renderWithProviders(
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow>
            <TableCell>Alice</TableCell>
          </TableRow>
        </TableBody>
      </Table>,
    );

    expect(screen.getByRole("table")).toHaveAttribute("data-slot", "table");
    expect(screen.getByText("Name")).toHaveAttribute("data-slot", "table-head");
    expect(screen.getByText("Alice")).toHaveAttribute("data-slot", "table-cell");
  });

  it("renders header and body slots", () => {
    const { container } = renderWithProviders(
      <Table>
        <TableHeader data-slot="table-header">
          <TableRow data-slot="table-row">
            <TableHead>H</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody data-slot="table-body">
          <TableRow data-slot="table-row">
            <TableCell>C</TableCell>
          </TableRow>
        </TableBody>
      </Table>,
    );

    expect(container.querySelector("[data-slot='table-header']")).toBeInTheDocument();
    expect(container.querySelector("[data-slot='table-body']")).toBeInTheDocument();
  });

  it("passes through a custom className", () => {
    renderWithProviders(<Table className="my-table" />);

    expect(screen.getByRole("table")).toHaveClass("my-table");
  });
});
