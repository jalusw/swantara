"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { useMemo, useState } from "react";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/tabs";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type {
  Item,
  StockBalance,
  StockLocation,
  StockMovement,
  Warehouse,
} from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber } from "@/lib/utils";

type OnHandRow = {
  id: string;
  productName: string;
  locationName: string;
  onHand: number;
  reserved: number;
  available: number;
};

type MoveRow = {
  id: string;
  productName: string;
  qty: number;
  srcLocation: string;
  dstLocation: string;
  state: StockMovement["state"];
  dateDone: Date | null;
  originType: string | null;
  originId: number | null;
};

function stateBadgeVariant(state: StockMovement["state"]) {
  switch (state) {
    case "done":
      return "default" as const;
    case "cancelled":
      return "outline" as const;
    case "draft":
      return "secondary" as const;
    default:
      return "secondary" as const;
  }
}

export function StockOverviewSection({ orgId: _orgId }: { orgId: string }) {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Stock");
  const [locationFilter, setLocationFilter] = useState<string>("all");
  const [activeTab, setActiveTab] = useState("onhand");

  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );

  const locationsQuery = useOrgListQuery<{ locations: StockLocation[] }, Record<string, never>>(
    "stockLocations",
    (organizationId) => getSwantaraService().inventory.stockLocations(organizationId),
  );

  const balancesQuery = useOrgListQuery<{ balances: StockBalance[] }, Record<string, never>>(
    "stockBalances",
    (organizationId) => getSwantaraService().inventory.balances(organizationId),
  );

  const movementsQuery = useOrgListQuery<{ movements: StockMovement[] }, Record<string, never>>(
    "stockMovements",
    (organizationId) => getSwantaraService().inventory.stockMovements(organizationId),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const locations = locationsQuery.data?.locations ?? [];
  const warehouses = warehousesQuery.data?.warehouses ?? [];
  const products = productsQuery.data?.products ?? [];

  const locationNameMap = useMemo(() => new Map(locations.map((l) => [l.id, l.name])), [locations]);
  const warehouseNameMap = useMemo(
    () => new Map(warehouses.map((w) => [w.id, w.name])),
    [warehouses],
  );
  const productNameMap = useMemo(() => new Map(products.map((p) => [p.id, p.name])), [products]);

  const balances = balancesQuery.data?.balances ?? [];
  const onHandRows: OnHandRow[] = useMemo(() => {
    const grouped = new Map<string, OnHandRow>();
    for (const balance of balances) {
      if (locationFilter !== "all" && balance.locationId !== Number(locationFilter)) {
        continue;
      }
      const key = `${balance.itemId}-${balance.locationId}`;
      if (!grouped.has(key)) {
        grouped.set(key, {
          id: key,
          productName: productNameMap.get(balance.itemId) ?? String(balance.itemId),
          locationName: locationNameMap.get(balance.locationId) ?? String(balance.locationId),
          onHand: 0,
          reserved: 0,
          available: 0,
        });
      }
      const row = grouped.get(key);
      if (row) {
        row.onHand += balance.quantity;
        row.reserved += balance.reservedQty;
        row.available += balance.quantity - balance.reservedQty;
      }
    }
    return [...grouped.values()];
  }, [balances, locationFilter, productNameMap, locationNameMap]);

  const movements = movementsQuery.data?.movements ?? [];
  const moveRows: MoveRow[] = useMemo(
    () =>
      movements.map((movement) => ({
        id: String(movement.id),
        productName: productNameMap.get(movement.itemId) ?? String(movement.itemId),
        qty: movement.qty,
        srcLocation: locationNameMap.get(movement.srcLocationId) ?? String(movement.srcLocationId),
        dstLocation: locationNameMap.get(movement.dstLocationId) ?? String(movement.dstLocationId),
        state: movement.state,
        dateDone: movement.dateDone,
        originType: movement.originType,
        originId: movement.originId,
      })),
    [movements, productNameMap, locationNameMap],
  );

  const onHandColumns: ColumnDef<OnHandRow>[] = [
    {
      accessorKey: "productName",
      header: () => t("fieldItem"),
    },
    {
      accessorKey: "locationName",
      header: () => t("fieldLocation"),
    },
    {
      accessorKey: "onHand",
      header: () => t("colOnHand"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.onHand)}</span>,
    },
    {
      accessorKey: "reserved",
      header: () => t("colReserved"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatNumber(row.original.reserved)}
        </span>
      ),
    },
    {
      accessorKey: "available",
      header: () => t("colAvailable"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatNumber(row.original.available)}</span>
      ),
    },
  ];

  const moveColumns: ColumnDef<MoveRow>[] = [
    {
      accessorKey: "productName",
      header: () => t("fieldItem"),
    },
    {
      accessorKey: "qty",
      header: () => t("colQuantity"),
      meta: { align: "right" },
      cell: ({ row }) => <span className="tabular-nums">{formatNumber(row.original.qty)}</span>,
    },
    {
      accessorKey: "srcLocation",
      header: () => t("colFrom"),
    },
    {
      accessorKey: "dstLocation",
      header: () => t("colTo"),
    },
    {
      accessorKey: "state",
      header: () => t("colStatus"),
      cell: ({ row }) => (
        <Badge variant={stateBadgeVariant(row.original.state)}>
          {(t as unknown as (k: string) => string)(`moveState_${row.original.state}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "dateDone",
      header: () => t("colDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.dateDone ? formatDate(row.original.dateDone, { nullFallback: "—" }) : "—"}
        </span>
      ),
    },
  ];

  const isLoading =
    warehousesQuery.isLoading ||
    locationsQuery.isLoading ||
    balancesQuery.isLoading ||
    movementsQuery.isLoading ||
    productsQuery.isLoading;
  const error =
    warehousesQuery.error ??
    locationsQuery.error ??
    balancesQuery.error ??
    movementsQuery.error ??
    productsQuery.error;

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-4">
        <Select value={locationFilter} onValueChange={(value) => setLocationFilter(value ?? "all")}>
          <SelectTrigger className="w-60" aria-label={t("filterByLocation")}>
            <SelectValue placeholder={t("allLocations")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t("allLocations")}</SelectItem>
            {locations.map((location) => (
              <SelectItem key={location.id} value={String(location.id)}>
                {location.name}
                {location.warehouseId
                  ? ` (${warehouseNameMap.get(location.warehouseId) ?? ""})`
                  : ""}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="onhand">{t("onHandTab")}</TabsTrigger>
          <TabsTrigger value="movements">{t("movementsTab")}</TabsTrigger>
        </TabsList>

        <TabsContent value="onhand">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t("onHandTitle")}</CardTitle>
            </CardHeader>
            <CardContent>
              <InteractiveEntityTable
                columns={onHandColumns}
                data={onHandRows}
                getRowId={(row) => row.id}
                searchKeys={["productName", "locationName"]}
                searchPlaceholder={t("searchStock")}
                ariaLabel={t("onHandTab")}
                emptyTitle={t("onHandEmpty")}
                status={
                  isLoading
                    ? { type: "loading" }
                    : error
                      ? {
                          type: "error",
                          message: error.message,
                          onRetry: () => {
                            void balancesQuery.refetch();
                            void movementsQuery.refetch();
                          },
                        }
                      : undefined
                }
              />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="movements">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">{t("movementsTitle")}</CardTitle>
            </CardHeader>
            <CardContent>
              <InteractiveEntityTable
                columns={moveColumns}
                data={moveRows}
                getRowId={(row) => row.id}
                searchKeys={["productName"]}
                searchPlaceholder={t("searchStock")}
                ariaLabel={t("movementsTab")}
                emptyTitle={t("movementsEmpty")}
                status={
                  isLoading
                    ? { type: "loading" }
                    : error
                      ? {
                          type: "error",
                          message: error.message,
                          onRetry: () => {
                            void movementsQuery.refetch();
                          },
                        }
                      : undefined
                }
              />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}
