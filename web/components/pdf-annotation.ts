export type PdfPoint = { x: number; y: number };

export type PdfRectTool = "highlight" | "underline" | "note";

export type PdfAnnotationTool = PdfRectTool | "ink" | "select" | "eraser";

export type PdfRectAnnotation = {
  id: string;
  tool: PdfRectTool;
  x: number;
  y: number;
  width: number;
  height: number;
  color: string;
  note?: string;
};

export type PdfInkAnnotation = {
  id: string;
  tool: "ink";
  points: PdfPoint[];
};

export type PdfAnnotation = PdfRectAnnotation | PdfInkAnnotation;

export const ANNOTATION_COLORS: Record<PdfAnnotationTool, string> = {
  highlight: "#facc15",
  underline: "#3b82f6",
  note: "#a855f7",
  ink: "#ef4444",
  select: "#000000",
  eraser: "#000000",
};

let idCounter = 0;

export function createAnnotationId(): string {
  return `ann-${Date.now()}-${++idCounter}`;
}

export function clampPoint(point: PdfPoint): PdfPoint {
  return {
    x: Math.max(0, Math.min(100, point.x)),
    y: Math.max(0, Math.min(100, point.y)),
  };
}

export function toRelativePoint(clientX: number, clientY: number, rect: DOMRect): PdfPoint {
  return {
    x: ((clientX - rect.left) / rect.width) * 100,
    y: ((clientY - rect.top) / rect.height) * 100,
  };
}

export function createRectAnnotation(
  tool: PdfRectTool,
  start: PdfPoint,
  end: PdfPoint,
): Omit<PdfRectAnnotation, "id"> {
  const x = Math.min(start.x, end.x);
  const y = Math.min(start.y, end.y);
  const width = Math.abs(end.x - start.x);
  const height = Math.abs(end.y - start.y);
  return { tool, x, y, width, height, color: ANNOTATION_COLORS[tool] };
}

export function toSvgPoints(points: PdfPoint[]): string {
  return points.map((p) => `${p.x},${p.y}`).join(" ");
}
