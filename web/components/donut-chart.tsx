"use client";

import React from "react";
import { cn } from "@/lib/utils";

export type DonutChartDatum = {
  label: string;
  value: number;
};

export type DonutChartProps = {
  data: DonutChartDatum[];
  ariaLabel: string;
  valueFormatter?: (value: number) => string;
  centerLabel?: string;
  showLegend?: boolean;
  thickness?: number;
  className?: string;
};

const chartColors = [
  "var(--chart-1)",
  "var(--chart-2)",
  "var(--chart-3)",
  "var(--chart-4)",
  "var(--chart-5)",
];

function formatValue(value: number, formatter?: (value: number) => string) {
  return formatter ? formatter(value) : String(value);
}

function DonutChartInner({
  data,
  ariaLabel,
  valueFormatter,
  centerLabel,
  showLegend = true,
  thickness = 14,
  className,
}: DonutChartProps) {
  const resolvedCenterLabel = centerLabel ?? "Total";
  const total = data.reduce((sum, datum) => sum + Math.max(0, datum.value), 0);

  const dataLabel = data
    .map((datum) => `${datum.label}: ${formatValue(datum.value, valueFormatter)}`)
    .join(", ");

  if (data.length === 0) {
    return (
      <div
        data-slot="donut-chart"
        role="img"
        aria-label={`${ariaLabel}. ${"No data"}`}
        className={cn("grid h-44 place-items-center", className)}
      >
        <span className="text-sm text-muted-foreground">{"No data"}</span>
      </div>
    );
  }

  if (total <= 0) {
    return (
      <div data-slot="donut-chart" className={cn("w-full", className)}>
        <div
          role="img"
          aria-label={`${ariaLabel}. ${dataLabel}`}
          className="flex items-center justify-center"
        >
          <div className="relative">
            <svg viewBox="0 0 100 100" className="size-44" aria-hidden="true">
              <circle
                cx="50"
                cy="50"
                r="42"
                fill="none"
                stroke="var(--muted)"
                strokeWidth={thickness}
              />
            </svg>
            <div className="absolute inset-0 grid place-items-center">
              <div className="flex flex-col items-center text-center">
                <span className="text-xs text-muted-foreground">{resolvedCenterLabel}</span>
                <span className="text-lg tabular-nums">{formatValue(0, valueFormatter)}</span>
              </div>
            </div>
          </div>
        </div>
        {showLegend ? (
          <ul className="mt-3 flex flex-wrap justify-center gap-x-4 gap-y-1.5">
            {data.map((datum, index) => (
              <li
                key={datum.label}
                className="flex items-center gap-1.5 text-xs text-muted-foreground"
              >
                <span
                  aria-hidden
                  className="size-2 rounded-full"
                  style={{
                    backgroundColor: chartColors[index % chartColors.length],
                  }}
                />
                <span className=" text-foreground">{datum.label}</span>
                <span className="tabular-nums">{formatValue(datum.value, valueFormatter)}</span>
                <span className="tabular-nums">(0%)</span>
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    );
  }

  let accumulated = 0;

  return (
    <div data-slot="donut-chart" className={cn("w-full", className)}>
      <div
        role="img"
        aria-label={`${ariaLabel}. ${dataLabel}`}
        className="flex items-center justify-center"
      >
        <div className="relative">
          <svg viewBox="0 0 100 100" className="size-44" aria-hidden="true">
            <g transform="rotate(-90 50 50)">
              {data.map((datum, index) => {
                const fraction = Math.max(0, datum.value) / total;
                const start = accumulated;
                accumulated += fraction;
                return (
                  <circle
                    key={datum.label}
                    cx="50"
                    cy="50"
                    r="42"
                    fill="none"
                    pathLength="100"
                    strokeWidth={thickness}
                    stroke={chartColors[index % chartColors.length]}
                    strokeDasharray={`${fraction * 100} ${100 - fraction * 100}`}
                    strokeDashoffset={-start * 100}
                  />
                );
              })}
            </g>
          </svg>
          <div className="absolute inset-0 grid place-items-center">
            <div className="flex flex-col items-center text-center">
              <span className="text-xs text-muted-foreground">{resolvedCenterLabel}</span>
              <span className="text-lg tabular-nums">{formatValue(total, valueFormatter)}</span>
            </div>
          </div>
        </div>
      </div>

      {showLegend ? (
        <ul className="mt-3 flex flex-wrap justify-center gap-x-4 gap-y-1.5">
          {data.map((datum, index) => {
            const fraction = total > 0 ? Math.max(0, datum.value) / total : 0;
            return (
              <li
                key={datum.label}
                className="flex items-center gap-1.5 text-xs text-muted-foreground"
              >
                <span
                  aria-hidden
                  className="size-2 rounded-full"
                  style={{
                    backgroundColor: chartColors[index % chartColors.length],
                  }}
                />
                <span className=" text-foreground">{datum.label}</span>
                <span className="tabular-nums">{formatValue(datum.value, valueFormatter)}</span>
                <span className="tabular-nums">({Math.round(fraction * 100)}%)</span>
              </li>
            );
          })}
        </ul>
      ) : null}
    </div>
  );
}

export const DonutChart = React.memo(DonutChartInner);
