"use client";

import type { CSSProperties, ReactNode } from "react";
import { useEffect, useRef, useState } from "react";

import { cn } from "@/lib/utils";

export type TextClampProps = {
  children: ReactNode;
  /** Number of visible lines when collapsed. */
  lines?: number;
  showMoreLabel?: string;
  showLessLabel?: string;
  className?: string;
};

export function TextClamp({
  children,
  lines = 3,
  showMoreLabel,
  showLessLabel,
  className,
}: TextClampProps) {
  const resolvedShowMore = showMoreLabel ?? "Show more";
  const resolvedShowLess = showLessLabel ?? "Show less";
  const [expanded, setExpanded] = useState(false);
  const [clamped, setClamped] = useState(false);
  const contentRef = useRef<HTMLParagraphElement>(null);
  const lastContentRef = useRef<ReactNode>(null);

  useEffect(() => {
    if (lastContentRef.current === children) return;
    lastContentRef.current = children;
    const element = contentRef.current;
    if (!element) return;
    setClamped(element.scrollHeight > element.clientHeight + 1);
  }, [children]);

  const clampStyle: CSSProperties | undefined = expanded
    ? undefined
    : ({
        display: "-webkit-box",
        WebkitLineClamp: lines,
        WebkitBoxOrient: "vertical",
        overflow: "hidden",
      } as CSSProperties);

  return (
    <div className={cn("flex flex-col items-start gap-1", className)}>
      <p ref={contentRef} data-slot="text-clamp-content" style={clampStyle} className="text-sm">
        {children}
      </p>
      {clamped ? (
        <button
          type="button"
          onClick={() => setExpanded((value) => !value)}
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          {expanded ? resolvedShowLess : resolvedShowMore}
        </button>
      ) : null}
    </div>
  );
}
