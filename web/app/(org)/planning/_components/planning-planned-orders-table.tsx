"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/badge";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { PlanningRun } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

type PlannedSupplysTableProps = {
  run: PlanningRun;
};

export function PlannedSupplysTable({ run }: PlannedSupplysTableProps) {
  const t = useTranslations("Planning");
  const runDetailQuery = useOrgQuery<{ run: PlanningRun }>(
    "planning.runs",
    run.id,
    (organizationId) => getSwantaraService().planning.runs.get(organizationId, run.id),
  );

  const runDetail = runDetailQuery.data?.run;
  const plannedOrders = runDetail?.plannedOrders ?? [];

  return (
    <div className="flex flex-col gap-4">
      <h3 className="text-lg">{t("plannedOrders")}</h3>
      {plannedOrders.length === 0 ? (
        <div className="text-muted-foreground text-sm">{t("noPlannedOrders")}</div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 pr-4 ">{t("colItem")}</th>
                <th className="pb-2 pr-4 text-right ">{t("colDemand")}</th>
                <th className="pb-2 pr-4 text-right ">{t("colOnHand")}</th>
                <th className="pb-2 pr-4 text-right ">{t("colPlanned")}</th>
                <th className="pb-2 pr-4 ">{t("colSource")}</th>
                <th className="pb-2 ">{t("colStatus")}</th>
              </tr>
            </thead>
            <tbody>
              {plannedOrders.map((order) => (
                <tr key={order.id} className="border-b last:border-0">
                  <td className="py-2 pr-4">
                    <div className="flex flex-col">
                      <span>{`#${order.itemId}`}</span>
                      <span className="text-xs text-muted-foreground">{order.type}</span>
                    </div>
                  </td>
                  <td className="py-2 pr-4 text-right tabular-nums">{order.qty}</td>
                  <td className="py-2 pr-4 text-right tabular-nums">
                    {formatDate(order.orderDate, { nullFallback: "—" })}
                  </td>
                  <td className="py-2 pr-4 text-right tabular-nums ">
                    {formatDate(order.dueDate, { nullFallback: "—" })}
                  </td>
                  <td className="py-2 pr-4">
                    <Badge
                      variant="outline"
                      className={
                        order.confirmed
                          ? "border-success text-success"
                          : "border-muted-foreground text-muted-foreground"
                      }
                    >
                      {order.confirmed ? t("confirmed") : t("pending")}
                    </Badge>
                  </td>
                  <td className="py-2">
                    <Badge variant="outline">{order.generatedDocType ?? "—"}</Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
