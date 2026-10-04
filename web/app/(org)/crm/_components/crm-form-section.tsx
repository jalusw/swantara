"use client";

import { Separator } from "@/components/separator";
import { cn } from "@/lib/utils";

export function CrmFormSection({
  title,
  description,
  children,
  className,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={cn("flex flex-col gap-3 py-4 first:pt-1 last:pb-0", className)}>
      <div className="flex flex-col gap-0.5">
        <h3 className="text-sm font-semibold text-foreground">{title}</h3>
        {description ? <p className="text-xs text-muted-foreground">{description}</p> : null}
      </div>
      <div className="grid gap-4 sm:grid-cols-2">{children}</div>
    </section>
  );
}

export function CrmFormDivider() {
  return <Separator className="my-1" />;
}
