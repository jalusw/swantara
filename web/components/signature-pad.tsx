"use client";

import { EraserIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export type SignaturePadProps = {
  "aria-label"?: string;
  disabled?: boolean;
  /** Emits a PNG data URL after each finished stroke, or `null` when cleared. */
  onChange?: (dataUrl: string | null) => void;
  className?: string;
};

type StrokeState = {
  drawing: boolean;
  lastX: number;
  lastY: number;
};

export function SignaturePad({
  "aria-label": ariaLabel = "Signature",
  disabled = false,
  onChange,
  className,
}: SignaturePadProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const frameRef = useRef<HTMLDivElement>(null);
  const strokeRef = useRef<StrokeState>({ drawing: false, lastX: 0, lastY: 0 });
  const [hasInk, setHasInk] = useState(false);

  useEffect(() => {
    const frame = frameRef.current;
    if (!frame) {
      return;
    }
    const resize = () => {
      const canvas = canvasRef.current;
      if (!canvas) {
        return;
      }
      const dpr = window.devicePixelRatio || 1;
      canvas.width = Math.max(1, Math.round(frame.clientWidth * dpr));
      canvas.height = Math.max(1, Math.round(frame.clientHeight * dpr));
      const context = canvas.getContext("2d");
      context?.setTransform(dpr, 0, 0, dpr, 0, 0);
    };
    resize();
    const observer = new ResizeObserver(() => resize());
    observer.observe(frame);
    return () => observer.disconnect();
  }, []);

  function localPoint(event: React.PointerEvent<HTMLButtonElement>) {
    const rect = frameRef.current?.getBoundingClientRect();
    return {
      x: event.clientX - (rect?.left ?? 0),
      y: event.clientY - (rect?.top ?? 0),
    };
  }

  function startStroke(event: React.PointerEvent<HTMLButtonElement>) {
    if (disabled || event.button !== 0) {
      return;
    }
    event.preventDefault();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }
    const { x, y } = localPoint(event);
    const context = canvas.getContext("2d");
    if (!context) {
      return;
    }
    context.strokeStyle = getComputedStyle(canvas).color;
    context.lineWidth = 3;
    context.lineCap = "round";
    context.lineJoin = "round";
    context.beginPath();
    context.moveTo(x, y);
    strokeRef.current = { drawing: true, lastX: x, lastY: y };
  }

  function moveStroke(event: React.PointerEvent<HTMLButtonElement>) {
    const state = strokeRef.current;
    const context = canvasRef.current?.getContext("2d");
    if (!state.drawing || !context) {
      return;
    }
    const { x, y } = localPoint(event);
    context.beginPath();
    context.moveTo(state.lastX, state.lastY);
    context.lineTo(x, y);
    context.stroke();
    strokeRef.current = { drawing: true, lastX: x, lastY: y };
  }

  function endStroke() {
    if (!strokeRef.current.drawing) {
      return;
    }
    strokeRef.current.drawing = false;
    setHasInk(true);
    emit();
  }

  function cancelStroke() {
    strokeRef.current.drawing = false;
  }

  function emit() {
    const canvas = canvasRef.current;
    if (!canvas || !onChange) {
      return;
    }
    onChange(canvas.toDataURL("image/png"));
  }

  function clear() {
    const canvas = canvasRef.current;
    const context = canvas?.getContext("2d");
    if (canvas && context) {
      context.clearRect(0, 0, canvas.width, canvas.height);
    }
    strokeRef.current.drawing = false;
    setHasInk(false);
    onChange?.(null);
  }

  return (
    <div data-slot="signature-pad" className={cn("space-y-2", className)}>
      <div
        ref={frameRef}
        data-slot="signature-pad-frame"
        className="relative h-28 w-full overflow-hidden rounded-lg border border-dashed border-border bg-background"
      >
        <canvas
          ref={canvasRef}
          aria-hidden
          className="absolute inset-0 size-full text-foreground"
        />
        {!hasInk ? (
          <div
            aria-hidden
            className="pointer-events-none absolute inset-x-0 bottom-4 flex flex-col items-center gap-1 text-muted-foreground"
          >
            <span className="w-2/3 border-t border-border-muted text-transparent" />
            <span className="text-xs">{"Sign here"}</span>
          </div>
        ) : null}
        <button
          type="button"
          aria-label={`${ariaLabel}, ${"draw with a mouse, stylus or touch"}`}
          disabled={disabled}
          className="absolute inset-0 cursor-crosshair touch-none outline-none select-none focus-visible:ring-2 ring-offset-2 ring-offset-background focus-visible:ring-ring disabled:pointer-events-none"
          onPointerDown={startStroke}
          onPointerMove={moveStroke}
          onPointerUp={endStroke}
          onPointerCancel={cancelStroke}
        />
      </div>
      {hasInk ? (
        <div className="flex items-center justify-end">
          <Button type="button" variant="ghost" size="sm" disabled={disabled} onClick={clear}>
            <EraserIcon aria-hidden />
            {"Clear signature"}
          </Button>
        </div>
      ) : null}
    </div>
  );
}
