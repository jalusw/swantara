"use client";

import { ChevronDownIcon, PlusIcon } from "lucide-react";
import { type ReactNode, useMemo, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Select, SelectContent, SelectItem, SelectTrigger } from "./select";

export type KanbanColumn = {
  id: string;
  title: ReactNode;
};

export type KanbanCard = {
  id: string;
  columnId: string;
  title: ReactNode;
  subtitle?: ReactNode;
  meta?: ReactNode;
};

export type KanbanBoardProps = {
  columns: KanbanColumn[];
  cards: KanbanCard[];
  onMoveCard?: (cardId: string, columnId: string) => void;
  onAddCard?: (columnId: string) => void;
  className?: string;
  "aria-label"?: string;
};

export function KanbanBoard({
  columns,
  cards,
  onMoveCard,
  onAddCard,
  className,
  "aria-label": ariaLabel = "Kanban board",
}: KanbanBoardProps) {
  const [draggingId, setDraggingId] = useState<string | null>(null);
  const [announcement, setAnnouncement] = useState("");
  const draggingIdRef = useRef<string | null>(null);
  const dragGhostRef = useRef<HTMLElement | null>(null);

  const cardsById = useMemo(() => new Map(cards.map((card) => [card.id, card])), [cards]);
  const cardsByColumn = useMemo(
    () =>
      new Map(
        columns.map((column) => [column.id, cards.filter((card) => card.columnId === column.id)]),
      ),
    [columns, cards],
  );

  function move(cardId: string, columnId: string) {
    const card = cardsById.get(cardId);
    setDraggingId(null);
    draggingIdRef.current = null;
    if (card && card.columnId !== columnId) {
      onMoveCard?.(cardId, columnId);
      setAnnouncement(
        `Moved ${String(card.title)} to ${String(
          columns.find((column) => column.id === columnId)?.title ?? columnId,
        )}.`,
      );
    }
  }

  function handleDrop(columnId: string) {
    const cardId = draggingIdRef.current;
    if (cardId) {
      move(cardId, columnId);
    }
  }

  function startDrag(event: React.DragEvent<HTMLElement>, cardId: string) {
    draggingIdRef.current = cardId;
    setDraggingId(cardId);

    const dataTransfer = event.dataTransfer;
    if (!dataTransfer) {
      return;
    }

    // Chrome mispositions the native drag ghost when the card lives inside a
    // scrolled overflow container. Anchor an explicit drag image to the cursor
    // so it stays glued to the pointer regardless of scroll.
    const rect = event.currentTarget.getBoundingClientRect();
    const ghost = event.currentTarget.cloneNode(true) as HTMLElement;
    ghost.style.position = "fixed";
    ghost.style.top = "0";
    ghost.style.left = "-9999px";
    ghost.style.margin = "0";
    ghost.style.pointerEvents = "none";
    ghost.style.width = `${rect.width}px`;
    document.body.appendChild(ghost);
    dragGhostRef.current = ghost;
    dataTransfer.setDragImage(ghost, event.clientX - rect.left, event.clientY - rect.top);
  }

  function endDrag() {
    draggingIdRef.current = null;
    setDraggingId(null);
    dragGhostRef.current?.remove();
    dragGhostRef.current = null;
  }

  return (
    // biome-ignore lint/a11y/useSemanticElements: scrollable board region; grouping semantics come from role on a generic container
    <div
      data-slot="kanban-board"
      role="group"
      aria-label={ariaLabel}
      className={cn("flex gap-3 overflow-x-auto pb-2", className)}
    >
      <output aria-live="polite" className="sr-only">
        {announcement}
      </output>
      {columns.map((column) => (
        // biome-ignore lint/a11y/noNoninteractiveElementInteractions lint/a11y/noStaticElementInteractions: column is a native DnD drop target for cards
        <section
          key={column.id}
          data-slot="kanban-column"
          className="flex min-w-56 flex-1 flex-col gap-2 rounded-lg bg-muted/40 p-2 border border-border"
          onDragOver={(event) => event.preventDefault()}
          onDrop={() => handleDrop(column.id)}
        >
          <header className="flex items-center justify-between gap-2 px-1">
            <div className="flex items-center gap-2">
              <h3 className="text-sm">{column.title}</h3>
              <span className="rounded-full bg-muted px-1.5 text-xs tabular-nums text-muted-foreground">
                {(cardsByColumn.get(column.id) ?? []).length}
              </span>
            </div>
            {onAddCard ? (
              <Button
                type="button"
                size="icon-xs"
                variant="ghost"
                aria-label={`Add card to ${String(column.title)}`}
                onClick={() => onAddCard(column.id)}
              >
                <PlusIcon aria-hidden />
              </Button>
            ) : null}
          </header>

          <div className="flex min-h-16 flex-col gap-2">
            {(cardsByColumn.get(column.id) ?? []).map((card) => (
              // biome-ignore lint/a11y/noNoninteractiveElementInteractions: card is a native draggable DnD source
              <article
                key={card.id}
                data-slot="kanban-card"
                draggable
                onDragStart={(event) => startDrag(event, card.id)}
                onDragEnd={endDrag}
                className={cn(
                  "group/card cursor-grab rounded-md bg-background p-3 shadow-sm border border-border transition-shadow hover:shadow-md active:cursor-grabbing",
                  draggingId === card.id && "opacity-50",
                )}
              >
                <div className="flex items-start justify-between gap-2">
                  <p className="text-sm">{card.title}</p>
                  {columns.length > 1 && onMoveCard ? (
                    <Select
                      value={card.columnId}
                      onValueChange={(columnId) => {
                        if (columnId != null) {
                          move(card.id, columnId);
                        }
                      }}
                    >
                      <SelectTrigger
                        size="sm"
                        aria-label={`Move ${String(card.title)}`}
                        className="!border-transparent !p-0 !size-11 !after:-inset-1"
                      >
                        <ChevronDownIcon className="size-3 text-muted-foreground" />
                      </SelectTrigger>
                      <SelectContent>
                        {columns.map((option) => (
                          <SelectItem
                            key={option.id}
                            value={option.id}
                            disabled={option.id === card.columnId}
                          >
                            {option.title}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  ) : null}
                </div>
                {card.subtitle ? (
                  <p className="mt-1 text-xs text-muted-foreground">{card.subtitle}</p>
                ) : null}
                {card.meta ? (
                  <div className="mt-2 flex items-center gap-1 text-xs text-muted-foreground">
                    {card.meta}
                  </div>
                ) : null}
              </article>
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
