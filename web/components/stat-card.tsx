import type { LucideIcon } from "lucide-react";
import { ArrowDownRight, ArrowUpRight } from "lucide-react";
import React from "react";
import { cn } from "@/lib/utils";
import { Card, CardContent, CardHeader, CardTitle } from "./card";

export type StatCardProps = {
  label: string;
  value: string;
  icon: LucideIcon;
  trend?: string;
  trendDirection?: "up" | "down";
  className?: string;
};

function StatCardInner({
  label,
  value,
  icon: Icon,
  trend,
  trendDirection = "up",
  className,
}: StatCardProps) {
  return (
    <Card data-slot="stat-card" size="sm" className={cn(className)}>
      <CardHeader className="flex-row items-start justify-between gap-2">
        <CardTitle className="text-base text-foreground">{label}</CardTitle>
        <span className="grid size-9 shrink-0 place-items-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/15">
          <Icon className="size-5" aria-hidden />
        </span>
      </CardHeader>
      <CardContent>
        <p className="font-heading text-2xl tracking-tight sm:text-3xl">{value}</p>
        {trend ? (
          <p className="mt-1 flex items-center gap-1 text-sm text-foreground">
            {trendDirection === "up" ? (
              <ArrowUpRight className="size-3.5 text-success" aria-hidden />
            ) : (
              <ArrowDownRight className="size-3.5 text-destructive" aria-hidden />
            )}
            <span>{trend}</span>
          </p>
        ) : null}
      </CardContent>
    </Card>
  );
}

export const StatCard = React.memo(StatCardInner);
