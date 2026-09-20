"use client";

import type { LucideIcon } from "lucide-react";
import {
  AlertCircleIcon,
  CheckCircle2Icon,
  InfoIcon,
  TriangleAlertIcon,
  XIcon,
} from "lucide-react";
import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

export type CalloutVariant = "info" | "warning" | "success" | "danger";

export type CalloutProps = {
  variant?: CalloutVariant;
  title?: string;
  icon?: LucideIcon;
  onDismiss?: () => void;
  children?: ReactNode;
  className?: string;
};

const iconByVariant: Record<CalloutVariant, LucideIcon> = {
  info: InfoIcon,
  warning: TriangleAlertIcon,
  success: CheckCircle2Icon,
  danger: AlertCircleIcon,
};

const toneClass: Record<CalloutVariant, string> = {
  info: "border-info/30 bg-info/5 text-info",
  warning: "border-warning/40 bg-warning/10 text-warning",
  success: "border-success/30 bg-success/10 text-success",
  danger: "border-destructive/30 bg-destructive/10 text-destructive",
};

export function Callout({
  variant = "info",
  title,
  icon,
  onDismiss,
  children,
  className,
}: CalloutProps) {
  const Icon = icon ?? iconByVariant[variant];

  return (
    <div
      data-slot="callout"
      data-variant={variant}
      role={variant === "danger" ? "alert" : undefined}
      className={cn(
        "flex gap-3 rounded-lg border px-3 py-2.5 text-sm",
        toneClass[variant],
        className,
      )}
    >
      <Icon className="mt-0.5 size-4 shrink-0" aria-hidden />
      <div className="flex-1 space-y-1">
        {title ? <p className=" text-foreground">{title}</p> : null}
        {children ? <div className="text-foreground/80">{children}</div> : null}
      </div>
      {onDismiss ? (
        <button
          type="button"
          data-slot="callout-dismiss"
          aria-label={"Dismiss"}
          onClick={onDismiss}
          className="relative grid size-11 min-h-11 min-w-11 shrink-0 place-items-center rounded-md opacity-60 transition-opacity outline-none hover:opacity-100 focus-visible:ring-2 ring-offset-2 ring-offset-background focus-visible:ring-ring"
        >
          <XIcon className="size-3.5" aria-hidden />
        </button>
      ) : null}
    </div>
  );
}
