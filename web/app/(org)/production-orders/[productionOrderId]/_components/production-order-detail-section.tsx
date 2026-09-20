"use client";

import { ArrowLeft, CheckCircle2, Play, XCircle } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Progress } from "@/components/progress";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { MOComponent, ProductionOrder, ShopTask } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDateTime } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import {
  canCancelMo,
  canConfirmMo,
  canPlanMo,
  canProduceMo,
  canStartMo,
  componentConsumptionProgress,
  type MoState,
  moProgress,
  moProgressColor,
  moStateLabel,
  moStateTone,
} from "../../_components/production-order-utils";
import { ConsumeDialog } from "./consume-dialog";
import { ProduceDialog } from "./produce-dialog";

export function MoDetailSection({
  orgId,
  productionOrderId,
}: {
  orgId: string;
  productionOrderId: string;
}) {
  const [consumeOpen, setConsumeOpen] = useState(false);
  const [produceOpen, setProduceOpen] = useState(false);

  const moQuery = useOrgQuery<{ productionOrder: ProductionOrder }>(
    "productionOrders",
    productionOrderId,
    (organizationId) =>
      getSwantaraService().productionOrders.get(organizationId, Number(productionOrderId)),
  );

  const componentsQuery = useOrgQuery<{ components: MOComponent[] }>(
    "productionOrders",
    `${productionOrderId}-components`,
    (organizationId) =>
      getSwantaraService().productionOrders.components(organizationId, Number(productionOrderId)),
  );

  const shopTasksQuery = useOrgQuery<{ shopTasks: ShopTask[] }>(
    "productionOrders",
    `${productionOrderId}-shopTasks`,
    (organizationId) =>
      getSwantaraService().productionOrders.shopTasks.list(
        organizationId,
        Number(productionOrderId),
      ),
  );

  const productionOrder = moQuery.data?.productionOrder;
  const components = componentsQuery.data?.components ?? [];
  const shopTasks = shopTasksQuery.data?.shopTasks ?? [];

  const state = (productionOrder?.state ?? "draft") as MoState;
  const progress = productionOrder ? moProgress(productionOrder) : 0;

  function handleAction(action: "confirm" | "plan" | "start" | "cancel") {
    const service = getSwantaraService().productionOrders;
    const numOrgId = Number(orgId);
    const numProductionOrderId = Number(productionOrderId);

    let promise: Promise<unknown>;
    switch (action) {
      case "confirm":
        promise = service.confirm(numOrgId, numProductionOrderId);
        break;
      case "plan":
        promise = service.plan(numOrgId, numProductionOrderId);
        break;
      case "start":
        promise = service.start(numOrgId, numProductionOrderId);
        break;
      case "cancel":
        promise = service.cancel(numOrgId, numProductionOrderId);
        break;
    }

    void promise.then(() => {
      toast.success(String(action));
      void moQuery.refetch();
    });
  }

  if (moQuery.isLoading) {
    return <div className="text-muted-foreground">{"Loading..."}</div>;
  }

  if (!productionOrder) {
    return <div className="text-muted-foreground">{"Manufacturing order not found"}</div>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-2">
        <Button variant="ghost" size="sm" asChild>
          <Link href={"/production-orders"}>
            <ArrowLeft />
            <span>{"Back to orders"}</span>
          </Link>
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <CardTitle>{productionOrder.name ?? `MO-${productionOrder.id}`}</CardTitle>
              <Badge variant="outline" className={moStateTone(state)}>
                {moStateLabel(state)}
              </Badge>
            </div>
            <div className="flex items-center gap-2">
              {canConfirmMo(state) ? (
                <Button size="sm" onClick={() => handleAction("confirm")}>
                  <CheckCircle2 />
                  <span>{"Confirm"}</span>
                </Button>
              ) : null}
              {canPlanMo(state) ? (
                <Button size="sm" onClick={() => handleAction("plan")}>
                  <Play />
                  <span>{"Plan"}</span>
                </Button>
              ) : null}
              {canStartMo(state) ? (
                <Button size="sm" onClick={() => handleAction("start")}>
                  <Play />
                  <span>{"Start"}</span>
                </Button>
              ) : null}
              {canProduceMo(state) ? (
                <>
                  <Button size="sm" onClick={() => setConsumeOpen(true)}>
                    <span>{"Consume"}</span>
                  </Button>
                  <Button size="sm" onClick={() => setProduceOpen(true)}>
                    <span>{"Produce"}</span>
                  </Button>
                </>
              ) : null}
              {canCancelMo(state) ? (
                <Button size="sm" variant="destructive" onClick={() => handleAction("cancel")}>
                  <XCircle />
                  <span>{"Cancel"}</span>
                </Button>
              ) : null}
            </div>
          </div>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{"Item"}</span>
            <span className="">{productionOrder.itemId ? `#${productionOrder.itemId}` : "—"}</span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{"Quantity"}</span>
            <span className=" tabular-nums">
              {productionOrder.qtyProduced} / {productionOrder.qtyToProduce}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{"Planned Start"}</span>
            <span className="">
              {productionOrder.datePlannedStart
                ? formatDateTime(productionOrder.datePlannedStart)
                : "—"}
            </span>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{"Planned End"}</span>
            <span className="">
              {productionOrder.datePlannedFinish
                ? formatDateTime(productionOrder.datePlannedFinish)
                : "—"}
            </span>
          </div>
          <div className="sm:col-span-2 flex flex-col gap-1">
            <span className="text-xs text-muted-foreground">{"Progress"}</span>
            <div className="flex items-center gap-2">
              <Progress value={progress} className="flex-1" />
              <span className={cn("text-sm tabular-nums", moProgressColor(progress))}>
                {progress}%
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{"Components"}</CardTitle>
        </CardHeader>
        <CardContent>
          {components.length === 0 ? (
            <div className="text-muted-foreground text-sm">{"No components defined"}</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2 pr-4 ">{"Component"}</th>
                    <th className="pb-2 pr-4 text-right ">{"Planned"}</th>
                    <th className="pb-2 pr-4 text-right ">{"Consumed"}</th>
                    <th className="pb-2 ">{"Progress"}</th>
                  </tr>
                </thead>
                <tbody>
                  {components.map((c) => {
                    const cProgress = componentConsumptionProgress(c);
                    return (
                      <tr key={c.id} className="border-b last:border-0">
                        <td className="py-2 pr-4">
                          <div className="flex flex-col">
                            <span>{`#${c.itemId}`}</span>
                            <span className="text-xs text-muted-foreground">
                              {c.qtyConsumed}/{c.qtyPlanned}
                            </span>
                          </div>
                        </td>
                        <td className="py-2 pr-4 text-right tabular-nums">{c.qtyPlanned}</td>
                        <td className="py-2 pr-4 text-right tabular-nums">{c.qtyConsumed}</td>
                        <td className="py-2">
                          <div className="flex items-center gap-2">
                            <Progress value={cProgress} className="w-16" />
                            <span className="text-xs tabular-nums text-muted-foreground">
                              {cProgress}%
                            </span>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>{"Work Orders"}</CardTitle>
        </CardHeader>
        <CardContent>
          {shopTasks.length === 0 ? (
            <div className="text-muted-foreground text-sm">{"No work orders"}</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b text-left text-muted-foreground">
                    <th className="pb-2 pr-4 ">{"Work Order"}</th>
                    <th className="pb-2 pr-4 ">{"Work Center"}</th>
                    <th className="pb-2 pr-4 text-right ">{"Planned Duration"}</th>
                    <th className="pb-2 pr-4 text-right ">{"Actual Duration"}</th>
                    <th className="pb-2 ">{"State"}</th>
                  </tr>
                </thead>
                <tbody>
                  {shopTasks.map((wo) => (
                    <tr key={wo.id} className="border-b last:border-0">
                      <td className="py-2 pr-4">{wo.name ?? `WO-${wo.id}`}</td>
                      <td className="py-2 pr-4 text-muted-foreground">
                        {wo.workCenterId ? `#${wo.workCenterId}` : "—"}
                      </td>
                      <td className="py-2 pr-4 text-right tabular-nums">
                        {wo.plannedMinutes ?? "—"}
                      </td>
                      <td className="py-2 pr-4 text-right tabular-nums">
                        {wo.actualMinutes ?? "—"}
                      </td>
                      <td className="py-2">
                        <Badge variant="outline">{wo.state}</Badge>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {consumeOpen ? (
        <ConsumeDialog
          open={consumeOpen}
          onOpenChange={setConsumeOpen}
          orgId={orgId}
          productionOrderId={productionOrderId}
          components={components}
          onSave={() => {
            setConsumeOpen(false);
            void moQuery.refetch();
            void componentsQuery.refetch();
          }}
        />
      ) : null}
      {produceOpen ? (
        <ProduceDialog
          open={produceOpen}
          onOpenChange={setProduceOpen}
          orgId={orgId}
          productionOrderId={productionOrderId}
          productionOrder={productionOrder}
          onSave={() => {
            setProduceOpen(false);
            void moQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
