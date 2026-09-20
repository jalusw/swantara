"use client";

import { CheckIcon } from "lucide-react";

import { cn } from "@/lib/utils";

export type WorkflowStep = {
  label: string;
};

export type WorkflowStepsProps = {
  steps: WorkflowStep[];
  currentIndex: number;
  onStepClick?: (index: number) => void;
  className?: string;
};

export function WorkflowSteps({ steps, currentIndex, onStepClick, className }: WorkflowStepsProps) {
  return (
    <ol data-slot="workflow-steps" className={cn("flex w-full items-center", className)}>
      {steps.map((step, index) => {
        const done = index < currentIndex;
        const active = index === currentIndex;
        const isLast = index === steps.length - 1;
        return (
          <li
            key={`${step.label}-${index}`}
            className={cn("flex items-center", !isLast && "flex-1")}
          >
            <button
              type="button"
              disabled={onStepClick == null}
              aria-current={active ? "step" : undefined}
              onClick={() => onStepClick?.(index)}
              className={cn(
                "group relative flex min-h-11 items-center gap-2 rounded-full px-1.5 py-1.5 outline-none after:absolute after:-inset-1 after:content-['']",
                onStepClick != null && "cursor-pointer",
              )}
            >
              <span
                data-slot="workflow-step-indicator"
                className={cn(
                  "grid size-11 min-h-11 min-w-11 shrink-0 place-items-center rounded-full border text-sm transition-colors",
                  done && "border-transparent bg-success text-success-foreground",
                  active &&
                    "border-primary bg-primary text-primary-foreground ring-3 ring-primary/20",
                  !done && !active && "border-border bg-muted text-muted-foreground",
                )}
              >
                {done ? <CheckIcon className="size-4" aria-hidden /> : index + 1}
              </span>
              <span
                className={cn(
                  "hidden text-sm sm:block",
                  active ? "text-foreground" : "text-muted-foreground",
                )}
              >
                {step.label}
              </span>
              <span className="sr-only">
                {`${done ? "completed" : active ? "current step" : "upcoming step"}`}
              </span>
            </button>
            {!isLast ? (
              <span
                aria-hidden
                data-slot="workflow-steps-connector"
                className={cn(
                  "mx-2 h-px flex-1 transition-colors",
                  index < currentIndex ? "bg-success" : "bg-border",
                )}
              />
            ) : null}
          </li>
        );
      })}
    </ol>
  );
}
