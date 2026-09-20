import { describe, expect, it } from "vitest";
import {
  ANNOTATION_COLORS,
  clampPoint,
  createAnnotationId,
  createRectAnnotation,
  toRelativePoint,
  toSvgPoints,
} from "../pdf-annotation";

describe("createAnnotationId", () => {
  it("returns unique ids with the ann prefix", () => {
    const first = createAnnotationId();
    const second = createAnnotationId();
    expect(first).toMatch(/^ann-/);
    expect(first).not.toBe(second);
  });
});

describe("clampPoint", () => {
  it("keeps points inside the 0-100 range", () => {
    expect(clampPoint({ x: 50, y: 50 })).toEqual({ x: 50, y: 50 });
  });

  it("clamps out-of-range coordinates", () => {
    expect(clampPoint({ x: -10, y: 120 })).toEqual({ x: 0, y: 100 });
  });
});

describe("toRelativePoint", () => {
  it("converts client coordinates to percentages", () => {
    const rect = { left: 10, top: 20, width: 200, height: 100 } as DOMRect;
    expect(toRelativePoint(110, 70, rect)).toEqual({ x: 50, y: 50 });
  });
});

describe("createRectAnnotation", () => {
  it("normalizes inverted start and end points", () => {
    const annotation = createRectAnnotation("highlight", { x: 80, y: 70 }, { x: 20, y: 30 });
    expect(annotation).toMatchObject({ tool: "highlight", x: 20, y: 30, width: 60, height: 40 });
  });

  it("assigns the tool color", () => {
    const annotation = createRectAnnotation("note", { x: 0, y: 0 }, { x: 10, y: 10 });
    expect(annotation.color).toBe(ANNOTATION_COLORS.note);
  });
});

describe("toSvgPoints", () => {
  it("joins points into an svg polyline string", () => {
    expect(
      toSvgPoints([
        { x: 0, y: 0 },
        { x: 10, y: 20 },
      ]),
    ).toBe("0,0 10,20");
  });

  it("returns an empty string for no points", () => {
    expect(toSvgPoints([])).toBe("");
  });
});
