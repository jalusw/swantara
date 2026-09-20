"use client";

import { FileDown } from "lucide-react";

import { BackLink } from "@/components/back-link";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useAgingReport } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { agingBucketSortKey, agingTotalByType, groupAgingByType } from "@/lib/utils/report-utils";

export default function AgingReportPage() {
  const { data, isLoading } = useAgingReport();
  const rows = data?.rows ?? [];
  const grouped = groupAgingByType(rows);
  const grandTotal = agingTotalByType(rows);

  function handleExportCsv() {
    const headers = ["Type", "Age bucket", "Currency", "Amount"];
    const csvRows = rows.map((r) => [r.type, r.bucket, r.currencyCode, r.amount]);
    exportCsv("aging-report", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{"Back to reports"}</BackLink>

      <PageHeader
        title={"Accounts Aging"}
        description={"Outstanding receivables and payables by age bucket."}
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

      <Card>
        <CardContent className="p-4">
          <p className="text-xs text-muted-foreground">{"Grand total"}</p>
          <p className="font-heading text-lg font-bold">
            {formatMoney(grandTotal, { currency: DEFAULT_CURRENCY })}
          </p>
        </CardContent>
      </Card>

      {isLoading ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {"Loading aging report…"}
          </CardContent>
        </Card>
      ) : rows.length === 0 ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {"No aging data available."}
          </CardContent>
        </Card>
      ) : (
        <div className="flex flex-col gap-4">
          {Array.from(grouped.entries()).map(([type, typeRows]) => (
            <Card key={type}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>{type}</CardTitle>
                  <Badge variant="secondary">
                    {formatMoney(agingTotalByType(typeRows), { currency: DEFAULT_CURRENCY })}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b border-border">
                        <th className="px-3 py-2 text-left text-muted-foreground">
                          {"Age bucket"}
                        </th>
                        <th className="px-3 py-2 text-left text-muted-foreground">{"Currency"}</th>
                        <th className="px-3 py-2 text-right text-muted-foreground">{"Amount"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {[...typeRows]
                        .sort((a, b) => agingBucketSortKey(a.bucket) - agingBucketSortKey(b.bucket))
                        .map((row) => (
                          <tr
                            key={`${row.type}-${row.bucket}-${row.currencyCode}`}
                            className="border-b border-border/50"
                          >
                            <td className="px-3 py-2">
                              <Badge variant="outline">{row.bucket}</Badge>
                            </td>
                            <td className="px-3 py-2 font-mono text-xs">{row.currencyCode}</td>
                            <td className="px-3 py-2 text-right tabular-nums ">
                              {formatMoney(row.amount, { currency: DEFAULT_CURRENCY })}
                            </td>
                          </tr>
                        ))}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
