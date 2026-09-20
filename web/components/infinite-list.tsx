"use client";

import type { Key, ReactNode } from "react";
import { useCallback, useEffect, useRef } from "react";

import { cn } from "@/lib/utils";

export type InfiniteListProps<T> = {
  items: T[];
  renderItem: (item: T, index: number) => ReactNode;
  getItemKey: (item: T, index: number) => Key;
  /** Called when the end of the list scrolls into view. */
  onEndReached?: () => void;
  /** Whether more items can still be loaded. Hides the sentinel when false. */
  hasMore?: boolean;
  /** Suppresses further `onEndReached` calls while true. */
  loading?: boolean;
  loadingIndicator?: ReactNode;
  emptyState?: ReactNode;
  "aria-label"?: string;
  className?: string;
};

function DefaultLoading() {
  return (
    <output
      data-slot="infinite-list-loading"
      className="flex justify-center py-4"
      aria-live="polite"
    >
      <span className="flex items-center gap-2 text-sm text-muted-foreground">
        <span
          aria-hidden
          className="size-3 animate-spin rounded-full border-2 border-muted-foreground/30 border-t-muted-foreground"
        />
      </span>
    </output>
  );
}

export function InfiniteList<T>({
  items,
  renderItem,
  getItemKey,
  onEndReached,
  hasMore = true,
  loading = false,
  loadingIndicator,
  emptyState,
  "aria-label": ariaLabel = "Infinite list",
  className,
}: InfiniteListProps<T>) {
  const containerRef = useRef<HTMLDivElement>(null);
  const sentinelRef = useRef<HTMLDivElement>(null);
  const onEndReachedRef = useRef(onEndReached);
  const hasMoreRef = useRef(hasMore);
  const loadingRef = useRef(loading);
  onEndReachedRef.current = onEndReached;
  hasMoreRef.current = hasMore;
  loadingRef.current = loading;

  const handleIntersection = useCallback((entries: IntersectionObserverEntry[]) => {
    if (loadingRef.current || !hasMoreRef.current) return;
    if (entries.some((entry) => entry.isIntersecting)) {
      onEndReachedRef.current?.();
    }
  }, []);

  useEffect(() => {
    const sentinel = sentinelRef.current;
    const root = containerRef.current;
    if (!sentinel || !root || !hasMore) return;

    const observer = new IntersectionObserver(handleIntersection, {
      root,
      rootMargin: "200px 0px",
    });
    observer.observe(sentinel);

    return () => observer.disconnect();
  }, [handleIntersection, hasMore]);

  return (
    <div
      ref={containerRef}
      data-slot="infinite-list"
      role="feed"
      aria-label={ariaLabel}
      aria-busy={loading || undefined}
      className={cn("flex flex-col overflow-y-auto", className)}
    >
      {items.length === 0 && !loading ? (emptyState ?? null) : null}

      {items.map((item, index) => (
        <article
          key={getItemKey(item, index)}
          className="min-w-0 shrink-0 [contain-intrinsic-size:auto_8rem] [content-visibility:auto]"
        >
          {renderItem(item, index)}
        </article>
      ))}

      {hasMore ? <div ref={sentinelRef} aria-hidden data-slot="infinite-list-sentinel" /> : null}

      {loading ? (loadingIndicator ?? <DefaultLoading />) : null}
    </div>
  );
}
