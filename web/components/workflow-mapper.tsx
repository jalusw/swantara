"use client";

import { ArrowRightIcon, CheckIcon } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export const NODE_WIDTH = 184;
export const NODE_HEIGHT = 64;

const WORLD_SIZE = 2048;
const PORT_HIT = 24;
const CLICK_THRESHOLD = 5;
const MIN_ZOOM = 0.25;
const MAX_ZOOM = 2;
const POPUP_WIDTH = 288;
const POPUP_GAP = 16;

type Point = { x: number; y: number };

export type WorkflowNodeTone = "primary" | "success" | "warning" | "info" | "violet";

export type WorkflowDirection = "ltr" | "rtl";

export type WorkflowNodeDef = {
  type: string;
  label: string;
  tone: WorkflowNodeTone;
  icon?: React.ComponentType<{ className?: string }>;
};

export type WorkflowNode = {
  id: string;
  def: string;
  label: string;
  x: number;
  y: number;
};

export type WorkflowEdge = {
  id: string;
  sourceId: string;
  targetId: string;
};

export type WorkflowMapperProps = {
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
  available?: WorkflowNodeDef[];
  onChange?: (nodes: WorkflowNode[], edges: WorkflowEdge[]) => void;
  /** Reading direction of the flow — defaults to left-to-right. */
  direction?: WorkflowDirection;
  className?: string;
};

const TONE_ACCENT: Record<WorkflowNodeTone, string> = {
  primary: "bg-primary",
  success: "bg-success",
  warning: "bg-warning",
  info: "bg-info",
  violet: "bg-foreground",
};

const TONE_BADGE: Record<WorkflowNodeTone, string> = {
  primary: "text-primary",
  success: "text-success",
  warning: "text-warning",
  info: "text-info",
  violet: "text-foreground",
};

type DragState =
  | {
      kind: "pan";
      pointerX: number;
      pointerY: number;
      originX: number;
      originY: number;
    }
  | {
      kind: "node";
      id: string;
      startWorld: Point;
      originX: number;
      originY: number;
      pointerX: number;
      pointerY: number;
    }
  | { kind: "link"; sourceId: string };

type PendingLink = { sourceId: string; x: number; y: number };

function portPosition(node: WorkflowNode, side: "in" | "out", direction: WorkflowDirection): Point {
  const midY = node.y + NODE_HEIGHT / 2;
  if (direction === "rtl") {
    return side === "in" ? { x: node.x + NODE_WIDTH, y: midY } : { x: node.x, y: midY };
  }
  return side === "in" ? { x: node.x, y: midY } : { x: node.x + NODE_WIDTH, y: midY };
}

function edgePath(start: Point, end: Point): string {
  const distance = Math.max(24, Math.abs(end.x - start.x) * 0.5);
  return `M ${start.x} ${start.y} C ${start.x + distance} ${start.y}, ${
    end.x - distance
  } ${end.y}, ${end.x} ${end.y}`;
}

