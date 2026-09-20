"use client";

import { FileDown } from "lucide-react";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useCashFlowReport } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";

export default function CashFlowPage() {
  const { data, isLoading } = useCashFlowReport();
  const cf = data?.cashFlow;

  function handleExportCsv() {
    if (!cf) return;
    const headers = [
      "Operating",
      "Investing",
      "Financing",
      "Net change",
      "Opening cash",
      "Closing cash",
    ];
    const csvRows = [
      [cf.operating, cf.investing, cf.financing, cf.netChange, cf.openingCash, cf.closingCash],
    ];
    exportCsv("cash-flow", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{"Back to reports"}</BackLink>

      <PageHeader
        title={"Cash Flow"}
        description={"Cash movements across operating, investing, and financing."}
        actions={
          <Button variant="outline" size="sm" onClick={handleExportCsv} disabled={!cf}>
            <FileDown className="size-4" aria-hidden />
            {"Export CSV"}
          </Button>
        }
      />

      {isLoading ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {"Loading cash flow…"}
          </CardContent>
        </Card>
      ) : !cf ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {"No cash flow data available."}
          </CardContent>
        </Card>
      ) : (
        <>
          <section className="grid gap-3 sm:grid-cols-3">
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Operating"}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.operating >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.operating, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Investing"}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.investing >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.investing, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Financing"}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.financing >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.financing, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
          </section>

          <section className="grid gap-3 sm:grid-cols-3">
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Net change"}</p>
                <p
                  className={cn(
                    "font-heading text-lg font-bold",
                    cf.netChange >= 0 ? "text-success" : "text-destructive",
                  )}
                >
                  {formatMoney(cf.netChange, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Opening cash"}</p>
                <p className="font-heading text-lg font-bold">
                  {formatMoney(cf.openingCash, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardContent className="p-4">
                <p className="text-xs text-muted-foreground">{"Closing cash"}</p>
                <p className="font-heading text-lg font-bold">
                  {formatMoney(cf.closingCash, { currency: DEFAULT_CURRENCY })}
                </p>
              </CardContent>
            </Card>
          </section>
        </>
      )}
    </div>
  );
}
