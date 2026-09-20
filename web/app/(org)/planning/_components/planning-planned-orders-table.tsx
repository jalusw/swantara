"use client";

import { Badge } from "@/components/badge";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { PlanningRun } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

type PlannedSupplysTableProps = {
  run: PlanningRun;
};

export function PlannedSupplysTable({ run }: PlannedSupplysTableProps) {
  const runDetailQuery = useOrgQuery<{ run: PlanningRun }>(
    "planning.runs",
    run.id,
    (organizationId) => getSwantaraService().planning.runs.get(organizationId, run.id),
  );

  const runDetail = runDetailQuery.data?.run;
  const plannedOrders = runDetail?.plannedOrders ?? [];

  return (
    <div className="flex flex-col gap-4">
      <h3 className="text-lg">{"Planned Orders"}</h3>
      {plannedOrders.length === 0 ? (
        <div className="text-muted-foreground text-sm">{"No planned orders in this run"}</div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-left text-muted-foreground">
                <th className="pb-2 pr-4 ">{"Item"}</th>
                <th className="pb-2 pr-4 text-right ">{"Demand"}</th>
                <th className="pb-2 pr-4 text-right ">{"On Hand"}</th>
                <th className="pb-2 pr-4 text-right ">{"Planned"}</th>
                <th className="pb-2 pr-4 ">{"Source"}</th>
                <th className="pb-2 ">{"Status"}</th>
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
                      {order.confirmed ? "Confirmed" : "Pending"}
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
