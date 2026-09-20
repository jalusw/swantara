import { describe, expect, it } from "vitest";

type StockBalance = {
  id: number;
  itemId: number;
  locationId: number;
  batchId: number | null;
  quantity: number;
  reservedQty: number;
};

type OnHandRow = {
  id: string;
  productName: string;
  locationName: string;
  onHand: number;
  reserved: number;
  available: number;
};

function aggregateOnHand(
  balances: StockBalance[],
  locationFilter: string,
  productNameMap: Map<number, string>,
  locationNameMap: Map<number, string>,
): OnHandRow[] {
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
}

describe("aggregateOnHand", () => {
  const productNameMap = new Map([
    [1, "Widget"],
    [2, "Gadget"],
  ]);
  const locationNameMap = new Map([
    [10, "Main Warehouse"],
    [20, "Store Front"],
  ]);

  it("returns an empty array when no balances exist", () => {
    expect(aggregateOnHand([], "all", productNameMap, locationNameMap)).toEqual([]);
  });

  it("groups balances by item and location", () => {
    const balances: StockBalance[] = [
      {
        id: 1,
        itemId: 1,
        locationId: 10,
        batchId: null,
        quantity: 100,
        reservedQty: 10,
      },
      {
        id: 2,
        itemId: 1,
        locationId: 10,
        batchId: 1,
        quantity: 50,
        reservedQty: 5,
      },
    ];
    const result = aggregateOnHand(balances, "all", productNameMap, locationNameMap);
    expect(result).toHaveLength(1);
    expect(result[0]!.onHand).toBe(150);
    expect(result[0]!.reserved).toBe(15);
    expect(result[0]!.available).toBe(135);
  });

  it("splits into separate rows for different locations", () => {
    const balances: StockBalance[] = [
      {
        id: 1,
        itemId: 1,
        locationId: 10,
        batchId: null,
        quantity: 100,
        reservedQty: 0,
      },
      {
        id: 2,
        itemId: 1,
        locationId: 20,
        batchId: null,
        quantity: 50,
        reservedQty: 0,
      },
    ];
    const result = aggregateOnHand(balances, "all", productNameMap, locationNameMap);
    expect(result).toHaveLength(2);
  });

  it("filters by location when locationFilter is set", () => {
    const balances: StockBalance[] = [
      {
        id: 1,
        itemId: 1,
        locationId: 10,
        batchId: null,
        quantity: 100,
        reservedQty: 0,
      },
      {
        id: 2,
        itemId: 1,
        locationId: 20,
        batchId: null,
        quantity: 50,
        reservedQty: 0,
      },
    ];
    const result = aggregateOnHand(balances, "10", productNameMap, locationNameMap);
    expect(result).toHaveLength(1);
    expect(result[0]!.locationName).toBe("Main Warehouse");
    expect(result[0]!.onHand).toBe(100);
  });

  it("uses item and location names from the maps", () => {
    const balances: StockBalance[] = [
      {
        id: 1,
        itemId: 1,
        locationId: 10,
        batchId: null,
        quantity: 10,
        reservedQty: 0,
      },
    ];
    const result = aggregateOnHand(balances, "all", productNameMap, locationNameMap);
    expect(result[0]!.productName).toBe("Widget");
    expect(result[0]!.locationName).toBe("Main Warehouse");
  });

  it("falls back to numeric IDs when name maps are empty", () => {
    const balances: StockBalance[] = [
      {
        id: 1,
        itemId: 99,
        locationId: 99,
        batchId: null,
        quantity: 5,
        reservedQty: 0,
      },
    ];
    const result = aggregateOnHand(balances, "all", new Map(), new Map());
    expect(result[0]!.productName).toBe("99");
    expect(result[0]!.locationName).toBe("99");
  });
});
