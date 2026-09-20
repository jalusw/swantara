"use client";

import { ChevronDownIcon, ChevronRightIcon } from "lucide-react";
import { type ReactNode, useMemo, useRef, useState } from "react";

import { cn } from "@/lib/utils";

export type TreeNode = {
  id: string;
  label: ReactNode;
  icon?: ReactNode;
  children?: TreeNode[];
  actions?: ReactNode;
};

export type TreeViewProps = {
  items: TreeNode[];
  defaultExpandedIds?: string[];
  expandedIds?: string[];
  onExpandedChange?: (ids: string[]) => void;
  selectedId?: string | null;
  onSelect?: (id: string) => void;
  className?: string;
  "aria-label"?: string;
};

type FlatNode = {
  node: TreeNode;
  depth: number;
  id: string;
  parentId: string | null;
};

function flattenVisible(nodes: TreeNode[], expanded: Set<string>): FlatNode[] {
  const result: FlatNode[] = [];
  const walk = (nodeList: TreeNode[], depth: number, parentId: string | null) => {
    for (const node of nodeList) {
      result.push({ node, depth, id: node.id, parentId });
      if (node.children && node.children.length > 0 && expanded.has(node.id)) {
        walk(node.children, depth + 1, node.id);
      }
    }
  };
  walk(nodes, 0, null);
  return result;
}

function TreeBranch({
  node,
  depth,
  flatNodes,
  expanded,
  onToggle,
  focusedIndex,
  onKeyDown,
  selectedId,
  onSelect,
  rowRef,
}: {
  node: TreeNode;
  depth: number;
  flatNodes: FlatNode[];
  expanded: Set<string>;
  onToggle: (id: string) => void;
  focusedIndex?: number;
  onKeyDown: (event: React.KeyboardEvent, index: number) => void;
  selectedId?: string | null;
  onSelect?: (id: string) => void;
  rowRef: (id: string) => (element: HTMLDivElement | null) => void;
}) {
  const index = flatNodes.findIndex((flat) => flat.id === node.id);
  const hasChildren = Boolean(node.children && node.children.length > 0);
  const isExpanded = expanded.has(node.id);
  const isFocused = focusedIndex === index;

  const Chevron = isExpanded ? ChevronDownIcon : ChevronRightIcon;

  return (
    <div
      role="treeitem"
      ref={rowRef(node.id)}
      tabIndex={isFocused ? 0 : -1}
      aria-expanded={hasChildren ? isExpanded : undefined}
      aria-selected={selectedId ? selectedId === node.id : undefined}
      aria-level={depth + 1}
      data-selected={selectedId === node.id || undefined}
      data-slot="tree-row"
      onKeyDown={(event) => onKeyDown(event, index)}
      onClick={(event) => {
        if (
          event.target instanceof HTMLElement &&
          event.target.closest("[data-slot='tree-row-actions']")
        ) {
          return;
        }
        if (hasChildren) {
          onToggle(node.id);
        }
        onSelect?.(node.id);
      }}
      className={cn(
        "group/tree-row flex min-h-11 cursor-pointer items-center gap-1.5 rounded-md px-2 py-1.5 text-sm outline-none select-none focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring",
        selectedId === node.id && "bg-accent text-accent-foreground",
      )}
      style={{ paddingLeft: `${24 + depth * 20}px` }}
    >
      <span className="grid size-5 shrink-0 place-items-center text-muted-foreground">
        {hasChildren ? (
          <Chevron className="size-3.5" aria-hidden />
        ) : (
          <span className="size-3.5" aria-hidden />
        )}
      </span>
      {node.icon ? <span className="shrink-0 text-muted-foreground">{node.icon}</span> : null}
      <span className="min-w-0 flex-1 truncate">{node.label}</span>
      {node.actions ? (
        <span
          data-slot="tree-row-actions"
          className="shrink-0 opacity-0 transition-opacity group-hover/tree-row:opacity-100 group-focus-visible/tree-row:opacity-100"
        >
          {node.actions}
        </span>
      ) : null}
    </div>
  );
}

export function TreeView({
  items,
  defaultExpandedIds = [],
  expandedIds,
  onExpandedChange,
  selectedId,
  onSelect,
  className,
  "aria-label": ariaLabel = "Tree",
}: TreeViewProps) {
  const [internalExpanded, setInternalExpanded] = useState<Set<string>>(
    new Set(defaultExpandedIds),
  );
  const expanded = expandedIds ? new Set(expandedIds) : internalExpanded;

  const flatNodes = useMemo(() => flattenVisible(items, expanded), [items, expanded]);

  const [focusedIndex, setFocusedIndex] = useState(0);
  const rowRefs = useRef<Record<string, HTMLDivElement | null>>({});

  function setRowRef(id: string) {
    return (element: HTMLDivElement | null) => {
      rowRefs.current[id] = element;
    };
  }

  function moveTo(index: number) {
    const clamped = Math.max(0, Math.min(index, flatNodes.length - 1));
    setFocusedIndex(clamped);
    const target = flatNodes[clamped];
    if (target) {
      rowRefs.current[target.id]?.focus();
    }
  }

  function toggle(id: string) {
    const next = new Set(expanded);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    if (onExpandedChange) {
      onExpandedChange([...next]);
    } else {
      setInternalExpanded(next);
    }
  }

  function handleKeyDown(event: React.KeyboardEvent, index: number) {
    const flat = flatNodes[index];
    if (!flat) {
      return;
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      moveTo(index + 1);
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      moveTo(index - 1);
      return;
    }
    if (event.key === "Home") {
      event.preventDefault();
      moveTo(0);
      return;
    }
    if (event.key === "End") {
      event.preventDefault();
      moveTo(flatNodes.length - 1);
      return;
    }

    const hasChildren = Boolean(flat.node.children?.length);
    if (event.key === "ArrowRight") {
      event.preventDefault();
      if (hasChildren && !expanded.has(flat.id)) {
        toggle(flat.id);
      }
    }
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      if (hasChildren && expanded.has(flat.id)) {
        toggle(flat.id);
      } else if (flat.parentId) {
        moveTo(flatNodes.findIndex((f) => f.id === flat.parentId));
      }
    }
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      if (hasChildren) {
        toggle(flat.id);
      }
      onSelect?.(flat.id);
    }
  }

  return (
    <div
      data-slot="tree-view"
      role="tree"
      aria-label={ariaLabel}
      className={cn("flex flex-col", className)}
    >
      {flatNodes.map((flat) => (
        <TreeBranch
          key={flat.id}
          node={flat.node}
          depth={flat.depth}
          flatNodes={flatNodes}
          expanded={expanded}
          onToggle={toggle}
          focusedIndex={focusedIndex}
          onKeyDown={handleKeyDown}
          selectedId={selectedId}
          onSelect={onSelect}
          rowRef={setRowRef}
        />
      ))}
    </div>
  );
}
