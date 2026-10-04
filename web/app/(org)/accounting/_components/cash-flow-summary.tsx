"use client";

import { ArrowDownRight, ArrowUpRight, PiggyBank } from "lucide-react";
import { useTranslations } from "next-intl";
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
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Accounting");
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("cashFlowTitle")}</CardTitle>
        <CardDescription>{t("cashFlowDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <ul className="space-y-3">
          <CashFlowItem label={t("cashFlowInflows")} value="—" />
          <CashFlowItem label={t("cashFlowOutflows")} value="—" />
          <CashFlowItem label={t("cashFlowNet")} value="—" />
        </ul>
      </CardContent>
    </Card>
  );
}
