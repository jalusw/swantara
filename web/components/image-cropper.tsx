"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";
import { Slider } from "./slider";

const MAX_ZOOM = 4;

type NaturalSize = { width: number; height: number };
type Translate = { x: number; y: number };
type DragState = {
  pointerX: number;
  pointerY: number;
  x: number;
  y: number;
};

export type ImageCropperProps = {
  src?: string;
  alt?: string;
  className?: string;
  minSize?: number;
  disabled?: boolean;
  onCrop?: (dataUrl: string) => void;
};

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

type Geometry = {
  scale: number;
  width: number;
  height: number;
  maxX: number;
  maxY: number;
};

function geometry(natural: NaturalSize, size: number, zoom: number): Geometry {
  const cover = Math.max(size / natural.width, size / natural.height);
  const scale = cover * zoom;
  const width = natural.width * scale;
  const height = natural.height * scale;
  return {
    scale,
    width,
    height,
    maxX: Math.min(0, size - width),
    maxY: Math.min(0, size - height),
  };
}

export function ImageCropper({
  src,
  alt = "Image to crop",
  className,
  minSize = 20,
  disabled = false,
  onCrop,
}: ImageCropperProps) {
  const sourceRef = useRef<HTMLImageElement | null>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const frameRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<DragState | null>(null);

  const [natural, setNatural] = useState<NaturalSize>();
  const [size, setSize] = useState(0);
  const [zoom, setZoom] = useState(1);
  const [translate, setTranslate] = useState<Translate>({ x: 0, y: 0 });
  const [dragging, setDragging] = useState(false);
  const [status, setStatus] = useState("");

  const ready = Boolean(natural && size > 0);
  const maxZoom = natural
    ? Math.max(
        1,
        Math.min(MAX_ZOOM, Math.min(natural.width, natural.height) / Math.max(1, minSize)),
      )
    : MAX_ZOOM;

  const clampedTranslate = useMemo(() => {
    if (!natural || size <= 0) return translate;
    const { maxX, maxY } = geometry(natural, size, zoom);
    return {
      x: clamp(translate.x, maxX, 0),
      y: clamp(translate.y, maxY, 0),
    };
  }, [natural, size, zoom, translate]);

  useEffect(() => {
    if (!src) {
      setNatural(undefined);
      return;
    }
    let cancelled = false;
    const image = new window.Image();
    image.onload = () => {
      if (cancelled) {
        return;
      }
      sourceRef.current = image;
      setZoom(1);
      setNatural({ width: image.naturalWidth, height: image.naturalHeight });
    };
    image.src = src;
    return () => {
      cancelled = true;
    };
  }, [src]);

  useEffect(() => {
    const frame = frameRef.current;
    if (!frame) {
      return;
    }
    const observer = new ResizeObserver((entries) => {
      const rect = entries[0]?.contentRect;
      if (rect) {
        setSize(Math.max(1, Math.min(rect.width, rect.height)));
      }
    });
    observer.observe(frame);
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    const canvas = canvasRef.current;
    const image = sourceRef.current;
    if (!canvas || !image || !natural || size <= 0) {
      return;
    }
    const dpr = window.devicePixelRatio || 1;
    canvas.width = Math.round(size * dpr);
    canvas.height = Math.round(size * dpr);
    const context = canvas.getContext("2d");
    if (!context) {
      return;
    }
    const view = geometry(natural, size, zoom);
    const sx = -clampedTranslate.x / view.scale;
    const sy = -clampedTranslate.y / view.scale;
    const sourceSize = size / view.scale;
    context.setTransform(1, 0, 0, 1, 0, 0);
    context.scale(dpr, dpr);
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, sx, sy, sourceSize, sourceSize, 0, 0, size, size);
  }, [natural, size, zoom, clampedTranslate]);

  useEffect(() => {
    const frame = frameRef.current;
    if (!frame || !natural || size <= 0 || disabled) {
      return;
    }
    const handler = (event: WheelEvent) => {
      event.preventDefault();
      const nextZoom = clamp(zoom * 1.002 ** -event.deltaY, 1, maxZoom);
      if (nextZoom === zoom) {
        return;
      }
      const bounds = frame.getBoundingClientRect();
      const cursorX = event.clientX - bounds.left;
      const cursorY = event.clientY - bounds.top;
      const view = geometry(natural, size, zoom);
      const sourceX = (cursorX - clampedTranslate.x) / view.scale;
      const sourceY = (cursorY - clampedTranslate.y) / view.scale;
      const next = geometry(natural, size, nextZoom);
      setZoom(nextZoom);
      setTranslate({
        x: clamp(cursorX - sourceX * next.scale, next.maxX, 0),
        y: clamp(cursorY - sourceY * next.scale, next.maxY, 0),
      });
    };
    frame.addEventListener("wheel", handler, { passive: false });
    return () => frame.removeEventListener("wheel", handler);
  }, [natural, size, zoom, clampedTranslate, maxZoom, disabled]);

  function onPointerDown(event: React.PointerEvent<HTMLButtonElement>) {
    if (!ready || disabled || event.button !== 0) {
      return;
    }
    event.preventDefault();
    event.currentTarget.setPointerCapture(event.pointerId);
    dragRef.current = {
      pointerX: event.clientX,
      pointerY: event.clientY,
      x: clampedTranslate.x,
      y: clampedTranslate.y,
    };
    setDragging(true);
  }

  function onPointerMove(event: React.PointerEvent<HTMLButtonElement>) {
    const drag = dragRef.current;
    if (!drag || !natural || size <= 0) {
      return;
    }
    const { maxX, maxY } = geometry(natural, size, zoom);
    setTranslate({
      x: clamp(drag.x + (event.clientX - drag.pointerX), maxX, 0),
      y: clamp(drag.y + (event.clientY - drag.pointerY), maxY, 0),
    });
  }

  function endDrag() {
    dragRef.current = null;
    setDragging(false);
  }

  function onKeyDown(event: React.KeyboardEvent<HTMLButtonElement>) {
    if (!natural || size <= 0 || disabled) {
      return;
    }
    const step = event.shiftKey ? 40 : 10;
    let dx = 0;
    let dy = 0;
    switch (event.key) {
      case "ArrowLeft":
        dx = -step;
        break;
      case "ArrowRight":
        dx = step;
        break;
      case "ArrowUp":
        dy = -step;
        break;
      case "ArrowDown":
        dy = step;
        break;
      default:
        return;
    }
    event.preventDefault();
    const { maxX, maxY } = geometry(natural, size, zoom);
    setTranslate({
      x: clamp(clampedTranslate.x + dx, maxX, 0),
      y: clamp(clampedTranslate.y + dy, maxY, 0),
    });
  }

  function applyCrop() {
    const image = sourceRef.current;
    if (!image || !natural || size <= 0 || !onCrop) {
      return;
    }
    const view = geometry(natural, size, zoom);
    const sourceSize = size / view.scale;
    const sx = -clampedTranslate.x / view.scale;
    const sy = -clampedTranslate.y / view.scale;
    const output = Math.max(1, Math.round(sourceSize));
    const canvas = document.createElement("canvas");
    canvas.width = output;
    canvas.height = output;
    const context = canvas.getContext("2d");
    if (!context) {
      return;
    }
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, sx, sy, sourceSize, sourceSize, 0, 0, output, output);
    onCrop(canvas.toDataURL("image/jpeg", 0.9));
    setStatus("Crop applied.");
  }

  function onZoomChange(next: number) {
    setZoom(clamp(next, 1, maxZoom));
  }

  const placeholder = natural ? "Loading…" : src ? "Preparing image…" : "No image selected";

  return (
    <div data-slot="image-cropper" className={cn("space-y-3", className)}>
      <output aria-live="polite" className="sr-only">
        {status}
      </output>
      <div
        ref={frameRef}
        data-slot="image-cropper-frame"
        className={cn(
          "relative aspect-square w-full overflow-hidden rounded-lg bg-muted select-none",
          !ready && "grid place-items-center",
        )}
      >
        {ready ? (
          <>
            <canvas ref={canvasRef} aria-label={alt} className="absolute inset-0 size-full" />
            <div
              className="pointer-events-none absolute inset-0 rounded-lg ring-2 ring-primary"
              aria-hidden
            />
            <button
              type="button"
              className={cn(
                "absolute inset-0 cursor-move touch-none",
                dragging && "cursor-grabbing",
              )}
              aria-label={"Drag to move the crop area, or use arrow keys to reposition"}
              disabled={disabled}
              onPointerDown={onPointerDown}
              onPointerMove={onPointerMove}
              onPointerUp={endDrag}
              onPointerCancel={endDrag}
              onKeyDown={onKeyDown}
            />
          </>
        ) : (
          <div className="absolute inset-0 grid place-items-center text-sm text-muted-foreground">
            {placeholder}
          </div>
        )}
      </div>
      {ready ? (
        <div className="flex items-center gap-3">
          <span className="text-xs text-muted-foreground">Zoom</span>
          <Slider
            value={[zoom]}
            min={1}
            max={maxZoom}
            step={0.05}
            onValueChange={(values) =>
              onZoomChange(typeof values === "number" ? values : values[0]!)
            }
            className="flex-1"
            aria-label={"Zoom"}
          />
          <span className="w-10 text-right text-xs text-muted-foreground tabular-nums">
            ×{zoom.toFixed(2)}
          </span>
        </div>
      ) : null}
      {ready ? (
        <div className="flex items-center justify-end">
          <Button size="sm" disabled={disabled || !onCrop} onClick={applyCrop}>
            Apply crop
          </Button>
        </div>
      ) : null}
    </div>
  );
}
