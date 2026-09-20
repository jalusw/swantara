"use client";

import type { ReactNode } from "react";

export function KpiCard({
  label,
  children,
  className,
}: {
  label: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <div className={className ?? "rounded-xl border border-border bg-card p-4"}>
      <p className="text-[11px] font-medium uppercase tracking-widest text-muted-foreground">
        {label}
      </p>
      <div className="mt-1 font-heading text-xl font-semibold tracking-tight tabular-nums">
        {children}
      </div>
    </div>
  );
}
