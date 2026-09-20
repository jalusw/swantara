import type { LucideIcon } from "lucide-react";
import { CircleDotIcon } from "lucide-react";
import { cn } from "@/lib/utils";
import type { StateTone } from "./state-badge";

export type ActivityItem = {
  id: string;
  icon?: LucideIcon;
  title: string;
  description?: string;
  timestamp?: string;
  tone?: StateTone;
};

export type ActivityTimelineProps = {
  items: ActivityItem[];
  className?: string;
};

const toneClass: Record<StateTone, string> = {
  neutral: "bg-muted-foreground",
  success: "bg-success",
  warning: "bg-warning",
  danger: "bg-destructive",
  info: "bg-info",
};

export function ActivityTimeline({ items, className }: ActivityTimelineProps) {
  return (
    <ol data-slot="activity-timeline" className={cn("flex flex-col", className)}>
      {items.map((item, index) => {
        const Icon = item.icon ?? CircleDotIcon;
        const isLast = index === items.length - 1;
        return (
          <li key={item.id} className="relative flex gap-3 pb-5 last:pb-0">
            <div className="flex flex-col items-center">
              <span
                className={cn(
                  "flex size-6 shrink-0 items-center justify-center rounded-full ring-1 ring-border",
                  toneClass[item.tone ?? "neutral"],
                )}
              >
                <Icon className="size-3 text-background" aria-hidden />
              </span>
              {!isLast ? (
                <span
                  aria-hidden
                  className="w-px flex-1 bg-border"
                  data-slot="activity-timeline-line"
                />
              ) : null}
            </div>
            <div className="flex flex-1 flex-col gap-0.5 pt-0.5">
              <div className="flex items-baseline justify-between gap-2">
                <p className="text-sm text-foreground">
                  {item.tone && item.tone !== "neutral" ? (
                    <span className="sr-only">{item.tone}: </span>
                  ) : null}
                  {item.title}
                </p>
                {item.timestamp ? (
                  <time className="shrink-0 text-xs text-muted-foreground">{item.timestamp}</time>
                ) : null}
              </div>
              {item.description ? (
                <p className="text-sm text-muted-foreground">{item.description}</p>
              ) : null}
            </div>
          </li>
        );
      })}
    </ol>
  );
}
