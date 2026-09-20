import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export type MetricGridProps = {
  children: ReactNode;
  className?: string;
};

export function MetricGrid({ children, className }: MetricGridProps) {
  return (
    <div
      data-slot="metric-grid"
      className={cn(
        "grid grid-cols-[repeat(auto-fit,minmax(min(14rem,100%),1fr))] items-stretch gap-4",
        className,
      )}
    >
      {children}
    </div>
  );
}