export function WorkflowMapper({
  nodes,
  edges,
  available = [],
  onChange,
  direction = "ltr",
  className,
}: WorkflowMapperProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<DragState | null>(null);
  const zoomRef = useRef(1);
  const panRef = useRef<Point>({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState<Point>({ x: 0, y: 0 });
  const [selected, setSelected] = useState<{
    kind: "node" | "edge";
    id: string;
  } | null>(null);
  const [pending, setPending] = useState<PendingLink | null>(null);
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [popupNodeId, setPopupNodeId] = useState<string | null>(null);
  const popupRef = useRef<HTMLDivElement>(null);

  const popupNode = nodes.find((node) => node.id === popupNodeId) ?? null;

  useEffect(() => {
    if (popupNodeId) {
      popupRef.current?.focus();
    }
  }, [popupNodeId]);

  const popupPosition = popupNode
    ? {
        left: Math.min(
          Math.max(
            pan.x +
              popupNode.x * zoom +
              (direction === "rtl" ? -POPUP_WIDTH - POPUP_GAP : NODE_WIDTH + POPUP_GAP),
            8,
          ),
          Math.max(8, (containerRef.current?.clientWidth ?? 0) - POPUP_WIDTH - 8),
        ),
        top: Math.min(
          Math.max(pan.y + popupNode.y * zoom, 8),
          Math.max(8, (containerRef.current?.clientHeight ?? 0) - 8),
        ),
      }
    : null;

  function commit(nextNodes: WorkflowNode[], nextEdges: WorkflowEdge[]) {
    onChange?.(nextNodes, nextEdges);
  }

  function pointFromEvent(event: { clientX: number; clientY: number }): Point {
    const rect = containerRef.current?.getBoundingClientRect();
    const left = rect?.left ?? 0;
    const top = rect?.top ?? 0;
    return {
      x: (event.clientX - left - pan.x) / zoom,
      y: (event.clientY - top - pan.y) / zoom,
    };
  }

  function startPan(event: React.PointerEvent) {
    if (event.button !== 0) {
      return;
    }
    dragRef.current = {
      kind: "pan",
      pointerX: event.clientX,
      pointerY: event.clientY,
      originX: pan.x,
      originY: pan.y,
    };
  }

  function startDragNode(event: React.PointerEvent, node: WorkflowNode) {
    if (event.button !== 0) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    dragRef.current = {
      kind: "node",
      id: node.id,
      startWorld: pointFromEvent(event),
      originX: node.x,
      originY: node.y,
      pointerX: event.clientX,
      pointerY: event.clientY,
    };
    setDraggingId(node.id);
    setSelected({ kind: "node", id: node.id });
  }

  function startLink(event: React.PointerEvent, node: WorkflowNode) {
    if (event.button !== 0) {
      return;
    }
    event.preventDefault();
    event.stopPropagation();
    const point = pointFromEvent(event);
    setPending({ sourceId: node.id, x: point.x, y: point.y });
    dragRef.current = { kind: "link", sourceId: node.id };
  }

  function handlePointerMove(event: React.PointerEvent) {
    const drag = dragRef.current;
    if (!drag) {
      return;
    }
    if (drag.kind === "pan") {
      setPan({
        x: drag.originX + (event.clientX - drag.pointerX),
        y: drag.originY + (event.clientY - drag.pointerY),
      });
    } else if (drag.kind === "node") {
      const point = pointFromEvent(event);
      commit(
        nodes.map((node) =>
          node.id === drag.id
            ? {
                ...node,
                x: drag.originX + (point.x - drag.startWorld.x),
                y: drag.originY + (point.y - drag.startWorld.y),
              }
            : node,
        ),
        edges,
      );
    } else if (drag.kind === "link") {
      setPending((current) => {
        if (!current) {
          return current;
        }
        const point = pointFromEvent(event);
        return { ...current, x: point.x, y: point.y };
      });
    }
  }

  function endDrag(event: React.PointerEvent) {
    const drag = dragRef.current;
    if (!drag) {
      return;
    }
    dragRef.current = null;
    setDraggingId(null);
    if (drag.kind === "node") {
      const click = Math.hypot(event.clientX - drag.pointerX, event.clientY - drag.pointerY);
      if (click <= CLICK_THRESHOLD) {
        setPopupNodeId((current) => (current === drag.id ? null : drag.id));
      }
      return;
    }
    if (drag.kind !== "link") {
      return;
    }
    const point = pointFromEvent(event);
    const target = nodes
      .filter((node) => node.id !== drag.sourceId)
      .find((node) => {
        const port = portPosition(node, "in", direction);
        return Math.hypot(port.x - point.x, port.y - point.y) <= PORT_HIT;
      });
    if (target) {
      const duplicate = edges.some(
        (edge) => edge.sourceId === drag.sourceId && edge.targetId === target.id,
      );
      if (!duplicate) {
        commit(nodes, [
          ...edges,
          {
            id: `${drag.sourceId}-${target.id}`,
            sourceId: drag.sourceId,
            targetId: target.id,
          },
        ]);
      }
    }
    setPending(null);
  }

  function removeSelected() {
    if (!selected) {
      return;
    }
    if (selected.kind === "node") {
      commit(
        nodes.filter((node) => node.id !== selected.id),
        edges.filter((edge) => edge.sourceId !== selected.id && edge.targetId !== selected.id),
      );
    } else {
      commit(
        nodes,
        edges.filter((edge) => edge.id !== selected.id),
      );
    }
    setSelected(null);
    setPopupNodeId(null);
  }

  function toggleConnection(targetId: string) {
    if (!popupNodeId || targetId === popupNodeId) {
      return;
    }
    const existing = edges.find(
      (edge) => edge.sourceId === popupNodeId && edge.targetId === targetId,
    );
    if (existing) {
      commit(
        nodes,
        edges.filter((edge) => edge.id !== existing.id),
      );
      return;
    }
    commit(nodes, [
      ...edges,
      {
        id: `${popupNodeId}-${targetId}`,
        sourceId: popupNodeId,
        targetId,
      },
    ]);
  }

  useEffect(() => {
    if (popupNodeId && !nodes.some((node) => node.id === popupNodeId)) {
      setPopupNodeId(null);
    }
  }, [nodes, popupNodeId]);

  useEffect(() => {
    if (!popupNodeId) {
      return;
    }
    const handlePointerDown = (event: PointerEvent) => {
      const target = event.target as HTMLElement | null;
      if (!target) {
        return;
      }
      if (target.closest("[data-slot='workflow-popup']")) {
        return;
      }
      const nodeId = target.closest("[data-node-id]")?.getAttribute("data-node-id");
      if (nodeId === popupNodeId) {
        return;
      }
      setPopupNodeId(null);
    };
    document.addEventListener("pointerdown", handlePointerDown);
    return () => document.removeEventListener("pointerdown", handlePointerDown);
  }, [popupNodeId]);

  const zoomTo = useCallback((nextZoom: number, cursor?: { clientX: number; clientY: number }) => {
    const clamped = Math.min(MAX_ZOOM, Math.max(MIN_ZOOM, nextZoom));
    zoomRef.current = clamped;
    setZoom(clamped);
    const container = containerRef.current;
    if (!cursor || !container) {
      return;
    }
    const rect = container.getBoundingClientRect();
    const world = {
      x: (cursor.clientX - rect.left - panRef.current.x) / zoomRef.current,
      y: (cursor.clientY - rect.top - panRef.current.y) / zoomRef.current,
    };
    const nextPan = {
      x: cursor.clientX - rect.left - world.x * clamped,
      y: cursor.clientY - rect.top - world.y * clamped,
    };
    panRef.current = nextPan;
    setPan(nextPan);
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }
    const handler = (event: WheelEvent) => {
      event.preventDefault();
      zoomTo(zoomRef.current * 1.002 ** -event.deltaY, {
        clientX: event.clientX,
        clientY: event.clientY,
      });
    };
    container.addEventListener("wheel", handler, { passive: false });
    return () => container.removeEventListener("wheel", handler);
  }, [zoomTo]);

  function handleKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === "Delete" || event.key === "Backspace") {
      event.preventDefault();
      removeSelected();
    }
    if (event.key === "Escape" && popupNodeId) {
      event.preventDefault();
      setPopupNodeId(null);
    }
  }

  return (
    <div
      data-slot="workflow-mapper"
      className={cn(
        "flex min-h-[420px] flex-col overflow-hidden rounded-lg border border-border bg-background",
        className,
      )}
    >
      {/* biome-ignore lint/a11y/noNoninteractiveElementInteractions: canvas region uses role="application" for pointer/keyboard pan+zoom */}
      <div
        ref={containerRef}
        role="application"
        aria-label={"Workflow canvas"}
        className="relative min-h-0 flex-1 touch-none cursor-grab overflow-hidden outline-none select-none active:cursor-grabbing"
        onKeyDown={handleKeyDown}
        onPointerMove={handlePointerMove}
        onPointerUp={endDrag}
      >
        <div
          className="absolute top-0 left-0"
          style={{
            width: WORLD_SIZE,
            height: WORLD_SIZE,
            transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom})`,
            transformOrigin: "0 0",
          }}
          onPointerDown={startPan}
        >
          <svg
            viewBox={`0 0 ${WORLD_SIZE} ${WORLD_SIZE}`}
            className="absolute top-0 left-0 overflow-visible"
            width={WORLD_SIZE}
            height={WORLD_SIZE}
          >
            <title>{"Workflow connections"}</title>
            <defs>
              <marker
                id="workflow-edge-arrow"
                viewBox="0 0 10 10"
                refX="8"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" className="fill-[var(--color-border)]" />
              </marker>
              <marker
                id="workflow-edge-arrow-active"
                viewBox="0 0 10 10"
                refX="8"
                refY="5"
                markerWidth="6"
                markerHeight="6"
                orient="auto-start-reverse"
              >
                <path d="M 0 0 L 10 5 L 0 10 z" className="fill-[var(--color-primary)]" />
              </marker>
            </defs>
            {edges.map((edge) => {
              const source = nodes.find((node) => node.id === edge.sourceId);
              const target = nodes.find((node) => node.id === edge.targetId);
              if (!source || !target) {
                return null;
              }
              const path = edgePath(
                portPosition(source, "out", direction),
                portPosition(target, "in", direction),
              );
              const active = selected?.kind === "edge" && selected.id === edge.id;
              return (
                <g key={edge.id} aria-hidden>
                  <path
                    d={path}
                    className="fill-transparent stroke-transparent"
                    strokeWidth={14}
                    onPointerDown={(event) => {
                      event.stopPropagation();
                      setSelected({ kind: "edge", id: edge.id });
                    }}
                  />
                  <path
                    d={path}
                    markerEnd={`url(#${active ? "workflow-edge-arrow-active" : "workflow-edge-arrow"})`}
                    className={cn(
                      "pointer-events-none fill-transparent stroke-border",
                      active && "stroke-primary",
                    )}
                    strokeWidth={2}
                  />
                </g>
              );
            })}
            {pending && nodes.some((node) => node.id === pending.sourceId) ? (
              <path
                d={edgePath(
                  portPosition(
                    nodes.find((node) => node.id === pending.sourceId)! ?? nodes[0],
                    "out",
                    direction,
                  ),
                  { x: pending.x, y: pending.y },
                )}
                markerEnd="url(#workflow-edge-arrow-active)"
                className="pointer-events-none fill-transparent stroke-primary"
                strokeWidth={2}
                strokeDasharray="6 4"
                aria-hidden
              />
            ) : null}
          </svg>

          {nodes.map((node) => {
            const def = available.find((item) => item.type === node.def);
            const active = selected?.kind === "node" && selected.id === node.id;
            const dragging = draggingId === node.id;
            return (
              <div
                key={node.id}
                aria-label={node.label}
                role="option"
                aria-selected={active}
                data-node-id={node.id}
                tabIndex={0}
                className={cn(
                  "absolute rounded-lg border bg-card shadow-sm text-card-foreground focus-visible:ring-2 focus-visible:ring-ring outline-none",
                  active ? "border-primary ring-2 ring-primary" : "border-border",
                  dragging && "shadow-lg",
                )}
                style={{ left: node.x, top: node.y, width: NODE_WIDTH }}
                onPointerDown={(event) => startDragNode(event, node)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    setPopupNodeId((current) => (current === node.id ? null : node.id));
                  }
                }}
              >
                <span
                  className={cn("block h-1 rounded-t-lg", TONE_ACCENT[def?.tone ?? "primary"])}
                />
                <div className="flex items-start gap-2 p-2.5">
                  {def?.icon ? (
                    <def.icon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                  ) : null}
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm">{node.label}</p>
                    <p className={cn("truncate text-xs", TONE_BADGE[def?.tone ?? "primary"])}>
                      {node.def}
                    </p>
                  </div>
                  <button
                    type="button"
                    aria-label={`Delete ${node.label}`}
                    className="relative rounded p-0.5 text-muted-foreground outline-none after:absolute after:-inset-2.5 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                    onPointerDown={(event) => event.stopPropagation()}
                    onClick={() => {
                      commit(
                        nodes.filter((candidate) => candidate.id !== node.id),
                        edges.filter(
                          (edge) => edge.sourceId !== node.id && edge.targetId !== node.id,
                        ),
                      );
                      setSelected((prev) => (prev?.id === node.id ? null : prev));
                      setPopupNodeId((prev) => (prev === node.id ? null : prev));
                    }}
                  >
                    <span aria-hidden className="text-base leading-none">
                      ×
                    </span>
                  </button>
                </div>
              </div>
            );
          })}

          {nodes.map((node) => {
            const inPort = portPosition(node, "in", direction);
            return (
              <button
                key={`in-${node.id}`}
                type="button"
                aria-label={`${node.label} input`}
                className="absolute -translate-x-1/2 -translate-y-1/2 rounded-full focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                style={{ left: inPort.x, top: inPort.y }}
                onPointerDown={(event) => event.stopPropagation()}
              >
                <span className="block size-3 rounded-full border-2 border-background bg-muted-foreground/40" />
              </button>
            );
          })}
          {nodes.map((node) => {
            const outPort = portPosition(node, "out", direction);
            return (
              <button
                key={`out-${node.id}`}
                type="button"
                aria-label={`${node.label} output`}
                className="absolute -translate-x-1/2 -translate-y-1/2 rounded-full p-1.5 outline-none after:absolute after:-inset-2.5 after:content-[''] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                style={{
                  left: outPort.x,
                  top: outPort.y,
                }}
                onPointerDown={(event) => startLink(event, node)}
              >
                <span className="block size-3 rounded-full border-2 border-background bg-primary" />
              </button>
            );
          })}

          {popupNode && popupPosition ? (
            <div
              ref={popupRef}
              tabIndex={-1}
              role="dialog"
              data-slot="workflow-popup"
              aria-label={`${popupNode.label} details`}
              className="absolute z-popover w-72 rounded-lg bg-popover p-2.5 text-sm text-popover-foreground shadow-md ring-1 ring-foreground/10 outline-none focus-visible:ring-2 focus-visible:ring-ring"
              style={{ left: popupPosition.left, top: popupPosition.top }}
            >
              <div className="flex items-start justify-between gap-2">
                <div className="flex flex-col gap-0.5">
                  <h3 className="">{popupNode.label}</h3>
                  <p className="text-muted-foreground">{popupNode.def}</p>
                </div>
                <button
                  type="button"
                  aria-label={"Close details"}
                  className="relative rounded p-0.5 text-muted-foreground outline-none after:absolute after:-inset-2.5 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
                  onClick={() => setPopupNodeId(null)}
                >
                  <span aria-hidden className="text-base leading-none">
                    ×
                  </span>
                </button>
              </div>

              <div>
                <p className="mb-1.5 text-xs text-muted-foreground">{"Connections"}</p>
                <ul className="space-y-1">
                  {nodes
                    .filter((node) => node.id !== popupNodeId)
                    .map((node) => {
                      const connected = edges.some(
                        (edge) => edge.sourceId === popupNodeId && edge.targetId === node.id,
                      );
                      return (
                        <li key={node.id}>
                          <Button
                            size="sm"
                            variant="outline"
                            aria-label={
                              connected
                                ? `Disconnect ${popupNode.label} from ${node.label}`
                                : `Connect ${popupNode.label} to ${node.label}`
                            }
                            className={cn(
                              "w-full justify-between",
                              connected && "border-transparent bg-success/10",
                            )}
                            onClick={() => toggleConnection(node.id)}
                          >
                            <span className="min-w-0 flex-1 truncate text-left">{node.label}</span>
                            {connected ? (
                              <CheckIcon className="size-3.5 text-success" aria-hidden />
                            ) : (
                              <ArrowRightIcon
                                className="size-3.5 text-muted-foreground"
                                aria-hidden
                              />
                            )}
                          </Button>
                        </li>
                      );
                    })}
                </ul>
              </div>

              <p className="text-xs text-muted-foreground">
                Members toggles which node follows this one. You can also drag from this node’s
                output port to another node’s input port.
              </p>
            </div>
          ) : null}
        </div>
      </div>

      {available.length > 0 ? (
        <div className="shrink-0 border-t border-border p-3">
          <div className="flex flex-wrap items-center gap-2">
            <span className="mr-1 text-xs text-muted-foreground">Add node:</span>
            {available.map((def) => (
              <Button
                key={def.type}
                size="sm"
                variant="outline"
                disabled={!onChange}
                onClick={() => {
                  const id = `node-${Math.random().toString(36).slice(2, 8)}`;
                  commit(
                    [
                      ...nodes,
                      {
                        id,
                        def: def.type,
                        label: def.label,
                        x: 120 + (nodes.length % 3) * 48,
                        y: 96 + (Math.floor(nodes.length / 3) % 4) * 28,
                      },
                    ],
                    edges,
                  );
                  setSelected({ kind: "node", id });
                }}
              >
                {def.icon ? <def.icon className="size-3.5" /> : null}
                {def.label}
              </Button>
            ))}
          </div>
          <div className="mt-2 flex items-center justify-between gap-2">
            <p className="text-xs text-muted-foreground">
              Click a node to open its details and connect it to another step. You can also drag
              from a node’s output (●) to another node’s input to map it. Select a node or
              connection and press Backspace to remove it.
            </p>
            <div className="flex shrink-0 items-center gap-1">
              <Button
                size="sm"
                variant="ghost"
                aria-label={"Zoom out"}
                onClick={() => zoomTo(zoomRef.current * 0.8)}
              >
                −
              </Button>
              <span className="w-10 text-center text-xs text-muted-foreground tabular-nums">
                {Math.round(zoom * 100)}%
              </span>
              <Button
                size="sm"
                variant="ghost"
                aria-label={"Zoom in"}
                onClick={() => zoomTo(zoomRef.current * 1.25)}
              >
                +
              </Button>
              <Button
                size="sm"
                variant="ghost"
                disabled={!onChange}
                onClick={() => {
                  zoomRef.current = 1;
                  setZoom(1);
                  panRef.current = { x: 0, y: 0 };
                  setPan({ x: 0, y: 0 });
                }}
              >
                {"Reset"}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
