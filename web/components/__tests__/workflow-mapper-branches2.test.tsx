import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { WorkflowEdge, WorkflowNode } from "@/components/workflow-mapper";
import { WorkflowMapper } from "@/components/workflow-mapper";
import { renderWithProviders } from "@/lib/tests";

const NODES: WorkflowNode[] = [
  { id: "a", def: "draft", label: "Draft", x: 10, y: 10 },
  { id: "b", def: "review", label: "Review", x: 220, y: 10 },
];

const EDGES: WorkflowEdge[] = [{ id: "a-b", sourceId: "a", targetId: "b" }];

const AVAILABLE = [
  { type: "draft", label: "Draft", tone: "primary" as const },
  { type: "review", label: "Review", tone: "success" as const },
];

describe("WorkflowMapper branches2", () => {
  it("renders nodes with fallback tone when def is unknown", () => {
    renderWithProviders(
      <WorkflowMapper
        nodes={[{ id: "x", def: "unknown", label: "Mystery", x: 0, y: 0 }]}
        edges={[]}
      />,
    );

    expect(screen.getByLabelText("Mystery")).toBeInTheDocument();
  });

  it("renders rtl port positions without crashing", () => {
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} direction="rtl" />);

    expect(screen.getByLabelText("Draft")).toBeInTheDocument();
    expect(screen.getByLabelText("Review")).toBeInTheDocument();
  });

  it("skips edges with missing endpoints", () => {
    renderWithProviders(
      <WorkflowMapper
        nodes={NODES}
        edges={[{ id: "ghost", sourceId: "a", targetId: "missing" }]}
      />,
    );

    expect(screen.getByLabelText("Draft")).toBeInTheDocument();
  });

  it("adds a node through the available palette", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <WorkflowMapper nodes={NODES} edges={[]} available={AVAILABLE} onChange={onChange} />,
    );

    await user.click(screen.getByRole("button", { name: "Draft" }));
    expect(onChange).toHaveBeenCalled();
  });

  it("disables palette buttons without onChange and zooms", async () => {
    const user = userEvent.setup();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={[]} available={AVAILABLE} />);

    expect(screen.getByRole("button", { name: "Draft" })).toBeDisabled();
    expect(screen.getByText("100%")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Zoom in" }));
    expect(await screen.findByText("125%")).toBeInTheDocument();
  });

  it("deletes a node through its delete button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Delete Draft" }));
    expect(onChange).toHaveBeenCalledWith(
      expect.arrayContaining([expect.objectContaining({ id: "b" })]),
      [],
    );
  });
});
