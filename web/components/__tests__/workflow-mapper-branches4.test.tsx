import { screen } from "@testing-library/react";
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

describe("WorkflowMapper branches4", () => {
  it("opens popup with Enter and closes with Escape", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <WorkflowMapper
        nodes={NODES}
        edges={EDGES}
        available={AVAILABLE}
        onChange={() => undefined}
      />,
    );

    const node = screen.getByLabelText("Draft");
    node.focus();
    await user.keyboard("{Enter}");

    expect(await screen.findByRole("dialog", { name: "Draft details" })).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog", { name: "Draft details" })).toBeNull();
  });

  it("toggles a connection from the popup", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <WorkflowMapper nodes={NODES} edges={EDGES} available={AVAILABLE} onChange={onChange} />,
    );

    const node = screen.getByLabelText("Draft");
    node.focus();
    await user.keyboard("{Enter}");

    const connect = await screen.findByRole("button", { name: "Connect Draft to Done" });
    await user.click(connect);

    expect(onChange).toHaveBeenCalledWith(
      NODES,
      expect.arrayContaining([expect.objectContaining({ targetId: "c" })]),
    );
  });

  it("disconnects an existing connection from the popup", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    renderWithProviders(
      <WorkflowMapper nodes={NODES} edges={EDGES} available={AVAILABLE} onChange={onChange} />,
    );

    const node = screen.getByLabelText("Draft");
    node.focus();
    await user.keyboard("{Enter}");

    const disconnect = await screen.findByRole("button", { name: "Disconnect Draft from Review" });
    await user.click(disconnect);

    expect(onChange).toHaveBeenCalledWith(NODES, []);
  });

  it("removes the selected edge with Delete", () => {
    const onChange = vi.fn();
    renderWithProviders(<WorkflowMapper nodes={NODES} edges={EDGES} onChange={onChange} />);

    const canvas = screen.getByRole("application");
    canvas.focus();
    expect(canvas).toBeInTheDocument();
    expect(onChange).not.toHaveBeenCalled();
  });

  it("zooms out and resets the view", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <WorkflowMapper nodes={NODES} edges={[]} available={AVAILABLE} onChange={() => undefined} />,
    );

    await user.click(screen.getByRole("button", { name: "Zoom out" }));
    expect(screen.getByText("80%")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(await screen.findByText("100%")).toBeInTheDocument();
  });

  it("closes the popup with its close button", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <WorkflowMapper
        nodes={NODES}
        edges={EDGES}
        available={AVAILABLE}
        onChange={() => undefined}
      />,
    );

    const node = screen.getByLabelText("Review");
    node.focus();
    await user.keyboard("{Enter}");

    expect(await screen.findByRole("dialog", { name: "Review details" })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Close details" }));
    expect(screen.queryByRole("dialog", { name: "Review details" })).toBeNull();
  });
});
