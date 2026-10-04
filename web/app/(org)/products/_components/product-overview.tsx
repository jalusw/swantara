"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { Card, CardContent } from "@/components/card";
import type { Item } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function ProductOverview({ item }: { item: Item }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const general = [
    {
      label: t("fieldType"),
      value: (t as unknown as (k: string) => string)(`productType_${item.type}`),
    },
    {
      label: t("fieldTracking"),
      value: (t as unknown as (k: string) => string)(`trackingMode_${item.tracking}`),
    },
    {
      label: t("fieldUnit"),
      value: item.unitId ? String(item.unitId) : "—",
    },
  ];
  const pricing = [
    { label: t("colListPrice"), value: formatNumber(item.listPrice) },
    {
      label: t("colStandardCost"),
      value: formatNumber(item.standardCost),
    },
  ];

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardContent>
          <Section title={t("general")} items={general} />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Section title={t("pricing")} items={pricing} />
          <div className="mt-4 flex flex-col gap-2 border-t pt-4">
            <p className="text-sm">{t("salesPurchase")}</p>
            <div className="flex flex-wrap gap-2">
              {item.isSellable ? <Badge>{t("canBeSold")}</Badge> : null}
              {item.isPurchasable ? <Badge variant="secondary">{t("canBePurchased")}</Badge> : null}
              {item.isManufactured ? <Badge variant="outline">{t("isManufactured")}</Badge> : null}
              {!item.isSellable && !item.isPurchasable && !item.isManufactured ? (
                <span className="text-sm text-muted-foreground">—</span>
              ) : null}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function Section({ title, items }: { title: string; items: { label: string; value: string }[] }) {
  return (
    <dl className="grid gap-3 sm:grid-cols-2">
      <p className="text-sm sm:col-span-2">{title}</p>
      {items.map((item) => (
        <div key={item.label} className="flex flex-col gap-0.5">
          <dt className="text-xs text-muted-foreground">{item.label}</dt>
          <dd className="text-sm">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
