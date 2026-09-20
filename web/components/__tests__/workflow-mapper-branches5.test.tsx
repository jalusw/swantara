import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import type { WorkflowEdge, WorkflowNode } from "@/components/workflow-mapper";
import { WorkflowMapper } from "@/components/workflow-mapper";
import { renderWithProviders } from "@/lib/tests";

const NODES: WorkflowNode[] = [
  { id: "a", def: "draft", label: "Draft", x: 10, y: 10 },
  { id: "b", def: "review", label: "Review", x: 220, y: 10 },
  { id: "c", def: "done", label: "Done", x: 430, y: 10 },
];

const EDGES: WorkflowEdge[] = [{ id: "a-b", sourceId: "a", targetId: "b" }];

const AVAILABLE = [
  { type: "draft", label: "Draft", tone: "primary" as const },
  { type: "review", label: "Review", tone: "success" as const },
];

function clickNode(label: string, clientX = 100, clientY = 50) {
  const node = screen.getByLabelText(label);
  fireEvent.pointerDown(node, { button: 0, clientX, clientY });
  fireEvent.pointerUp(node, { button: 0, clientX, clientY });
}

describe("WorkflowMapper branches5", () => {
  it("positions the popup for right-to-left flows", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <WorkflowMapper
        nodes={NODES}
        edges={EDGES}
        available={AVAILABLE}
        direction="rtl"
        onChange={() => undefined}
      />,
    );

    const node = screen.getByLabelText("Draft");
    node.focus();
    await user.keyboard("{Enter}");

    expect(await screen.findByRole("dialog", { name: "Draft details" })).toBeInTheDocument();
  });

  it("toggles the popup when clicking a node twice", () => {
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} onChange={() => undefined} />);

    clickNode("Draft");
    expect(screen.getByRole("dialog", { name: "Draft details" })).toBeInTheDocument();

    clickNode("Draft");
    expect(screen.queryByRole("dialog", { name: "Draft details" })).toBeNull();
  });

  it("removes the selected node with Delete", () => {
    const onChange = vi.fn();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} onChange={onChange} />);

    clickNode("Review", 300, 50);
    const canvas = screen.getByRole("application");
    canvas.focus();
    fireEvent.keyDown(canvas, { key: "Delete" });

    expect(onChange).toHaveBeenCalledWith(
      expect.arrayContaining([expect.objectContaining({ id: "a" })]),
      [],
    );
    expect(onChange.mock.calls[0]?.[0]).not.toContainEqual(expect.objectContaining({ id: "b" }));
  });

  it("deletes a node with an incoming edge through its delete button", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} onChange={onChange} />);

    await user.click(screen.getByRole("button", { name: "Delete Review" }));

    expect(onChange).toHaveBeenCalledWith(
      [expect.objectContaining({ id: "a" }), expect.objectContaining({ id: "c" })],
      [],
    );
  });

  it("creates an edge by dragging from an output port", () => {
    const onChange = vi.fn();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={[]} onChange={onChange} />);

    fireEvent.pointerDown(screen.getByLabelText("Draft output"), {
      button: 0,
      clientX: 194,
      clientY: 42,
    });
    fireEvent.pointerMove(screen.getByRole("application"), { clientX: 300, clientY: 42 });
    fireEvent.pointerUp(screen.getByRole("application"), { clientX: 430, clientY: 42 });

    expect(onChange).toHaveBeenCalledWith(
      NODES,
      expect.arrayContaining([expect.objectContaining({ sourceId: "a", targetId: "c" })]),
    );
  });
});
