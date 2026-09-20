import { cn } from "@/lib/utils";

export type MultiProgressSegment = {
  id: string;
  label: string;
  /** Relative weight of this segment; the bar uses segment value / total. */
  value: number;
  /** CSS color; defaults to the shared segment palette. */
  color?: string;
};

export type MultiProgressProps = {
  segments: MultiProgressSegment[];
  ariaLabel: string;
  className?: string;
};

const DEFAULT_COLORS = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

export function MultiProgress({ segments, ariaLabel, className }: MultiProgressProps) {
  const total = segments.reduce((sum, segment) => sum + segment.value, 0);
  const hasSegments = total > 0;

  return (
    <fieldset
      data-slot="multi-progress"
      aria-label={ariaLabel}
      className={cn("m-0 space-y-2 border-0 p-0", className)}
    >
      <legend className="sr-only">{ariaLabel}</legend>
      <div className="flex h-2.5 w-full overflow-hidden rounded-full bg-muted" aria-hidden="true">
        {hasSegments
          ? segments.map((segment, index) => {
              const width = (segment.value / total) * 100;
              return (
                <div
                  key={segment.id}
                  style={{
                    width: `${width}%`,
                    backgroundColor: segment.color ?? DEFAULT_COLORS[index % DEFAULT_COLORS.length],
                  }}
                />
              );
            })
          : null}
      </div>
      <ul className="flex flex-wrap gap-x-4 gap-y-1">
        {segments.map((segment, index) => {
          const percent = hasSegments ? (segment.value / total) * 100 : 0;
          return (
            <li key={segment.id} className="flex items-center gap-1.5 text-sm">
              <span
                className="size-2 rounded-full"
                aria-hidden="true"
                style={{
                  backgroundColor: segment.color ?? DEFAULT_COLORS[index % DEFAULT_COLORS.length],
                }}
              />
              <span className="text-muted-foreground">{segment.label}</span>
              <span className=" tabular-nums">{Math.round(percent)}%</span>
            </li>
          );
        })}
      </ul>
    </fieldset>
  );
}
