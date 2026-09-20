"use client";

import { ChevronLeftIcon, ChevronRightIcon, XIcon } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { createPortal } from "react-dom";
import { cn } from "@/lib/utils";
import { Button } from "./button";

export type GalleryImage = {
  src: string;
  alt: string;
  caption?: string;
};

export type LightboxProps = {
  images: GalleryImage[];
  index: number;
  onIndexChange?: (index: number) => void;
  onClose: () => void;
  "aria-label"?: string;
};

export function Lightbox({
  images,
  index,
  onIndexChange,
  onClose,
  "aria-label": ariaLabel,
}: LightboxProps) {
  const resolvedAriaLabel = ariaLabel ?? "Image preview";
  const dialogRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const [mounted, setMounted] = useState(false);

  const image = images[index];
  const count = images.length;

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      } else if (event.key === "ArrowLeft") {
        onIndexChange?.((index - 1 + count) % count);
      } else if (event.key === "ArrowRight") {
        onIndexChange?.((index + 1) % count);
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [index, count, onClose, onIndexChange]);

  useEffect(() => {
    closeRef.current?.focus();
    const dialog = dialogRef.current;
    if (!dialog) {
      return;
    }
    const onTab = (event: KeyboardEvent) => {
      if (event.key !== "Tab") {
        return;
      }
      const focusable = dialog.querySelectorAll<HTMLElement>(
        'button, [href], [tabindex]:not([tabindex="-1"])',
      );
      if (focusable.length === 0) {
        return;
      }
      const first = focusable[0];
      const last = focusable[focusable.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last?.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first?.focus();
      }
    };
    dialog.addEventListener("keydown", onTab);
    return () => dialog.removeEventListener("keydown", onTab);
  }, []);

  useEffect(() => {
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previous;
    };
  }, []);

  if (!mounted || !image || count === 0) {
    return null;
  }

  const preview = createPortal(
    <div
      data-slot="lightbox"
      ref={dialogRef}
      role="dialog"
      aria-modal="true"
      aria-label={resolvedAriaLabel}
      className="fixed inset-0 z-popover"
    >
      <button
        type="button"
        tabIndex={-1}
        aria-hidden="true"
        aria-label={"Close preview"}
        onClick={onClose}
        className="absolute inset-0 size-full cursor-default bg-overlay-strong outline-none"
      />

      <div className="relative flex h-full flex-col">
        <header className="flex items-center justify-between gap-2 px-4 py-3 text-sm text-white">
          <span className="truncate ">{image.caption ?? image.alt}</span>
          <Button
            ref={closeRef}
            type="button"
            variant="ghost"
            size="icon-sm"
            aria-label={"Close preview"}
            onClick={onClose}
            className="shrink-0 text-white hover:bg-white/10 hover:text-white dark:hover:bg-white/10"
          >
            <XIcon aria-hidden />
          </Button>
        </header>

        <div className="flex min-h-0 flex-1 items-center justify-center gap-2 px-2 pb-4 sm:gap-4">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={"Previous image"}
            disabled={count <= 1}
            onClick={() => onIndexChange?.((index - 1 + count) % count)}
            className="shrink-0 text-white hover:bg-white/10 hover:text-white disabled:opacity-20"
          >
            <ChevronLeftIcon aria-hidden />
          </Button>

          <div className="flex min-h-0 flex-1 items-center justify-center">
            <div
              role="img"
              aria-label={image.alt}
              className="min-h-0 h-[70vh] w-full max-w-3xl bg-contain bg-no-repeat bg-center"
              style={{ backgroundImage: `url(${image.src})` }}
            />
          </div>

          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={"Next image"}
            disabled={count <= 1}
            onClick={() => onIndexChange?.((index + 1) % count)}
            className="shrink-0 text-white hover:bg-white/10 hover:text-white disabled:opacity-20"
          >
            <ChevronRightIcon aria-hidden />
          </Button>
        </div>

        <footer className="px-4 py-2 text-center text-xs text-white/70 tabular-nums">
          {index + 1} / {count}
        </footer>
      </div>
    </div>,
    document.body,
  );

  return preview;
}

export type ImageGalleryProps = {
  images: GalleryImage[];
  columns?: number;
  className?: string;
  "aria-label"?: string;
};

export function ImageGallery({
  images,
  columns = 3,
  className,
  "aria-label": ariaLabel,
}: ImageGalleryProps) {
  const resolvedAriaLabel = ariaLabel ?? "Image gallery";
  const [index, setIndex] = useState<number | null>(null);

  return (
    <div
      data-slot="image-gallery"
      className={cn("grid gap-2", className)}
      style={{ gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))` }}
      aria-label={resolvedAriaLabel}
    >
      {images.map((image, imageIndex) => (
        <button
          key={image.src}
          type="button"
          onClick={() => setIndex(imageIndex)}
          aria-label={`Open ${image.alt}`}
          className="relative aspect-[4/3] overflow-hidden rounded-md border border-border bg-muted outline-none focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring"
        >
          <span
            aria-hidden
            className="absolute inset-0 bg-cover bg-center"
            style={{ backgroundImage: `url(${image.src})` }}
          />
        </button>
      ))}

      {index !== null && index < images.length ? (
        <Lightbox
          images={images}
          index={index}
          onIndexChange={setIndex}
          onClose={() => setIndex(null)}
          aria-label={resolvedAriaLabel}
        />
      ) : null}
    </div>
  );
}
