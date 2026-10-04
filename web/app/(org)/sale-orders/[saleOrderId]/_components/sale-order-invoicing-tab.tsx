"use client";

import { useTranslations } from "next-intl";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import type { SaleOrderLine } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type SaleOrderInvoicingTabProps = {
  deliveredUnbilledLines: SaleOrderLine[];
  onInvoice: () => void;
};

export function SaleOrderInvoicingTab({
  deliveredUnbilledLines,
  onInvoice,
}: SaleOrderInvoicingTabProps) {
  const t = useTranslations("Sales");
  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t("tabInvoicing")}</CardTitle>
          <p className="text-sm text-muted-foreground">{t("invoiceDescription")}</p>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {deliveredUnbilledLines.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t("noBillableLines")}</p>
          ) : (
            <>
              <p className="text-sm">
                {t("linesToInvoice")}: {deliveredUnbilledLines.length}
              </p>
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-muted-foreground">
                      <th className="pb-2 text-left">{t("fieldItem")}</th>
                      <th className="pb-2 text-right">{t("deliveredNotInvoiced")}</th>
                      <th className="pb-2 text-right">{t("unitPrice")}</th>
                    </tr>
                  </thead>
                  <tbody>
                    {deliveredUnbilledLines.map((line) => (
                      <tr key={line.id} className="border-b">
                        <td className="py-2">{line.itemId}</td>
                        <td className="py-2 text-right tabular-nums">
                          {formatNumber(line.qtyDelivered - line.qtyInvoiced)}
                        </td>
                        <td className="py-2 text-right tabular-nums">
                          {formatNumber(line.unitPrice)}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </>
          )}
          <Button size="sm" onClick={onInvoice} disabled={deliveredUnbilledLines.length === 0}>
            {t("invoiceTitle")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
