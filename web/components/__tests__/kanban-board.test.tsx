import { screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { KanbanBoard } from "@/components/kanban-board";
import { renderWithProviders } from "@/lib/tests";

const columns = [
  { id: "lead", title: "Lead" },
  { id: "quote", title: "Quote" },
  { id: "won", title: "Won" },
];

const cards = [
  { id: "c1", columnId: "lead", title: "Acme Inc" },
  { id: "c2", columnId: "lead", title: "Beta Corp" },
  { id: "c3", columnId: "quote", title: "Gamma LLC" },
];

describe("KanbanBoard", () => {
  it("groups cards into their columns and shows counts", () => {
    renderWithProviders(<KanbanBoard columns={columns} cards={cards} />);

    const lead = screen.getByText("Lead").closest("section") as HTMLElement;
    expect(within(lead).getByText("Acme Inc")).toBeInTheDocument();
    expect(within(lead).getByText("Beta Corp")).toBeInTheDocument();
    expect(within(lead).getByText("2")).toBeInTheDocument();

    const quote = screen.getByText("Quote").closest("section") as HTMLElement;
    expect(within(quote).getByText("Gamma LLC")).toBeInTheDocument();
    expect(within(quote).getByText("1")).toBeInTheDocument();
  });

  it("notifies on drag & drop across columns", () => {
    const onMoveCard = vi.fn();
    renderWithProviders(<KanbanBoard columns={columns} cards={cards} onMoveCard={onMoveCard} />);

    const card = screen.getByText("Acme Inc").closest("article") as HTMLElement;
    const wonColumn = screen.getByText("Won").closest("section") as HTMLElement;

    card.dispatchEvent(new Event("dragstart", { bubbles: true }));
    wonColumn.dispatchEvent(new Event("dragover", { bubbles: true }));
    wonColumn.dispatchEvent(new Event("drop", { bubbles: true }));

    expect(onMoveCard).toHaveBeenCalledWith("c1", "won");
  });

  it("anchors the drag ghost to the cursor and removes it on drop", () => {
    renderWithProviders(<KanbanBoard columns={columns} cards={cards} />);

    const card = screen.getByText("Acme Inc").closest("article") as HTMLElement;
    const cardCount = () => document.querySelectorAll('[data-slot="kanban-card"]').length;

    const setDragImage = vi.fn();
    const event = new Event("dragstart", { bubbles: true }) as DragEvent;
    Object.defineProperty(event, "dataTransfer", { value: { setDragImage } });
    Object.defineProperty(event, "clientX", { value: 120 });
    Object.defineProperty(event, "clientY", { value: 80 });

    card.dispatchEvent(event);

    expect(setDragImage).toHaveBeenCalledTimes(1);
    const [ghost, offsetX, offsetY] = setDragImage.mock.calls[0] as [HTMLElement, number, number];
    expect(ghost).toHaveAttribute("data-slot", "kanban-card");
    expect(document.body.contains(ghost)).toBe(true);
    expect(cardCount()).toBe(4);
    expect(offsetX).toBe(120);
    expect(offsetY).toBe(80);

    card.dispatchEvent(new Event("dragend", { bubbles: true }));

    expect(cardCount()).toBe(3);
    expect(document.body.contains(ghost)).toBe(false);
  });

  it("moves a card via the keyboard select", async () => {
    const user = userEvent.setup();
    const onMoveCard = vi.fn();
    renderWithProviders(<KanbanBoard columns={columns} cards={cards} onMoveCard={onMoveCard} />);

    await user.click(screen.getByRole("combobox", { name: "Move Gamma LLC" }));
    await user.click(await screen.findByRole("option", { name: "Won" }));

    expect(onMoveCard).toHaveBeenCalledWith("c3", "won");
  });

  it("runs the column add action", async () => {
    const user = userEvent.setup();
    const onAddCard = vi.fn();
    renderWithProviders(<KanbanBoard columns={columns} cards={cards} onAddCard={onAddCard} />);

    await user.click(screen.getByRole("button", { name: "Add card to Lead" }));

    expect(onAddCard).toHaveBeenCalledWith("lead");
  });
});
