import { ChevronRightIcon, HomeIcon } from "lucide-react";
import React from "react";

import { cn } from "@/lib/utils";

export type BreadcrumbItem = {
  label: string;
  href?: string;
};

export type BreadcrumbProps = {
  items: BreadcrumbItem[];
  className?: string;
};

function BreadcrumbInner({ items, className }: BreadcrumbProps) {
  const lastIndex = items.length - 1;

  return (
    <nav
      aria-label={"Breadcrumb"}
      data-slot="breadcrumb"
      className={cn("flex items-center gap-1 text-sm", className)}
    >
      {items.map((item, index) => {
        const isLast = index === lastIndex;
        const isFirst = index === 0;
        const content = (
          <>
            {isFirst ? <HomeIcon className="size-3.5" aria-hidden /> : null}
            {item.label}
          </>
        );
        return (
          <span key={`${item.label}-${index}`} className="flex items-center gap-1">
            {!isFirst ? (
              <ChevronRightIcon className="size-3.5 text-muted-foreground/50" aria-hidden />
            ) : null}
            {isLast || !item.href ? (
              <span
                aria-current={isLast ? "page" : undefined}
                className={cn(
                  "flex items-center gap-1.5",
                  isLast ? " text-foreground" : "text-muted-foreground",
                )}
              >
                {content}
              </span>
            ) : (
              <a
                href={item.href}
                className="relative flex min-h-11 items-center gap-1.5 rounded-sm px-1.5 py-2 text-muted-foreground transition-colors outline-none after:absolute after:-inset-x-0.5 after:-inset-y-1 after:content-[''] hover:text-foreground focus-visible:ring-3 ring-offset-2 ring-offset-background focus-visible:ring-ring"
              >
                {content}
              </a>
            )}
          </span>
        );
      })}
    </nav>
  );
}

export const Breadcrumb = React.memo(BreadcrumbInner);
