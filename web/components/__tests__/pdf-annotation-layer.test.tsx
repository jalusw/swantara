import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";
import type { PdfAnnotation, PdfAnnotationTool } from "@/components/pdf-annotation";
import { PdfAnnotationLayer } from "@/components/pdf-annotation-layer";
import { renderWithProviders } from "@/lib/tests";

const baseProps = {
  tool: "select" as const,
  annotations: [] as PdfAnnotation[],
  onAdd: () => {},
  onDelete: () => {},
  onUpdateNote: () => {},
};

function renderLayer(overrides: Record<string, unknown> = {}) {
  const onAdd = vi.fn();
  const onDelete = vi.fn();
  const onUpdateNote = vi.fn();
  renderWithProviders(
    <PdfAnnotationLayer
      {...baseProps}
      onAdd={onAdd}
      onDelete={onDelete}
      onUpdateNote={onUpdateNote}
      {...(overrides as object)}
    />,
  );
  return { onAdd, onDelete, onUpdateNote };
}

function renderStatefulLayer(tool: PdfAnnotationTool) {
  const onUpdateNote = vi.fn();
  function Harness() {
    const [annotations, setAnnotations] = useState<PdfAnnotation[]>([]);
    return (
      <PdfAnnotationLayer
        tool={tool}
        annotations={annotations}
        onAdd={(annotation) => setAnnotations((current) => [...current, annotation])}
        onDelete={(id) => setAnnotations((current) => current.filter((item) => item.id !== id))}
        onUpdateNote={(id, note) => {
          onUpdateNote(id, note);
          setAnnotations((current) =>
            current.map((item) => (item.id === id ? { ...item, note } : item)),
          );
        }}
      />
    );
  }
  renderWithProviders(<Harness />);
  return { onUpdateNote };
}

function stubLayerRect() {
  const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
  vi.spyOn(layer, "getBoundingClientRect").mockReturnValue({
    left: 0,
    top: 0,
    width: 400,
    height: 400,
  } as DOMRect);
}

function surface() {
  const element = document.querySelector("[data-slot='pdf-annotation-surface']") as Element;
  expect(element).toBeInTheDocument();
  return element;
}

function drawRect(element: Element, from: [number, number], to: [number, number]) {
  fireEvent.pointerDown(element, { clientX: from[0], clientY: from[1], pointerId: 1 });
  fireEvent.pointerMove(element, { clientX: to[0], clientY: to[1], pointerId: 1 });
  fireEvent.pointerUp(element, { pointerId: 1 });
}

function drawInk(element: Element, points: Array<[number, number]>) {
  const [first, ...rest] = points;
  fireEvent.pointerDown(element, { clientX: first![0], clientY: first![1], pointerId: 1 });
  for (const [x, y] of rest) {
    fireEvent.pointerMove(element, { clientX: x, clientY: y, pointerId: 1 });
  }
  fireEvent.pointerUp(element, { pointerId: 1 });
}

