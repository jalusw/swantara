import React from "react";
import { cn } from "@/lib/utils";

export type BarChartDatum = {
  label: string;
  value: number;
};

export type BarChartProps = {
  data: BarChartDatum[];
  ariaLabel: string;
  valueFormatter?: (value: number) => string;
  className?: string;
};

function BarChartInner({ data, ariaLabel, valueFormatter, className }: BarChartProps) {
  const max = Math.max(...data.map((datum) => datum.value), 1);
  const dataLabel = data
    .map((datum) => {
      const value = valueFormatter ? valueFormatter(datum.value) : String(datum.value);
      return `${datum.label}: ${value}`;
    })
    .join(", ");

  return (
    <div
      data-slot="bar-chart"
      role="img"
      aria-label={`${ariaLabel}. ${dataLabel}`}
      className={cn("h-44 w-full overflow-x-auto", className)}
    >
      <div className="flex h-full min-w-[min(100%,480px)] items-end gap-1 sm:gap-2">
        {data.map((datum) => (
          <div
            key={datum.label}
            className="group/bar flex h-full min-w-0 flex-1 flex-col items-center justify-end gap-2"
          >
            <div
              title={valueFormatter ? valueFormatter(datum.value) : String(datum.value)}
              className="w-full max-w-10 rounded-md bg-primary/15 transition-colors group-hover/bar:bg-primary/30"
              style={{ height: `${(datum.value / max) * 100}%`, minHeight: 4 }}
            />
            <span className="text-xs text-muted-foreground">{datum.label}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export const BarChart = React.memo(BarChartInner);
