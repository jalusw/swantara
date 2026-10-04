"use client";

import { FileDown } from "lucide-react";
import Link from "next/link";
import { useTranslations } from "next-intl";

import { BackLink } from "@/components/back-link";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { PageHeader } from "@/components/page-header";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useInventoryValuation } from "@/lib/hooks/use-report-queries";
import { exportCsv, formatMoney } from "@/lib/utils";
import { inventoryTotalQuantity, inventoryTotalValue } from "@/lib/utils/report-utils";

export default function InventoryValuationPage() {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Reports");
  const { data, isLoading } = useInventoryValuation();
  const rows = data?.rows ?? [];
  const totalValue = inventoryTotalValue(rows);
  const totalQty = inventoryTotalQuantity(rows);

  function handleExportCsv() {
    const headers = [t("fieldItem"), t("colQuantity"), t("colValue")];
    const csvRows = rows.map((r) => [r.productName, r.quantity, r.value]);
    exportCsv("inventory-valuation", headers, csvRows);
  }

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/reports"}>{t("backToReports")}</BackLink>

      <PageHeader
        title={t("inventoryValuationTitle")}
        description={t("inventoryValuationDescription")}
        actions={
          <Button
            variant="outline"
            size="sm"
            onClick={handleExportCsv}
            disabled={rows.length === 0}
          >
            <FileDown className="size-4" aria-hidden />
            {t("exportCsv")}
          </Button>
        }
      />

      <section className="grid gap-3 sm:grid-cols-2">
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{t("totalValue")}</p>
            <p className="font-heading text-lg font-bold">
              {formatMoney(totalValue, { currency: DEFAULT_CURRENCY })}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardContent className="p-4">
            <p className="text-xs text-muted-foreground">{t("totalQuantity")}</p>
            <p className="font-heading text-lg font-bold">
              {formatMoney(totalQty, { currency: DEFAULT_CURRENCY })}
            </p>
          </CardContent>
        </Card>
      </section>

      {isLoading ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {t("loadingInventoryValuation")}
          </CardContent>
        </Card>
      ) : rows.length === 0 ? (
        <Card>
          <CardContent className="p-6 text-center text-muted-foreground">
            {t("inventoryValuationEmpty")}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>{t("inventoryValues")}</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border">
                    <th className="px-3 py-2 text-left text-muted-foreground">{t("fieldItem")}</th>
                    <th className="px-3 py-2 text-right text-muted-foreground">
                      {t("colQuantity")}
                    </th>
                    <th className="px-3 py-2 text-right text-muted-foreground">{t("colValue")}</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((row) => (
                    <tr key={row.itemId} className="border-b border-border/50">
                      <td className="px-3 py-2">
                        <Link
                          href={`/products/${row.itemId}`}
                          className="text-primary hover:underline"
                        >
                          {row.productName}
                        </Link>
                      </td>
                      <td className="px-3 py-2 text-right tabular-nums">
                        {formatMoney(row.quantity, { currency: DEFAULT_CURRENCY })}
                      </td>
                      <td className="px-3 py-2 text-right tabular-nums ">
                        {formatMoney(row.value, { currency: DEFAULT_CURRENCY })}
                      </td>
                    </tr>
                  ))}
                </tbody>
                <tfoot>
                  <tr className="border-t-2 border-border ">
                    <td className="px-3 py-2">{t("colTotal")}</td>
                    <td className="px-3 py-2 text-right tabular-nums">
                      {formatMoney(totalQty, { currency: DEFAULT_CURRENCY })}
                    </td>
                    <td className="px-3 py-2 text-right tabular-nums">
                      {formatMoney(totalValue, { currency: DEFAULT_CURRENCY })}
                    </td>
                  </tr>
                </tfoot>
              </table>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