describe("PdfAnnotationLayer", () => {
  it("renders with data-slot", () => {
    renderLayer();
    expect(document.querySelector("[data-slot='pdf-annotation-layer']")).toBeInTheDocument();
  });

  it("hides the drawing surface for the select tool", () => {
    renderLayer({ tool: "select" });
    expect(document.querySelector("[data-slot='pdf-annotation-surface']")).not.toBeInTheDocument();
  });

  it("hides the drawing surface when disabled", () => {
    renderLayer({ tool: "highlight", disabled: true });
    expect(document.querySelector("[data-slot='pdf-annotation-surface']")).not.toBeInTheDocument();
  });

  it("shows the drawing surface for drawing tools", () => {
    for (const tool of ["highlight", "underline", "note", "ink"] as const) {
      const { unmount } = renderWithProviders(
        <PdfAnnotationLayer
          {...baseProps}
          tool={tool}
          onAdd={vi.fn()}
          onDelete={vi.fn()}
          onUpdateNote={vi.fn()}
        />,
      );
      expect(document.querySelector("[data-slot='pdf-annotation-surface']")).toBeInTheDocument();
      unmount();
    }
  });

  it("draws a highlight annotation", () => {
    const { onAdd } = renderLayer({ tool: "highlight" });
    stubLayerRect();
    drawRect(surface(), [40, 40], [120, 80]);
    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ tool: "highlight", x: 10, y: 10, width: 20, height: 10 }),
    );
  });

  it("draws an underline annotation", () => {
    const { onAdd } = renderLayer({ tool: "underline" });
    stubLayerRect();
    drawRect(surface(), [120, 80], [40, 40]);
    expect(onAdd).toHaveBeenCalledWith(expect.objectContaining({ tool: "underline" }));
  });

  it("ignores pointer moves and releases without an active draft", () => {
    const { onAdd } = renderLayer({ tool: "highlight" });
    stubLayerRect();
    const element = surface();
    fireEvent.pointerMove(element, { clientX: 100, clientY: 100, pointerId: 1 });
    fireEvent.pointerUp(element, { pointerId: 1 });
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("ignores drafts when the container has no rect", () => {
    const { onAdd } = renderLayer({ tool: "highlight" });
    const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
    vi.spyOn(layer, "getBoundingClientRect").mockReturnValue(undefined as unknown as DOMRect);
    const element = surface();
    fireEvent.pointerDown(element, { clientX: 40, clientY: 40, pointerId: 1 });
    fireEvent.pointerMove(element, { clientX: 120, clientY: 80, pointerId: 1 });
    fireEvent.pointerUp(element, { pointerId: 1 });
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("ignores draft moves when the container loses its rect", () => {
    const { onAdd } = renderLayer({ tool: "highlight" });
    const layer = document.querySelector("[data-slot='pdf-annotation-layer']") as HTMLElement;
    const rect = vi.spyOn(layer, "getBoundingClientRect").mockReturnValue({
      left: 0,
      top: 0,
      width: 400,
      height: 400,
    } as DOMRect);
    const element = surface();
    fireEvent.pointerDown(element, { clientX: 40, clientY: 40, pointerId: 1 });
    rect.mockReturnValue(undefined as unknown as DOMRect);
    fireEvent.pointerMove(element, { clientX: 120, clientY: 80, pointerId: 1 });
    fireEvent.pointerUp(element, { pointerId: 1 });
    expect(onAdd).toHaveBeenCalledTimes(1);
  });

  it("draws an ink annotation with multiple points", () => {
    const { onAdd } = renderLayer({ tool: "ink" });
    stubLayerRect();
    drawInk(surface(), [
      [10, 10],
      [50, 60],
      [90, 20],
    ]);
    expect(onAdd).toHaveBeenCalledTimes(1);
    expect(onAdd).toHaveBeenCalledWith(
      expect.objectContaining({ tool: "ink", points: expect.any(Array) }),
    );
  });

  it("drops ink drafts with a single point", () => {
    const { onAdd } = renderLayer({ tool: "ink" });
    stubLayerRect();
    const element = surface();
    fireEvent.pointerDown(element, { clientX: 10, clientY: 10, pointerId: 1 });
    fireEvent.pointerUp(element, { pointerId: 1 });
    expect(onAdd).not.toHaveBeenCalled();
  });

  it("renders note annotations", () => {
    renderLayer({
      annotations: [
        {
          id: "1",
          tool: "note",
          x: 0.1,
          y: 0.2,
          width: 0.3,
          height: 0.1,
          color: "#a855f7",
          note: "My note",
        },
      ],
    });
    const note = screen.getByRole("button", { name: "Note" });
    expect(note).toBeInTheDocument();
    expect(note).toHaveAttribute("title", "My note");
  });

  it("renders highlight and underline annotations with their styles", () => {
    renderLayer({
      annotations: [
        { id: "h", tool: "highlight", x: 0, y: 0, width: 0.2, height: 0.05, color: "#facc15" },
        { id: "u", tool: "underline", x: 0, y: 0.1, width: 0.2, height: 0.05, color: "#3b82f6" },
      ],
    });
    expect(document.querySelectorAll("[data-slot='pdf-annotation-rect']")).toHaveLength(2);
  });

  it("renders ink annotations as SVG", () => {
    renderLayer({
      annotations: [
        {
          id: "2",
          tool: "ink",
          points: [
            { x: 0, y: 0 },
            { x: 1, y: 1 },
          ],
        },
      ],
    });
    expect(document.querySelector("[data-slot='pdf-annotation-ink']")).toBeInTheDocument();
  });

  it("deletes an annotation from its delete button", async () => {
    const { onDelete } = renderLayer({
      annotations: [
        { id: "h", tool: "highlight", x: 0, y: 0, width: 0.2, height: 0.05, color: "#facc15" },
      ],
    });
    await userEvent.setup().click(screen.getByRole("button", { name: "Delete annotation" }));
    expect(onDelete).toHaveBeenCalledWith("h");
  });

  it("deletes through the eraser overlay when the eraser is active", async () => {
    const { onDelete } = renderLayer({
      tool: "eraser",
      annotations: [
        { id: "h", tool: "highlight", x: 0, y: 0, width: 0.2, height: 0.05, color: "#facc15" },
      ],
    });
    await userEvent.setup().click(screen.getByRole("button", { name: "Delete annotation" }));
    expect(onDelete).toHaveBeenCalledWith("h");
  });

  it("opens a note editor after drawing a note and saves it", async () => {
    const user = userEvent.setup();
    const { onUpdateNote } = renderStatefulLayer("note");
    stubLayerRect();
    drawRect(surface(), [40, 40], [120, 80]);

    const editor = await screen.findByLabelText("Write a note…");
    await user.type(editor, "Check the total");
    await user.click(screen.getByRole("button", { name: "Save note" }));
    expect(onUpdateNote).toHaveBeenCalledWith(expect.any(String), "Check the total");
    expect(screen.queryByLabelText("Write a note…")).not.toBeInTheDocument();
  });

  it("cancels the note editor", async () => {
    const user = userEvent.setup();
    const { onUpdateNote } = renderStatefulLayer("note");
    stubLayerRect();
    drawRect(surface(), [40, 40], [120, 80]);
    await screen.findByLabelText("Write a note…");
    await user.click(screen.getByRole("button", { name: "Cancel note" }));
    expect(screen.queryByLabelText("Write a note…")).not.toBeInTheDocument();
    expect(onUpdateNote).not.toHaveBeenCalled();
  });

  it("closes the note editor on Escape", async () => {
    const user = userEvent.setup();
    renderStatefulLayer("note");
    stubLayerRect();
    drawRect(surface(), [40, 40], [120, 80]);
    const editor = await screen.findByLabelText("Write a note…");
    await user.click(editor);
    await user.keyboard("{Escape}");
    expect(screen.queryByLabelText("Write a note…")).not.toBeInTheDocument();
  });

  it("edits an existing note with its saved text", async () => {
    const user = userEvent.setup();
    renderLayer({
      annotations: [
        {
          id: "1",
          tool: "note",
          x: 0.1,
          y: 0.2,
          width: 0.3,
          height: 0.1,
          color: "#a855f7",
          note: "Saved note",
        },
      ],
    });
    await user.click(screen.getByRole("button", { name: "Note" }));
    expect(await screen.findByDisplayValue("Saved note")).toBeInTheDocument();
  });

  it("uses label overrides", async () => {
    const user = userEvent.setup();
    renderLayer({
      tool: "note",
      labels: {
        toolNote: "Catatan",
        notePlaceholder: "Tulis catatan…",
        cancelNote: "Batal",
        saveNote: "Simpan",
        deleteAnnotation: "Hapus anotasi",
      },
      annotations: [
        {
          id: "1",
          tool: "note",
          x: 0.1,
          y: 0.2,
          width: 0.3,
          height: 0.1,
          color: "#a855f7",
        },
      ],
    });
    await user.click(screen.getByRole("button", { name: "Catatan" }));
    expect(await screen.findByLabelText("Tulis catatan…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Batal" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Simpan" })).toBeInTheDocument();
  });
});
