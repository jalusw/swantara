"use client";

import { ChevronDownIcon, ChevronRightIcon } from "lucide-react";
import { useMemo, useState } from "react";

import { cn } from "@/lib/utils";

export const ORG_CHART_NODE_WIDTH = 184;
export const ORG_CHART_NODE_HEIGHT = 64;

const GAP_X = 32;
const GAP_Y = 72;

export type OrgChartTone = "primary" | "success" | "warning" | "info" | "violet" | "neutral";

export type OrgChartNode = {
  id: string;
  label: string;
  description?: string;
  tone?: OrgChartTone;
  icon?: React.ComponentType<{ className?: string }>;
  children?: OrgChartNode[];
};

export type OrgChartProps = {
  root: OrgChartNode;
  ariaLabel?: string;
  /** Whether nodes with children render an expand/collapse toggle. Defaults to true. */
  collapsible?: boolean;
  /** Ids of subtrees that start collapsed. */
  defaultCollapsedIds?: string[];
  onSelect?: (node: OrgChartNode) => void;
  className?: string;
};

const TONE_ACCENT: Record<OrgChartTone, string> = {
  primary: "bg-primary",
  success: "bg-success",
  warning: "bg-warning",
  info: "bg-info",
  violet: "bg-foreground",
  neutral: "bg-muted-foreground",
};

type PositionedNode = {
  node: OrgChartNode;
  x: number;
  y: number;
  centerX: number;
};

type Layout = {
  nodes: PositionedNode[];
  edges: { from: PositionedNode; to: PositionedNode }[];
  width: number;
  height: number;
};

/**
 * Lays the tree out top-down. Leaves occupy one slot; an expanded parent spans
 * the combined width of its children's subtrees and is centered above them.
 */
function layoutChart(root: OrgChartNode, collapsed: ReadonlySet<string>): Layout {
  const nodes: PositionedNode[] = [];
  const edges: { from: PositionedNode; to: PositionedNode }[] = [];

  function place(
    node: OrgChartNode,
    depth: number,
    offsetX: number,
  ): { width: number; centerX: number; children: PositionedNode[] } {
    const children = node.children ?? [];
    const expanded = children.length > 0 && !collapsed.has(node.id);

    let width: number;
    let centerX: number;
    let childrenPositions: PositionedNode[];

    if (expanded) {
      let cursor = offsetX;
      const results = children.map((child) => {
        const result = place(child, depth + 1, cursor);
        cursor += result.width;
        return result;
      });
      width = results.reduce((sum, result) => sum + result.width, 0);
      centerX = offsetX + width / 2;
      childrenPositions = results.flatMap((result) => result.children);
    } else {
      width = ORG_CHART_NODE_WIDTH + GAP_X;
      centerX = offsetX + ORG_CHART_NODE_WIDTH / 2;
      childrenPositions = [];
    }

    const self: PositionedNode = {
      node,
      x: centerX - ORG_CHART_NODE_WIDTH / 2,
      y: depth * (ORG_CHART_NODE_HEIGHT + GAP_Y),
      centerX,
    };
    nodes.push(self);

    if (expanded) {
      for (const child of childrenPositions) {
        edges.push({ from: self, to: child });
      }
    }

    return {
      width,
      centerX,
      children: expanded ? childrenPositions : [self],
    };
  }

  const rootResult = place(root, 0, 0);
  const height = Math.max(...nodes.map((node) => node.y + ORG_CHART_NODE_HEIGHT)) ?? 0;
  return { nodes, edges, width: rootResult.width, height };
}

export function OrgChart({
  root,
  ariaLabel,
  collapsible = true,
  defaultCollapsedIds,
  onSelect,
  className,
}: OrgChartProps) {
  const resolvedAriaLabel = ariaLabel ?? "Organization chart";
  const [collapsed, setCollapsed] = useState<ReadonlySet<string>>(
    () => new Set(defaultCollapsedIds ?? []),
  );

  const layout = useMemo(() => layoutChart(root, collapsed), [root, collapsed]);

  function toggle(nodeId: string) {
    setCollapsed((current) => {
      const next = new Set(current);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  }

  return (
    <section
      data-slot="org-chart"
      aria-label={resolvedAriaLabel}
      className={cn("overflow-auto rounded-lg border border-border bg-background", className)}
    >
      <div
        className="relative p-4"
        style={{ width: layout.width + 32, minHeight: layout.height + 32 }}
      >
        <svg
          width={layout.width}
          height={layout.height}
          aria-hidden
          className="pointer-events-none absolute left-4 top-4 overflow-visible"
        >
          <title>{"Organization chart connections"}</title>
          {layout.edges.map((edge) => {
            const startY = edge.from.y + ORG_CHART_NODE_HEIGHT;
            const midY = startY + GAP_Y / 2;
            const path = `M ${edge.from.centerX} ${startY} V ${midY} H ${edge.to.centerX} V ${edge.to.y}`;
            return (
              <path
                key={`${edge.from.node.id}-${edge.to.node.id}`}
                d={path}
                fill="none"
                className="stroke-border"
                strokeWidth={2}
              />
            );
          })}
        </svg>

        {layout.nodes.map(({ node, x, y }) => {
          const children = node.children ?? [];
          const hasChildren = children.length > 0;
          const expanded = hasChildren && !collapsed.has(node.id);
          const Icon = node.icon;
          return (
            <div
              key={node.id}
              data-node-id={node.id}
              className={cn(
                "absolute flex items-stretch overflow-hidden rounded-lg border border-border bg-card text-card-foreground shadow-sm",
              )}
              style={{
                left: x,
                top: y,
                width: ORG_CHART_NODE_WIDTH,
                minHeight: ORG_CHART_NODE_HEIGHT,
              }}
            >
              <span
                aria-hidden
                className={cn("block w-1 shrink-0", TONE_ACCENT[node.tone ?? "neutral"])}
              />
              <button
                type="button"
                onClick={() => onSelect?.(node)}
                className="flex min-w-0 flex-1 items-center gap-2 p-2.5 text-left outline-none hover:bg-muted focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
              >
                {Icon ? <Icon className="size-4 shrink-0 text-muted-foreground" /> : null}
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-sm">{node.label}</span>
                  {node.description ? (
                    <span className="block truncate text-xs text-muted-foreground">
                      {node.description}
                    </span>
                  ) : null}
                </span>
              </button>
              {hasChildren && collapsible ? (
                <button
                  type="button"
                  aria-label={expanded ? `Collapse ${node.label}` : `Expand ${node.label}`}
                  aria-expanded={expanded}
                  onClick={() => toggle(node.id)}
                  className="grid w-8 shrink-0 place-items-center text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-ring"
                >
                  {expanded ? (
                    <ChevronDownIcon className="size-4" aria-hidden />
                  ) : (
                    <ChevronRightIcon className="size-4" aria-hidden />
                  )}
                </button>
              ) : null}
            </div>
          );
        })}
      </div>
    </section>
  );
}
