import type { LucideIcon } from "lucide-react";
import { InboxIcon } from "lucide-react";
import type * as React from "react";

import { cn } from "@/lib/utils";

export type EmptyStateProps = {
  title: string;
  description?: string;
  icon?: LucideIcon;
  headingLevel?: 2 | 3 | 4 | 5 | 6;
  children?: React.ReactNode;
  className?: string;
};

export function EmptyState({
  title,
  description,
  icon: Icon = InboxIcon,
  headingLevel = 3,
  children,
  className,
}: EmptyStateProps) {
  const Heading = `h${headingLevel}` as "h2" | "h3";
  return (
    <div
      data-slot="empty-state"
      className={cn(
        "flex w-full flex-col items-center justify-center gap-2 px-6 py-16 text-center",
        className,
      )}
    >
      <span className="grid size-12 place-items-center rounded-xl bg-muted text-muted-foreground ring-1 ring-border">
        <Icon className="size-6" aria-hidden />
      </span>
      <div className="space-y-1">
        <Heading className="font-heading text-base text-balance">{title}</Heading>
        {description ? (
          <p className="max-w-sm text-sm text-muted-foreground text-pretty">{description}</p>
        ) : null}
      </div>
      {children ? <div className="mt-2 flex items-center gap-2">{children}</div> : null}
    </div>
  );
}
