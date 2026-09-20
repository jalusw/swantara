"use client";

import { Badge } from "@/components/badge";
import { Card, CardContent } from "@/components/card";
import type { Item } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";

export function ProductOverview({ item }: { item: Item }) {
  const general = [
    { label: "Type", value: humanizeKey(String(item.type)) },
    {
      label: "Tracking",
      value: humanizeKey(String(item.tracking)),
    },
    {
      label: "Unit",
      value: item.unitId ? String(item.unitId) : "—",
    },
  ];
  const pricing = [
    { label: "Sales price", value: formatNumber(item.listPrice) },
    {
      label: "Standard Cost",
      value: formatNumber(item.standardCost),
    },
  ];

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardContent>
          <Section title={"General"} items={general} />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Section title={"Pricing"} items={pricing} />
          <div className="mt-4 flex flex-col gap-2 border-t pt-4">
            <p className="text-sm">{"Sales & purchase"}</p>
            <div className="flex flex-wrap gap-2">
              {item.isSellable ? <Badge>{"Is Sellable"}</Badge> : null}
              {item.isPurchasable ? <Badge variant="secondary">{"Is Purchasable"}</Badge> : null}
              {item.isManufactured ? <Badge variant="outline">{"Is Manufactured"}</Badge> : null}
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
