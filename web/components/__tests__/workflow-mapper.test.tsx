import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  NODE_WIDTH,
  type WorkflowDirection,
  type WorkflowEdge,
  WorkflowMapper,
  type WorkflowNode,
  type WorkflowNodeDef,
} from "@/components/workflow-mapper";
import { renderWithProviders } from "@/lib/tests";

const AVAILABLE: WorkflowNodeDef[] = [
  { type: "trigger", label: "Trigger", tone: "info" },
  { type: "process", label: "Process", tone: "primary" },
];

const INITIAL_NODES: WorkflowNode[] = [
  { id: "a", def: "trigger", label: "Order created", x: 120, y: 160 },
  { id: "b", def: "process", label: "Validate order", x: 480, y: 160 },
];

type Snapshot = { nodes: WorkflowNode[]; edges: WorkflowEdge[] };

function createHarness(direction?: WorkflowDirection) {
  const onChange = vi.fn();
  function Harness() {
    const [state, setState] = useState<Snapshot>({
      nodes: INITIAL_NODES,
      edges: [],
    });
    return (
      <WorkflowMapper
        nodes={state.nodes}
        edges={state.edges}
        available={AVAILABLE}
        direction={direction}
        onChange={(nodes, edges) => {
          onChange(nodes, edges);
          setState({ nodes, edges });
        }}
      />
    );
  }
  renderWithProviders(<Harness />);
  return {
    onChange,
    latest: (): Snapshot => {
      const [nodes, edges] = onChange.mock.calls.at(-1) ?? [INITIAL_NODES, []];
      return { nodes, edges };
    },
  };
}

function mockContainer(): Element {
  const container = document.querySelector('[aria-label="Workflow canvas"]');
  expect(container).toBeInstanceOf(HTMLElement);
  const rect = { left: 50, top: 60, width: 900, height: 600 } as DOMRect;
  vi.spyOn(container as Element, "getBoundingClientRect").mockReturnValue(rect);
  return container as Element;
}

function dragOutput(container: Element, sourceLabel: string, targetLabel: string) {
  const output = screen.getByRole("button", { name: sourceLabel });
  screen.getByRole("button", { name: targetLabel });
  const outputX = 50 + 120 + NODE_WIDTH;
  const outputY = 60 + 160 + 32;
  const inputX = 50 + 480;
  const inputY = 60 + 160 + 32;

  fireEvent.pointerDown(output, { clientX: outputX, clientY: outputY });
  fireEvent.pointerMove(container, {
    clientX: (inputX + outputX) / 2,
    clientY: inputY,
  });
  fireEvent.pointerUp(container, { clientX: inputX, clientY: inputY });
}

