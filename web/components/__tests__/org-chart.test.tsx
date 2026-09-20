import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { OrgChart, type OrgChartNode } from "@/components/org-chart";
import { renderWithProviders } from "@/lib/tests";

const TREE: OrgChartNode = {
  id: "ceo",
  label: "CEO",
  description: "Chief executive",
  children: [
    {
      id: "cfo",
      label: "CFO",
      children: [
        { id: "accounting", label: "Accounting" },
        { id: "finance", label: "Finance" },
      ],
    },
    {
      id: "cto",
      label: "CTO",
      children: [{ id: "engineering", label: "Engineering" }],
    },
  ],
};

describe("OrgChart", () => {
  it("renders the tree with nodes and descriptions", () => {
    renderWithProviders(<OrgChart root={TREE} />);

    expect(screen.getByText("CEO")).toBeInTheDocument();
    expect(screen.getByText("Chief executive")).toBeInTheDocument();
    expect(screen.getByText("CFO")).toBeInTheDocument();
    expect(screen.getByText("Accounting")).toBeInTheDocument();
    expect(screen.getByText("Engineering")).toBeInTheDocument();
  });

  it("collapses and re-expands a subtree from its toggle", async () => {
    const user = userEvent.setup();
    renderWithProviders(<OrgChart root={TREE} />);

    await user.click(screen.getByRole("button", { name: "Collapse CFO" }));
    expect(screen.queryByText("Accounting")).not.toBeInTheDocument();
    expect(screen.queryByText("Finance")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Expand CFO" }));
    expect(screen.getByText("Accounting")).toBeInTheDocument();
  });

  it("starts with defaultCollapsedIds subtrees collapsed", () => {
    renderWithProviders(<OrgChart root={TREE} defaultCollapsedIds={["cto"]} />);

    expect(screen.getByRole("button", { name: "Expand CTO" })).toBeInTheDocument();
    expect(screen.queryByText("Engineering")).not.toBeInTheDocument();
  });

  it("reports the selected node through onSelect", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    renderWithProviders(<OrgChart root={TREE} onSelect={onSelect} />);

    await user.click(screen.getByRole("button", { name: "CFO" }));
    expect(onSelect).toHaveBeenCalledWith(expect.objectContaining({ id: "cfo", label: "CFO" }));
  });

  it("exposes expand/collapse state on the toggle button", () => {
    renderWithProviders(<OrgChart root={TREE} />);

    expect(screen.getByRole("button", { name: "Collapse CFO" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
    expect(screen.getByRole("button", { name: "Collapse CTO" })).toHaveAttribute(
      "aria-expanded",
      "true",
    );
  });
});
