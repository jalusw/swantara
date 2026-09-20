import { fireEvent, screen } from "@testing-library/react";
import { Package } from "lucide-react";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  type WorkflowEdge,
  WorkflowMapper,
  type WorkflowNode,
  type WorkflowNodeDef,
} from "@/components/workflow-mapper";
import { renderWithProviders } from "@/lib/tests";

const AVAILABLE: WorkflowNodeDef[] = [
  { type: "trigger", label: "Trigger", tone: "info", icon: Package },
  { type: "process", label: "Process", tone: "primary" },
];

const INITIAL_NODES: WorkflowNode[] = [
  { id: "a", def: "trigger", label: "Order created", x: 120, y: 160 },
  { id: "b", def: "process", label: "Validate order", x: 480, y: 160 },
];

type Snapshot = { nodes: WorkflowNode[]; edges: WorkflowEdge[] };

function createHarness() {
  const onChange = vi.fn();
  function Harness() {
    const [state, setState] = useState<Snapshot>({ nodes: INITIAL_NODES, edges: [] });
    return (
      <WorkflowMapper
        nodes={state.nodes}
        edges={state.edges}
        available={AVAILABLE}
        onChange={(nodes, edges) => {
          onChange(nodes, edges);
          setState({ nodes, edges });
        }}
      />
    );
  }
  renderWithProviders(<Harness />);
  return { onChange };
}

function mockContainer(): Element {
  const container = document.querySelector('[aria-label="Workflow canvas"]');
  expect(container).toBeInstanceOf(HTMLElement);
  const rect = { left: 50, top: 60, width: 900, height: 600 } as DOMRect;
  vi.spyOn(container as Element, "getBoundingClientRect").mockReturnValue(rect);
  return container as Element;
}

describe("WorkflowMapper branches3", () => {
  it("hides the palette when no node types are available", () => {
    renderWithProviders(<WorkflowMapper nodes={INITIAL_NODES} edges={[]} />);

    expect(screen.queryByText("Add node:")).toBeNull();
    expect(screen.getByText("Order created")).toBeInTheDocument();
  });

  it("renders palette icons when defs provide one", () => {
    createHarness();

    expect(screen.getByRole("button", { name: "Trigger" })).toBeInTheDocument();
  });

  it("deletes the selected node with the Delete key", () => {
    const harness = createHarness();
    mockContainer();
    const node = screen.getByRole("option", { name: "Order created" });
    fireEvent.pointerDown(node, { button: 0, clientX: 170, clientY: 220 });
    fireEvent.keyDown(node, { key: "Delete" });

    const [, edges] = harness.onChange.mock.calls.at(-1) ?? [];
    expect(harness.onChange).toHaveBeenCalled();
    expect(edges).toEqual([]);
    expect(screen.queryByText("Order created")).toBeNull();
  });

  it("keeps the canvas unchanged on Escape without a popup", () => {
    createHarness();
    const container = mockContainer();

    fireEvent.keyDown(container, { key: "Escape" });

    expect(screen.getByText("Order created")).toBeInTheDocument();
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("ignores pointer moves with no active drag", () => {
    const harness = createHarness();
    const container = mockContainer();

    fireEvent.pointerMove(container, { clientX: 300, clientY: 300 });

    expect(harness.onChange).not.toHaveBeenCalled();
  });

  it("clears a pending link when dropped on empty canvas", () => {
    const harness = createHarness();
    const container = mockContainer();
    const output = screen.getByRole("button", { name: "Order created output" });

    fireEvent.pointerDown(output, { clientX: 50 + 120 + 184, clientY: 60 + 160 + 32 });
    fireEvent.pointerMove(container, { clientX: 55, clientY: 65 });
    fireEvent.pointerUp(container, { clientX: 55, clientY: 65 });

    expect(harness.onChange).not.toHaveBeenCalled();
    expect(document.querySelector('[stroke-dasharray="6 4"]')).toBeNull();
  });

  it("keeps nodes when the delete button has no change handler", () => {
    renderWithProviders(<WorkflowMapper nodes={INITIAL_NODES} edges={[]} available={AVAILABLE} />);

    fireEvent.click(screen.getByRole("button", { name: "Delete Order created" }));

    expect(screen.getByText("Order created")).toBeInTheDocument();
  });
});
