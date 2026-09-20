"use client";

import { ChevronDownIcon, ChevronUpIcon, GripVerticalIcon } from "lucide-react";
import { useCallback, useState } from "react";

import { cn } from "@/lib/utils";

type SortableItem = { id: string };

export type SortableListProps<T extends SortableItem> = {
  items: T[];
  onReorder: (items: T[]) => void;
  render: (item: T) => React.ReactNode;
  className?: string;
};

export function SortableList<T extends SortableItem>({
  items,
  onReorder,
  render,
  className,
}: SortableListProps<T>) {
  const [dragIndex, setDragIndex] = useState<number | null>(null);
  const [announcement, setAnnouncement] = useState("");

  const move = useCallback(
    (index: number, direction: -1 | 1) => {
      const target = index + direction;
      if (target < 0 || target >= items.length) {
        return;
      }
      const next = [...items];
      [next[index], next[target]] = [next[target]!, next[index]!];
      onReorder(next);
      setAnnouncement(direction === -1 ? "Moved item up." : "Moved item down.");
    },
    [items, onReorder],
  );

  const handleDrop = useCallback(
    (targetIndex: number) => {
      if (dragIndex === null || dragIndex === targetIndex) {
        setDragIndex(null);
        return;
      }
      const next = [...items];
      const [moved] = next.splice(dragIndex, 1);
      next.splice(targetIndex, 0, moved!);
      setDragIndex(null);
      onReorder(next);
    },
    [dragIndex, items, onReorder],
  );

  return (
    <div data-slot="sortable-list" className={cn("flex flex-col gap-2", className)}>
      <ul className="flex flex-col gap-2">
        {items.map((item, index) => (
          // biome-ignore lint/a11y/noNoninteractiveElementInteractions: drag target; keyboard reordering via the up/down buttons
          <li
            key={item.id}
            data-slot="sortable-item"
            draggable
            onDragStart={() => setDragIndex(index)}
            onDragOver={(event) => event.preventDefault()}
            onDrop={() => handleDrop(index)}
            className={cn(
              "flex items-center gap-2 rounded-lg border border-border bg-background p-2",
              dragIndex === index && "opacity-50",
            )}
          >
            <span
              aria-hidden
              title={"Drag to reorder"}
              className="relative flex min-h-11 min-w-11 cursor-grab touch-none items-center justify-center rounded-sm p-2.5 text-muted-foreground select-none hover:text-foreground active:cursor-grabbing after:absolute after:-inset-1 after:content-['']"
            >
              <GripVerticalIcon className="size-4" />
            </span>
            <div className="flex flex-1 min-w-0">{render(item)}</div>
            <div className="flex shrink-0 items-center gap-0.5">
              <button
                type="button"
                aria-label={"Move up"}
                title={"Move up"}
                disabled={index === 0}
                onClick={() => move(index, -1)}
                className="relative flex min-h-11 min-w-11 items-center justify-center rounded-sm p-2.5 text-muted-foreground outline-none after:absolute after:-inset-1 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40"
              >
                <ChevronUpIcon className="size-4" aria-hidden />
              </button>
              <button
                type="button"
                aria-label={"Move down"}
                title={"Move down"}
                disabled={index === items.length - 1}
                onClick={() => move(index, 1)}
                className="relative flex min-h-11 min-w-11 items-center justify-center rounded-sm p-2.5 text-muted-foreground outline-none after:absolute after:-inset-1 after:content-[''] hover:bg-muted hover:text-foreground focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-40"
              >
                <ChevronDownIcon className="size-4" aria-hidden />
              </button>
            </div>
          </li>
        ))}
      </ul>
      <output aria-live="polite" className="sr-only">
        {announcement}
      </output>
    </div>
  );
}
