"use client";

import { ArrowDownRight, ArrowUpRight, PiggyBank } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { cn } from "@/lib/utils/style";

const cashFlowIconMap: Record<string, typeof ArrowUpRight> = {
  success: ArrowUpRight,
  destructive: ArrowDownRight,
};
const cashFlowColorMap: Record<string, string> = {
  success: "text-success",
  destructive: "text-destructive",
};

function CashFlowItem({
  label,
  value,
  tone = "neutral",
}: {
  label: string;
  value: string;
  tone?: "success" | "destructive" | "neutral";
}) {
  const Icon = cashFlowIconMap[tone] ?? PiggyBank;
  const color = cashFlowColorMap[tone] ?? "text-foreground";
  return (
    <li className="flex items-center justify-between">
      <span className="flex items-center gap-2 text-sm text-muted-foreground">
        <Icon className={cn("size-4", color)} aria-hidden />
        {label}
      </span>
      <span className={cn("text-sm", color)}>{value}</span>
    </li>
  );
}

export function CashFlowSummary() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Cash flow"}</CardTitle>
        <CardDescription>{"Money in and out this month."}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="space-y-3">
          <CashFlowItem label={"Inflows"} value="—" />
          <CashFlowItem label={"Outflows"} value="—" />
          <CashFlowItem label={"Net"} value="—" />
        </ul>
      </CardContent>
    </Card>
  );
}
