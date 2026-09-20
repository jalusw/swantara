import { cn } from "@/lib/utils";
import { Badge } from "./badge";

export type StateTone = "neutral" | "success" | "warning" | "danger" | "info";

export type StatusConfig = {
  label: string;
  tone?: StateTone;
};

const dotToneClass: Record<StateTone, string> = {
  neutral: "bg-muted-foreground",
  success: "bg-success",
  warning: "bg-warning",
  danger: "bg-destructive",
  info: "bg-info",
};

export type StateBadgeProps = {
  value?: string;
  statuses?: Record<string, StatusConfig>;
  tone?: StateTone;
  label?: string;
  className?: string;
};

const toneToVariant: Record<StateTone, "default" | "destructive" | "secondary"> = {
  success: "default",
  danger: "destructive",
  neutral: "secondary",
  warning: "secondary",
  info: "secondary",
};

export function StateBadge({ value, statuses, tone, label, className }: StateBadgeProps) {
  if (tone && label) {
    return (
      <Badge
        data-slot="state-badge"
        variant={toneToVariant[tone]}
        className={cn("gap-1.5 normal-case", className)}
      >
        <span aria-hidden className="relative flex size-1.5">
          <span className={cn("relative inline-flex size-1.5 rounded-full", dotToneClass[tone])} />
        </span>
        {label}
      </Badge>
    );
  }

  if (!statuses || !value) {
    return null;
  }

  const status = statuses[value];
  if (!status) {
    return null;
  }
  return (
    <Badge
      data-slot="state-badge"
      variant="outline"
      className={cn("gap-1.5 normal-case", className)}
    >
      <span aria-hidden className="relative flex size-1.5">
        <span
          className={cn(
            "relative inline-flex size-1.5 rounded-full",
            dotToneClass[status.tone ?? "neutral"],
          )}
        />
      </span>
      {status.label}
    </Badge>
  );
}

export function StatusDot({ tone = "neutral" }: { tone?: StateTone }) {
  return (
    <span data-slot="status-dot" className={cn("size-1.5 rounded-full", dotToneClass[tone])} />
  );
}
