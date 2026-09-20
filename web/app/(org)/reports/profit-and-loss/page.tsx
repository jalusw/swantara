"use client";

import { FileDown } from "lucide-react";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useProfitAndLoss } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import { ProfitAndLossRows } from "../_components/profit-and-loss-rows";

export default function ProfitAndLossPage() {
  const { data } = useProfitAndLoss();
  const pl = data?.profitAndLoss;
  const rows = pl?.rows ?? [];

  function handleExportCsv() {
    const headers = ["Code", "Account", "Type", "Amount"];
    const csvRows = rows.map((r) => [r.code, r.name, r.accountType, r.amount]);
    exportCsv("profit-and-loss", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{"Back to reports"}</BackLink>

      <PageHeader
        title={"Profit & Loss"}
        description={"Income and expenses for the current period."}
        actions={
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCsv}
            disabled={rows.length === 0}
          >
            <FileDown className="size-4" aria-hidden />
            {"Export CSV"}
          </Button>
        }
      />

      {pl ? (
        <section className="grid gap-3 sm:grid-cols-3">
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Revenue"}</p>
              <p className="font-heading text-lg font-bold text-success">
                {formatMoney(pl.revenue, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Gross profit"}</p>
              <p className="font-heading text-lg font-bold">
                {formatMoney(pl.grossProfit, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <p className="text-xs text-muted-foreground">{"Net income"}</p>
              <p
                className={cn(
                  "font-heading text-lg font-bold",
                  pl.netIncome >= 0 ? "text-success" : "text-destructive",
                )}
              >
                {formatMoney(pl.netIncome, { currency: DEFAULT_CURRENCY })}
              </p>
            </CardContent>
          </Card>
        </section>
      ) : null}

      <ProfitAndLossRows />
    </div>
  );
}
