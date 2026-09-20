"use client";

import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { type ReactNode, useState } from "react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export type KpiWallSlide = {
  id: string;
  title: string;
  subtitle?: string;
  content: ReactNode;
};

export type KpiWallProps = {
  slides: KpiWallSlide[];
  interval?: number;
  ariaLabel?: string;
  className?: string;
};

export function KpiWall({ slides, ariaLabel = "KPI wall", className }: KpiWallProps) {
  const [index, setIndex] = useState(0);

  const slideCount = slides.length;
  const clampedIndex = index > slideCount - 1 ? 0 : index;
  const current = slides[clampedIndex]!;

  if (slideCount === 0) {
    return null;
  }

  return (
    <section
      data-slot="kpi-wall"
      aria-label={ariaLabel}
      className={cn(
        "flex min-h-96 flex-col overflow-hidden rounded-xl border border-border bg-background",
        className,
      )}
    >
      <header className="flex shrink-0 items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="min-w-0">
          <h2 className="truncate font-heading text-base tracking-tight">{current.title}</h2>
          {current.subtitle ? (
            <p className="truncate text-xs text-muted-foreground">{current.subtitle}</p>
          ) : null}
        </div>
        <div className="flex shrink-0 items-center gap-1">
          <Button
            size="icon-sm"
            variant="ghost"
            aria-label={"Previous slide"}
            onClick={() => setIndex((currentIndex) => (currentIndex - 1 + slideCount) % slideCount)}
          >
            <ChevronLeftIcon aria-hidden />
          </Button>
          <div className="flex items-center gap-1 px-1" role="tablist" aria-label={"Slides"}>
            {slides.map((slide, slideIndex) => (
              <button
                key={slide.id}
                type="button"
                role="tab"
                aria-selected={slideIndex === clampedIndex}
                aria-label={`Go to slide ${slideIndex + 1}: ${slide.title}`}
                onClick={() => setIndex(slideIndex)}
                className={cn(
                  "size-2 rounded-full outline-none focus-visible:ring-2 focus-visible:ring-ring",
                  slideIndex === clampedIndex
                    ? "bg-primary"
                    : "bg-muted-foreground/30 hover:bg-muted-foreground/60",
                )}
              />
            ))}
          </div>
          <Button
            size="icon-sm"
            variant="ghost"
            aria-label={"Next slide"}
            onClick={() => setIndex((currentIndex) => (currentIndex + 1) % slideCount)}
          >
            <ChevronRightIcon aria-hidden />
          </Button>
        </div>
      </header>

      <div className="min-h-0 flex-1 overflow-auto">
        <div key={current.id} data-slot="kpi-wall-slide" className="min-h-full p-4 sm:p-6">
          {current.content}
        </div>
      </div>
    </section>
  );
}