describe("WorkflowMapper", () => {
  it("renders nodes and their connection ports", () => {
    createHarness();

    expect(screen.getByText("Order created")).toBeInTheDocument();
    expect(screen.getByText("Validate order")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Order created output" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Validate order input" })).toBeInTheDocument();
  });

  it("adds a node from the palette", () => {
    const harness = createHarness();

    fireEvent.click(screen.getByRole("button", { name: "Process" }));

    const nodes = harness.latest().nodes;
    expect(nodes).toHaveLength(3);
    expect(nodes.at(-1)?.label).toBe("Process");
    expect(nodes.at(-1)?.def).toBe("process");
  });

  it("creates an edge when dragging from an output port to an input port", () => {
    const harness = createHarness();
    const container = mockContainer();

    dragOutput(container, "Order created output", "Validate order input");

    expect(harness.latest().edges).toEqual([{ id: "a-b", sourceId: "a", targetId: "b" }]);
  });

  it("rejects duplicate edges", () => {
    const harness = createHarness();
    const container = mockContainer();

    dragOutput(container, "Order created output", "Validate order input");
    dragOutput(container, "Order created output", "Validate order input");

    expect(harness.latest().edges).toHaveLength(1);
  });

  it("reverses port placement for right-to-left direction", () => {
    createHarness("rtl");

    expect(screen.getByRole("button", { name: "Order created output" })).toHaveStyle({
      left: `${INITIAL_NODES[0]!.x}px`,
    });
    expect(screen.getByRole("button", { name: "Order created input" })).toHaveStyle({
      left: `${INITIAL_NODES[0]!.x + NODE_WIDTH}px`,
    });
  });

  it("deletes the selected node and its edges with Backspace", () => {
    const harness = createHarness();
    const container = mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.keyDown(container, { key: "Backspace" });

    const latest = harness.latest();
    expect(latest.nodes.some((node) => node.id === "a")).toBe(false);
    expect(latest.edges).toEqual([]);
  });

  it("opens a popup when clicking a node and connects via the popup", () => {
    const harness = createHarness();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });

    const connectButton = screen.getByRole("button", {
      name: "Connect Order created to Validate order",
    });
    fireEvent.click(connectButton);

    expect(harness.latest().edges).toEqual([{ id: "a-b", sourceId: "a", targetId: "b" }]);
  });

  it("closes the popup when clicking outside it", () => {
    createHarness();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });

    expect(
      screen.getByRole("button", {
        name: "Connect Order created to Validate order",
      }),
    ).toBeInTheDocument();

    const container = mockContainer();
    fireEvent.pointerDown(container, {
      clientX: 120,
      clientY: 80,
    });

    expect(
      screen.queryByRole("button", {
        name: "Connect Order created to Validate order",
      }),
    ).not.toBeInTheDocument();
  });

  it("unconnects a node from the popup", () => {
    const harness = createHarness();
    const container = mockContainer();
    dragOutput(container, "Order created output", "Validate order input");

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });

    fireEvent.click(
      screen.getByRole("button", {
        name: "Disconnect Order created from Validate order",
      }),
    );

    expect(harness.latest().edges).toEqual([]);
  });

  it("renders edges with an arrow marker", () => {
    createHarness();
    const container = mockContainer();
    dragOutput(container, "Order created output", "Validate order input");

    const arrowMarker = document.querySelector("#workflow-edge-arrow");
    expect(arrowMarker).toBeInstanceOf(SVGElement);
    const visibleEdge = Array.from(document.querySelectorAll("path[marker-end]")).find(
      (path) => path.getAttribute("marker-end") === "url(#workflow-edge-arrow)",
    );
    expect(visibleEdge?.getAttribute("d")).toContain("C");
  });

  it("pans the canvas by dragging the background", () => {
    createHarness();
    const container = mockContainer();
    const world = document.querySelector(
      "[data-slot='workflow-mapper'] [style*='translate']",
    ) as Element;

    fireEvent.pointerDown(world, { button: 0, clientX: 200, clientY: 200 });
    fireEvent.pointerMove(container, { clientX: 250, clientY: 230 });
    fireEvent.pointerUp(container, { clientX: 250, clientY: 230 });

    expect(world.getAttribute("style")).toContain("translate(50px, 30px)");
  });

  it("ignores background drags with non-primary buttons", () => {
    const harness = createHarness();
    const container = mockContainer();
    const world = document.querySelector(
      "[data-slot='workflow-mapper'] [style*='translate']",
    ) as Element;

    fireEvent.pointerDown(world, { button: 1, clientX: 200, clientY: 200 });
    fireEvent.pointerMove(container, { clientX: 400, clientY: 400 });
    fireEvent.pointerUp(container, { clientX: 400, clientY: 400 });

    expect(harness.onChange).not.toHaveBeenCalled();
    expect(world.getAttribute("style")).toContain("translate(0px, 0px)");
  });

  it("moves a node when dragged", () => {
    const harness = createHarness();
    mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      button: 0,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerMove(document.querySelector('[aria-label="Workflow canvas"]') as Element, {
      clientX: 50 + 120 + 60,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(document.querySelector('[aria-label="Workflow canvas"]') as Element, {
      clientX: 50 + 120 + 60,
      clientY: 60 + 160 + 20,
    });

    expect(harness.latest().nodes.find((node) => node.id === "a")?.x).toBe(160);
  });

  it("ignores node drags and link starts with non-primary buttons", () => {
    const harness = createHarness();
    mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      button: 1,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerDown(screen.getByRole("button", { name: "Order created output" }), {
      button: 1,
      clientX: 50 + 120 + NODE_WIDTH,
      clientY: 60 + 160 + 32,
    });

    expect(harness.onChange).not.toHaveBeenCalled();
    expect(document.querySelector("path[stroke-dasharray]")).not.toBeInTheDocument();
  });

  it("ends a pointer sequence quietly when nothing is dragged", () => {
    const harness = createHarness();
    const container = mockContainer();
    fireEvent.pointerUp(container, { clientX: 100, clientY: 100 });
    expect(harness.onChange).not.toHaveBeenCalled();
  });

  it("ignores Delete with no selection", () => {
    const harness = createHarness();
    const container = mockContainer();
    fireEvent.keyDown(container, { key: "Delete" });
    expect(harness.onChange).not.toHaveBeenCalled();
  });

  it("selects an edge and deletes it with Backspace", () => {
    const harness = createHarness();
    const container = mockContainer();
    dragOutput(container, "Order created output", "Validate order input");
    expect(harness.latest().edges).toHaveLength(1);

    const widePath = document.querySelector("path.stroke-transparent") as Element;
    fireEvent.pointerDown(widePath, { button: 0 });
    expect(
      document
        .querySelector("path[marker-end='url(#workflow-edge-arrow-active)']")
        ?.getAttribute("d"),
    ).toContain("C");

    fireEvent.keyDown(container, { key: "Backspace" });
    expect(harness.latest().edges).toEqual([]);
  });

  it("opens the popup with Enter and toggles it with Space", () => {
    createHarness();
    mockContainer();
    const node = screen.getByRole("option", { name: "Order created" });

    fireEvent.keyDown(node, { key: "Enter" });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();

    fireEvent.keyDown(node, { key: " " });
    expect(
      screen.queryByRole("button", { name: "Connect Order created to Validate order" }),
    ).not.toBeInTheDocument();
  });

  it("deletes a node and its edges from its remove button", async () => {
    const user = userEvent.setup();
    const harness = createHarness();
    const container = mockContainer();
    dragOutput(container, "Order created output", "Validate order input");
    expect(harness.latest().edges).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "Delete Order created" }));

    const latest = harness.latest();
    expect(latest.nodes).toHaveLength(1);
    expect(latest.nodes[0]?.id).toBe("b");
    expect(latest.edges).toEqual([]);
  });

  it("deletes a node and its edges with Backspace", () => {
    const harness = createHarness();
    const container = mockContainer();
    dragOutput(container, "Order created output", "Validate order input");

    fireEvent.pointerDown(screen.getByText("Order created"), {
      button: 0,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.keyDown(container, { key: "Backspace" });

    const latest = harness.latest();
    expect(latest.nodes.some((node) => node.id === "a")).toBe(false);
    expect(latest.edges).toEqual([]);
  });

  it("closes the popup when its node disappears from props", () => {
    function Controlled() {
      const [nodes, setNodes] = useState(INITIAL_NODES);
      return (
        <>
          <WorkflowMapper nodes={nodes} edges={[]} available={AVAILABLE} onChange={() => {}} />
          <button type="button" onClick={() => setNodes([INITIAL_NODES[1]!])}>
            Drop first
          </button>
        </>
      );
    }
    renderWithProviders(<Controlled />);

    fireEvent.pointerDown(screen.getByText("Order created"), {
      button: 0,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      button: 0,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Drop first" }));
    expect(
      screen.queryByRole("button", { name: "Connect Order created to Validate order" }),
    ).not.toBeInTheDocument();
  });

  it("closes the popup when its node is deleted", () => {
    createHarness();
    mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Delete Order created" }));
    expect(
      screen.queryByRole("button", { name: "Connect Order created to Validate order" }),
    ).not.toBeInTheDocument();
  });

  it("closes the popup from its close button and with Escape", () => {
    createHarness();
    const container = mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.click(screen.getByRole("button", { name: "Close details" }));
    expect(
      screen.queryByRole("button", { name: "Connect Order created to Validate order" }),
    ).not.toBeInTheDocument();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.keyDown(container, { key: "Escape" });
    expect(
      screen.queryByRole("button", { name: "Connect Order created to Validate order" }),
    ).not.toBeInTheDocument();
  });

  it("keeps the popup open when pressing inside it or on its node", () => {
    createHarness();
    mockContainer();

    fireEvent.pointerDown(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    fireEvent.pointerUp(screen.getByText("Order created"), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    const popup = document.querySelector("[data-slot='workflow-popup']") as Element;

    fireEvent.pointerDown(popup, { clientX: 300, clientY: 300 });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();

    fireEvent.pointerDown(screen.getByRole("option", { name: "Order created" }), {
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();

    fireEvent.pointerDown(screen.getByRole("option", { name: "Order created" }), {
      button: 1,
      clientX: 50 + 120 + 20,
      clientY: 60 + 160 + 20,
    });
    expect(
      screen.getByRole("button", { name: "Connect Order created to Validate order" }),
    ).toBeInTheDocument();
  });

  it("zooms with buttons, wheel, and reset", () => {
    createHarness();
    const container = mockContainer();

    fireEvent.click(screen.getByRole("button", { name: "Zoom in" }));
    expect(screen.getByText("125%")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Zoom out" }));
    expect(screen.getByText("100%")).toBeInTheDocument();

    fireEvent.wheel(container, { clientX: 200, clientY: 200, deltaY: -200 });
    expect(screen.queryByText("100%")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.getByText("100%")).toBeInTheDocument();
  });

  it("ignores input port presses", () => {
    const harness = createHarness();
    mockContainer();
    fireEvent.pointerDown(screen.getByRole("button", { name: "Validate order input" }), {
      button: 0,
    });
    expect(harness.onChange).not.toHaveBeenCalled();
  });

  it("skips edges that reference missing nodes", () => {
    renderWithProviders(
      <WorkflowMapper
        nodes={INITIAL_NODES}
        edges={[{ id: "ghost", sourceId: "missing", targetId: "b" }]}
      />,
    );
    expect(screen.getByText("Order created")).toBeInTheDocument();
    expect(document.querySelector("path[marker-end]")).not.toBeInTheDocument();
  });

  it("hides the palette when no node types are available", () => {
    renderWithProviders(<WorkflowMapper nodes={INITIAL_NODES} edges={[]} />);
    expect(screen.queryByText("Add node:")).not.toBeInTheDocument();
  });

  it("disables palette buttons without a change handler", () => {
    renderWithProviders(<WorkflowMapper nodes={INITIAL_NODES} edges={[]} available={AVAILABLE} />);
    expect(screen.getByRole("button", { name: "Process" })).toBeDisabled();
  });
});
