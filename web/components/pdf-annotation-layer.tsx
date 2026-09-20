"use client";

import { StickyNoteIcon, XIcon } from "lucide-react";
import {
  type CSSProperties,
  type PointerEvent as ReactPointerEvent,
  useEffect,
  useRef,
  useState,
} from "react";

import {
  ANNOTATION_COLORS,
  clampPoint,
  createAnnotationId,
  createRectAnnotation,
  type PdfAnnotation,
  type PdfAnnotationTool,
  type PdfPoint,
  type PdfRectAnnotation,
  type PdfRectTool,
  toRelativePoint,
  toSvgPoints,
} from "./pdf-annotation";

type Draft =
  | { kind: PdfRectTool; start: PdfPoint; end: PdfPoint }
  | { kind: "ink"; points: PdfPoint[] };

const DRAWING_TOOLS = new Set<PdfAnnotationTool>(["highlight", "underline", "note", "ink"]);

function isDrawingTool(tool: PdfAnnotationTool): tool is Draft["kind"] {
  return DRAWING_TOOLS.has(tool);
}

function rectStyle(annotation: PdfRectAnnotation): CSSProperties {
  return {
    left: `${annotation.x * 100}%`,
    top: `${annotation.y * 100}%`,
    width: `${annotation.width * 100}%`,
    height: `${annotation.height * 100}%`,
  };
}

type AnnotationRectProps = {
  annotation: PdfRectAnnotation;
  activeTool: PdfAnnotationTool;
  onDelete: () => void;
  onOpenNote: () => void;
};

type AnnotationRectWithRefProps = AnnotationRectProps & {
  editingId: string | null;
  noteEditorTriggerRef: React.RefObject<HTMLButtonElement | null>;
  labels: PdfAnnotationLayerLabels;
};

function AnnotationRect({
  annotation,
  activeTool,
  onDelete,
  onOpenNote,
  editingId,
  noteEditorTriggerRef,
  labels,
}: AnnotationRectWithRefProps) {
  const erasing = activeTool === "eraser";

  return (
    <div
      data-slot="pdf-annotation-rect"
      data-annotation-tool={annotation.tool}
      className="absolute"
      style={rectStyle(annotation)}
    >
      {annotation.tool === "highlight" && (
        <div
          className="pointer-events-none h-full w-full rounded-[2px]"
          style={{ backgroundColor: ANNOTATION_COLORS.highlight }}
        />
      )}
      {annotation.tool === "underline" && (
        <div
          className="pointer-events-none h-full w-full"
          style={{ borderBottom: `3px solid ${ANNOTATION_COLORS.underline}` }}
        />
      )}
      {annotation.tool === "note" && (
        <button
          ref={annotation.id === editingId ? noteEditorTriggerRef : undefined}
          tabIndex={annotation.id === editingId ? -1 : 0}
          type="button"
          data-slot="pdf-annotation-note"
          aria-label={labels.toolNote}
          title={annotation.note}
          onClick={onOpenNote}
          className="pointer-events-auto flex h-full w-full cursor-pointer items-start justify-center rounded-[2px] px-1 pt-0.5"
          style={{ backgroundColor: ANNOTATION_COLORS.note }}
        >
          <StickyNoteIcon aria-hidden className="size-3.5 shrink-0 text-foreground" />
        </button>
      )}

      {erasing ? (
        <button
          type="button"
          data-slot="pdf-annotation-eraser"
          aria-label={labels.deleteAnnotation}
          onClick={onDelete}
          className="pointer-events-auto absolute inset-0 cursor-pointer"
        />
      ) : (
        <button
          type="button"
          data-slot="pdf-annotation-delete"
          aria-label={labels.deleteAnnotation}
          onClick={onDelete}
          className="pointer-events-auto absolute -right-3 -top-3 z-10 flex size-11 min-h-11 min-w-11 cursor-pointer items-center justify-center rounded-full bg-background text-muted-foreground shadow-sm ring-1 ring-border-subtle hover:text-destructive"
        >
          <XIcon aria-hidden className="size-3" />
        </button>
      )}
    </div>
  );
}

export type PdfAnnotationLayerLabels = {
  toolNote?: string;
  toolInk?: string;
  deleteAnnotation?: string;
  notePlaceholder?: string;
  cancelNote?: string;
  saveNote?: string;
};

export type PdfAnnotationLayerProps = {
  tool: PdfAnnotationTool;
  annotations: PdfAnnotation[];
  disabled?: boolean;
  onAdd: (annotation: PdfAnnotation) => void;
  onDelete: (id: string) => void;
  onUpdateNote: (id: string, note: string) => void;
  labels?: PdfAnnotationLayerLabels;
};

function useDefaultLabels(): Required<PdfAnnotationLayerLabels> {
  return {
    toolNote: "Note",
    toolInk: "Ink",
    deleteAnnotation: "Delete annotation",
    notePlaceholder: "Write a note…",
    cancelNote: "Cancel note",
    saveNote: "Save note",
  };
}

