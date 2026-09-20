"use client";

import { CheckIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type ShipmentStepState = "done" | "active" | "pending";

export type ShipmentStep = {
  id: string;
  label: string;
  description?: string;
  timestamp?: string;
  state?: ShipmentStepState;
};

export type ShipmentTrackProps = {
  steps: ShipmentStep[];
  ariaLabel?: string;
  className?: string;
};

const DOT_STATE_CLASS: Record<ShipmentStepState, string> = {
  done: "border-success bg-success text-success-foreground",
  active: "border-primary bg-primary text-primary-foreground",
  pending: "border-border bg-background text-muted-foreground",
};

function progressFraction(steps: ShipmentStep[]): number {
  if (steps.length <= 1) {
    return 0;
  }
  const active = steps.findIndex((step) => step.state === "active");
  const done = steps.filter((step) => step.state === "done").length;
  const index = active >= 0 ? active : Math.min(done, steps.length);
  return Math.min(1, index / (steps.length - 1));
}

export function ShipmentTrack({
  steps,
  ariaLabel = "Shipment progress",
  className,
}: ShipmentTrackProps) {
  const fraction = progressFraction(steps);

  return (
    <section
      data-slot="shipment-track"
      aria-label={ariaLabel}
      className={cn("overflow-x-auto py-2", className)}
    >
      <div className="relative min-w-[42rem] px-16 pb-2">
        <div aria-hidden className="absolute top-3 h-1.5 w-full rounded-full bg-muted" />
        <div
          data-slot="shipment-track-progress"
          aria-hidden
          className="absolute top-3 h-1.5 rounded-full bg-primary motion-safe:transition-[width] motion-safe:duration-700 motion-safe:ease-out"
          style={{ width: `${fraction * 100}%` }}
        />
        <ol className="relative">
          {steps.map((step, index) => {
            const state = step.state ?? "pending";
            const left = `${steps.length <= 1 ? 0 : (index / (steps.length - 1)) * 100}%`;
            return (
              <li
                key={step.id}
                className="absolute top-0 -translate-x-1/2 text-center"
                style={{ left }}
              >
                <span
                  aria-hidden
                  className={cn(
                    "relative mx-auto grid size-6 place-items-center rounded-full border-2 shadow-sm",
                    DOT_STATE_CLASS[state],
                  )}
                >
                  {state === "done" ? <CheckIcon className="size-3.5" /> : null}
                </span>
                <span className="sr-only">
                  {step.label} — {state}
                </span>
                <div className="mt-2 w-32">
                  <p className="text-sm leading-tight text-foreground">{step.label}</p>
                  {step.description ? (
                    <p className="mt-0.5 text-xs text-muted-foreground">{step.description}</p>
                  ) : null}
                  {step.timestamp ? (
                    <p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
                      <time>{step.timestamp}</time>
                    </p>
                  ) : null}
                </div>
              </li>
            );
          })}
        </ol>
      </div>
    </section>
  );
}
