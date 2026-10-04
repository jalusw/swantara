"use client";

import { useTranslations } from "next-intl";
import { BarChart } from "@/components/bar-chart";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { formatMoney } from "@/lib/utils";

const chartTabs = {
  netProfit: [
    { label: "Jan", value: 24 },
    { label: "Feb", value: 32 },
    { label: "Mar", value: 21 },
    { label: "Apr", value: 38 },
    { label: "May", value: 29 },
    { label: "Jun", value: 45 },
    { label: "Jul", value: 33 },
    { label: "Aug", value: 51 },
    { label: "Sep", value: 40 },
    { label: "Oct", value: 56 },
    { label: "Nov", value: 44 },
    { label: "Dec", value: 61 },
  ],
  revenue: [
    { label: "Jan", value: 40 },
    { label: "Feb", value: 48 },
    { label: "Mar", value: 37 },
    { label: "Apr", value: 56 },
    { label: "May", value: 46 },
    { label: "Jun", value: 63 },
    { label: "Jul", value: 52 },
    { label: "Aug", value: 70 },
    { label: "Sep", value: 61 },
    { label: "Oct", value: 76 },
    { label: "Nov", value: 65 },
    { label: "Dec", value: 82 },
  ],
  expenses: [
    { label: "Jan", value: 16 },
    { label: "Feb", value: 16 },
    { label: "Mar", value: 16 },
    { label: "Apr", value: 18 },
    { label: "May", value: 17 },
    { label: "Jun", value: 18 },
    { label: "Jul", value: 19 },
    { label: "Aug", value: 19 },
    { label: "Sep", value: 21 },
    { label: "Oct", value: 20 },
    { label: "Nov", value: 21 },
    { label: "Dec", value: 21 },
  ],
};

export function PerformanceChart() {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Accounting");
  return (
    <Card className="lg:col-span-2">
      <CardHeader>
        <CardTitle>{t("performanceTitle")}</CardTitle>
        <CardDescription>{t("performanceDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="netProfit">
          <TabsList aria-label={t("performanceTitle")}>
            <TabsTrigger value="netProfit">{t("statNetProfit")}</TabsTrigger>
            <TabsTrigger value="revenue">{t("statRevenue")}</TabsTrigger>
            <TabsTrigger value="expenses">{t("statExpenses")}</TabsTrigger>
          </TabsList>
          {(["netProfit", "revenue", "expenses"] as const).map((key) => (
            <TabsContent key={key} value={key}>
              <BarChart
                data={chartTabs[key]}
                ariaLabel={
                  key === "netProfit"
                    ? t("statNetProfit")
                    : key === "revenue"
                      ? t("statRevenue")
                      : t("statExpenses")
                }
                valueFormatter={(value) =>
                  formatMoney(value * 1000, { currency: DEFAULT_CURRENCY })
                }
              />
            </TabsContent>
          ))}
        </Tabs>
      </CardContent>
    </Card>
  );
}