export function PdfAnnotationLayer({
  tool,
  annotations,
  disabled = false,
  onAdd,
  onDelete,
  onUpdateNote,
  labels,
}: PdfAnnotationLayerProps) {
  const translatedDefaults = useDefaultLabels();
  const t = { ...translatedDefaults, ...labels };
  const containerRef = useRef<HTMLDivElement>(null);
  const noteEditorTriggerRef = useRef<HTMLButtonElement | null>(null);
  const noteEditorRef = useRef<HTMLTextAreaElement | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [noteDraft, setNoteDraft] = useState("");

  useEffect(() => {
    if (editingId) {
      noteEditorRef.current?.focus();
    }
  }, [editingId]);

  function beginDraft(event: ReactPointerEvent<HTMLDivElement>, kind: Draft["kind"]) {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) {
      return;
    }
    const point = clampPoint(toRelativePoint(event.clientX, event.clientY, rect));
    if (kind === "ink") {
      setDraft({ kind, points: [point] });
    } else {
      setDraft({ kind, start: point, end: point });
    }
    event.currentTarget.setPointerCapture?.(event.pointerId);
  }

  function moveDraft(event: ReactPointerEvent<HTMLDivElement>) {
    if (!draft) {
      return;
    }
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) {
      return;
    }
    const point = clampPoint(toRelativePoint(event.clientX, event.clientY, rect));
    setDraft((current) => {
      if (current?.kind === "ink") {
        return { kind: "ink", points: [...current.points, point] };
      }
      if (current) {
        return { kind: current.kind, start: current.start, end: point };
      }
      return current;
    });
  }

  function endDraft() {
    if (!draft) {
      return;
    }
    if (draft.kind === "ink") {
      if (draft.points.length >= 2) {
        onAdd({ id: createAnnotationId(), tool: "ink", points: draft.points });
      }
    } else {
      const annotation: PdfAnnotation = {
        id: createAnnotationId(),
        ...createRectAnnotation(draft.kind, draft.start, draft.end),
      };
      onAdd(annotation);
      if (draft.kind === "note") {
        setEditingId(annotation.id);
        setNoteDraft("");
      }
    }
    setDraft(null);
  }

  function openNoteEditor(annotation: PdfRectAnnotation) {
    setEditingId(annotation.id);
    setNoteDraft(annotation.note ?? "");
  }

  function saveNote(annotation: PdfRectAnnotation) {
    onUpdateNote(annotation.id, noteDraft);
    setEditingId(null);
    setNoteDraft("");
  }

  function closeNoteEditor() {
    setEditingId(null);
    setNoteDraft("");
    noteEditorTriggerRef.current?.focus();
  }

  const editingNote = annotations.find(
    (annotation) => annotation.id === editingId && annotation.tool === "note",
  );
  const noteTarget = editingNote && editingNote.tool === "note" ? editingNote : undefined;

  return (
    <div
      ref={containerRef}
      data-slot="pdf-annotation-layer"
      className="pointer-events-none absolute inset-0"
    >
      {!disabled && isDrawingTool(tool) && (
        <div
          data-slot="pdf-annotation-surface"
          className="pointer-events-auto absolute inset-0 cursor-crosshair touch-none"
          onPointerDown={(event) => beginDraft(event, tool)}
          onPointerMove={moveDraft}
          onPointerUp={endDraft}
          onPointerCancel={endDraft}
        />
      )}

      {annotations.map((annotation) =>
        annotation.tool === "ink" ? (
          <svg
            key={annotation.id}
            data-slot="pdf-annotation-ink"
            className="pointer-events-none absolute inset-0 h-full w-full"
            viewBox="0 0 100 100"
            preserveAspectRatio="none"
            aria-hidden
          >
            <title>{t.toolInk}</title>
            <polyline
              points={toSvgPoints(annotation.points)}
              fill="none"
              stroke={ANNOTATION_COLORS.ink}
              strokeWidth={2.5}
              strokeLinecap="round"
              strokeLinejoin="round"
              vectorEffect="non-scaling-stroke"
            />
          </svg>
        ) : (
          <AnnotationRect
            key={annotation.id}
            annotation={annotation}
            activeTool={tool}
            editingId={editingId}
            noteEditorTriggerRef={noteEditorTriggerRef}
            onDelete={() => onDelete(annotation.id)}
            onOpenNote={() => openNoteEditor(annotation)}
            labels={t}
          />
        ),
      )}

      {noteTarget && (
        <div className="pointer-events-auto absolute inset-x-0 top-0 z-10 flex justify-center p-4">
          <div className="flex min-w-64 flex-col gap-2 rounded-lg bg-background p-3 shadow-md ring-1 ring-border-subtle">
            <div className="flex items-center gap-2 text-sm">
              <StickyNoteIcon aria-hidden className="size-4 text-foreground" />
              {t.toolNote}
            </div>
            <textarea
              ref={noteEditorRef}
              aria-label={t.notePlaceholder}
              value={noteDraft}
              placeholder={t.notePlaceholder}
              onChange={(event) => setNoteDraft(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Escape") {
                  event.preventDefault();
                  closeNoteEditor();
                }
              }}
              rows={3}
              className="min-h-20 resize-y rounded-md border border-input bg-background px-2.5 py-1.5 text-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
            />
            <div className="flex justify-end gap-2">
              <button
                type="button"
                className="relative cursor-pointer rounded-md px-2.5 py-1.5 text-sm text-muted-foreground after:absolute after:-inset-y-2 after:content-[''] hover:bg-muted hover:text-foreground"
                onClick={closeNoteEditor}
              >
                {t.cancelNote}
              </button>
              <button
                type="button"
                className="relative cursor-pointer rounded-md bg-primary px-2.5 py-1.5 text-sm text-primary-foreground after:absolute after:-inset-y-2 after:content-[''] hover:brightness-110"
                onClick={() => saveNote(noteTarget)}
              >
                {t.saveNote}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
