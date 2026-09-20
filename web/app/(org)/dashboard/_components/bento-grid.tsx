"use client";

import { cn } from "@/lib/utils";

export type BentoGridProps = React.ComponentProps<"div">;

export function BentoGrid({ className, ...props }: BentoGridProps) {
  return (
    <div data-slot="bento-grid" className={cn("grid grid-cols-12 gap-4", className)} {...props} />
  );
}

export type BentoCellProps = React.ComponentProps<"div"> & {
  span?: string;
};

export function BentoCell({ className, span, ...props }: BentoCellProps) {
  return <div data-slot="bento-cell" className={cn(span, className)} {...props} />;
}
