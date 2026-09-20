export const productTypes = ["stockable", "consumable", "service", "digital"] as const;
export type ProductType = (typeof productTypes)[number];

export const trackingModes = ["none", "batch", "serial"] as const;
export type TrackingMode = (typeof trackingModes)[number];

export const costMethods = ["standard", "fifo", "average"] as const;
export type CostMethod = (typeof costMethods)[number];

export const valuationModes = ["manual", "automated"] as const;
export type ValuationMode = (typeof valuationModes)[number];

export const bomTypes = ["manufacture", "kit", "subcontract"] as const;
export type BomType = (typeof bomTypes)[number];

export type StubItemCategory = {
  id: string;
  name: string;
  parentId: string | null;
  incomeAccountId: string | null;
  expenseAccountId: string | null;
  stockCostAccountId: string | null;
  stockInputAccountId: string | null;
  stockOutputAccountId: string | null;
  cogsAccountId: string | null;
  costMethod: CostMethod | null;
  valuation: ValuationMode | null;
};

export type StubItem = {
  id: string;
  name: string;
  categoryId: string | null;
  type: ProductType;
  unitId: string | null;
  purchaseUnitId: string | null;
  listPrice: number;
  standardCost: number;
  isPurchasable: boolean;
  isSellable: boolean;
  isManufactured: boolean;
  tracking: TrackingMode;
  active: boolean;
};

export type StubVariant = {
  id: string;
  templateId: string;
  sku: string;
  barcode: string | null;
  attributes: Record<string, string>;
  extraCost: number;
  active: boolean;
};

export type StubBomLine = {
  id: string;
  componentId: string;
  qty: number;
  unitId: string | null;
  scrapPct: number;
};

export type StubBom = {
  id: string;
  itemId: string;
  code: string | null;
  qty: number;
  unitId: string | null;
  type: BomType;
  version: number;
  active: boolean;
  lines: StubBomLine[];
};

export function bomLineTotalQty(line: Pick<StubBomLine, "qty" | "scrapPct">) {
  return line.qty * (1 + line.scrapPct / 100);
}
