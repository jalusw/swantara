import type { AxiosInstance } from "axios";
import { endpoints } from "./endpoints";
import type {
  AddServiceOrderLineRequest,
  AssetCategory,
  AssetDepreciationLine,
  AvailableToPromise,
  Batch,
  BestSupplierOfferRequest,
  BillPurchaseOrderRequest,
  BillServiceOrderRequest,
  BomComponentRequirement,
  CompleteServiceOrderRequest,
  ConfirmPlannedRequest,
  ConsumeMaterialRequest,
  CostLayer,
  CreateAssetCategoryRequest,
  CreateBatchRequest,
  CreateBomRequest,
  CreateEquipmentRequest,
  CreateForecastRequest,
  CreateInboundCostRequest,
  CreateItemCategoryRequest,
  CreateMaintenancePlanRequest,
  CreatePriceBookRequest,
  CreatePriceRuleRequest,
  CreateProductionOrderRequest,
  CreateProductRequest,
  CreatePurchaseOrderRequest,
  CreatePurchaseRequestRequest,
  CreateQualityPointRequest,
  CreateServiceContractRequest,
  CreateServiceOrderRequest,
  CreateStockCountRequest,
  CreateStockLocationRequest,
  CreateStockMovementRequest,
  CreateOutsideProcessingOrderRequest,
  CreateSupplierProductRequest,
  CreateSupplierQuoteRequestFromRequisitionRequest,
  CreateSupplierQuoteRequestRequest,
  CreateVariantRequest,
  CreateWarehouseRequest,
  CreateWarehouseTransferRequest,
  DemandPlan,
  DisposeAssetRequest,
  Equipment,
  ExplodeBomRequest,
  FixedAsset,
  InboundCost,
  InboundCostAdjustment,
  InboundCostLine,
  InvoiceSummary,
  Item,
  ItemCategory,
  ItemVariant,
  ListQuery,
  MaintenancePlan,
  MOComponent,
  OnHand,
  PaymentSummary,
  PayPurchaseOrderRequest,
  PlannedSupply,
  PlanningRun,
  PostDepreciationRequest,
  PostStockCountRequest,
  PriceBook,
  PriceRule,
  ProduceGoodsRequest,
  ProductionOrder,
  PurchaseOrder,
  PurchaseRequest,
  QualityAlert,
  QualityCheck,
  QualityPoint,
  ReceiveMoveRequest,
  ReceivePurchaseOrderRequest,
  ReceiveOutsideProcessingOrderRequest,
  Recipe,
  RecordLaborRequest,
  RecordQualityCheckRequest,
  RegisterAssetRequest,
  ReleaseByMovementRequest,
  ReorderRule,
  ReorderRuleRequest,
  ReplenishmentCandidate,
  ReserveStockRequest,
  ResolvePriceRequest,
  ResolvePriceResponse,
  RouteToScrapRequest,
  RunMrpRequest,
  ScheduleServiceOrderRequest,
  SendOutsideProcessingOrderRequest,
  ServiceContract,
  ServiceOrder,
  ServiceOrderLine,
  SettleVarianceRequest,
  ShipMoveRequest,
  Shipment,
  ShopTask,
  StockBalance,
  StockCount,
  StockCountLine,
  StockHold,
  StockLocation,
  StockMovement,
  OutsideProcessingOrder,
  SubmitSupplierQuoteRequest,
  SuccessEnvelope,
  SupplierProduct,
  SupplierQuote,
  SupplierQuoteRequest,
  SupplierQuoteRequestLine,
  TransferActionRequest,
  UpdateBomRequest,
  UpdateProductRequest,
  UpdateQualityAlertRequest,
  UpdateSupplierProductRequest,
  Warehouse,
  WarehouseTransfer,
} from "./types";
import { withListMeta } from "./types";

export type UpdatePriceBookRequest = {
  name: string;
  currencyCode: string | null;
  active: boolean | null;
};

export class ProductCategories {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ categories: ItemCategory[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ categories: ItemCategory[] }>>(
      endpoints.itemCategories.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ category: ItemCategory }> {
    const response = await this.axios.get<SuccessEnvelope<{ category: ItemCategory }>>(
      endpoints.itemCategories.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateItemCategoryRequest,
  ): Promise<{ category: ItemCategory }> {
    const response = await this.axios.post<SuccessEnvelope<{ category: ItemCategory }>>(
      endpoints.itemCategories.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: Partial<CreateItemCategoryRequest>,
  ): Promise<{ category: ItemCategory }> {
    const response = await this.axios.put<SuccessEnvelope<{ category: ItemCategory }>>(
      endpoints.itemCategories.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.itemCategories.delete(String(organizationId), String(id)));
  }
}

export class Products {
  constructor(private readonly axios: AxiosInstance) {}

  variants = {
    list: async (
      organizationId: number,
      itemId: number,
      params?: ListQuery,
    ): Promise<{ variants: ItemVariant[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ variants: ItemVariant[] }>>(
        endpoints.products.variants.list(String(organizationId), String(itemId)),
        { params },
      );
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      itemId: number,
      request: CreateVariantRequest,
    ): Promise<{ variant: ItemVariant }> => {
      const response = await this.axios.post<SuccessEnvelope<{ variant: ItemVariant }>>(
        endpoints.products.variants.create(String(organizationId), String(itemId)),
        request,
      );
      return withListMeta(response.data);
    },
  };

  async list(organizationId: number, params?: ListQuery): Promise<{ products: Item[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ products: Item[] }>>(
      endpoints.products.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ item: Item }> {
    const response = await this.axios.get<SuccessEnvelope<{ item: Item }>>(
      endpoints.products.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, request: CreateProductRequest): Promise<{ item: Item }> {
    const response = await this.axios.post<SuccessEnvelope<{ item: Item }>>(
      endpoints.products.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateProductRequest,
  ): Promise<{ item: Item }> {
    const response = await this.axios.put<SuccessEnvelope<{ item: Item }>>(
      endpoints.products.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.products.delete(String(organizationId), String(id)));
  }
}

export class PriceBooks {
  constructor(private readonly axios: AxiosInstance) {}

  rules = {
    list: async (
      organizationId: number,
      priceBookId: number,
      params?: ListQuery,
    ): Promise<{ rules: PriceRule[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ rules: PriceRule[] }>>(
        endpoints.priceBooks.rules.list(String(organizationId), String(priceBookId)),
        { params },
      );
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      priceBookId: number,
      request: CreatePriceRuleRequest,
    ): Promise<{ rule: PriceRule }> => {
      const response = await this.axios.post<SuccessEnvelope<{ rule: PriceRule }>>(
        endpoints.priceBooks.rules.create(String(organizationId), String(priceBookId)),
        request,
      );
      return withListMeta(response.data);
    },
  };

  async list(organizationId: number, params?: ListQuery): Promise<{ priceBooks: PriceBook[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ priceBooks: PriceBook[] }>>(
      endpoints.priceBooks.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ priceBook: PriceBook }> {
    const response = await this.axios.get<SuccessEnvelope<{ priceBook: PriceBook }>>(
      endpoints.priceBooks.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreatePriceBookRequest,
  ): Promise<{ priceBook: PriceBook }> {
    const response = await this.axios.post<SuccessEnvelope<{ priceBook: PriceBook }>>(
      endpoints.priceBooks.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdatePriceBookRequest,
  ): Promise<{ priceBook: PriceBook }> {
    const response = await this.axios.put<SuccessEnvelope<{ priceBook: PriceBook }>>(
      endpoints.priceBooks.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.priceBooks.delete(String(organizationId), String(id)));
  }

  async resolve(
    organizationId: number,
    priceBookId: number,
    request: ResolvePriceRequest,
  ): Promise<{ price: ResolvePriceResponse }> {
    const response = await this.axios.post<SuccessEnvelope<{ price: ResolvePriceResponse }>>(
      endpoints.priceBooks.resolve(String(organizationId), String(priceBookId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class SupplierProducts {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ supplierProducts: SupplierProduct[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ supplierProducts: SupplierProduct[] }>>(
      endpoints.supplierProducts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ supplierProduct: SupplierProduct }> {
    const response = await this.axios.get<SuccessEnvelope<{ supplierProduct: SupplierProduct }>>(
      endpoints.supplierProducts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateSupplierProductRequest,
  ): Promise<{ supplierProduct: SupplierProduct }> {
    const response = await this.axios.post<SuccessEnvelope<{ supplierProduct: SupplierProduct }>>(
      endpoints.supplierProducts.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateSupplierProductRequest,
  ): Promise<{ supplierProduct: SupplierProduct }> {
    const response = await this.axios.put<SuccessEnvelope<{ supplierProduct: SupplierProduct }>>(
      endpoints.supplierProducts.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.supplierProducts.delete(String(organizationId), String(id)));
  }

  async bestOffer(
    organizationId: number,
    request: BestSupplierOfferRequest,
  ): Promise<{ supplierProduct: SupplierProduct }> {
    const response = await this.axios.post<SuccessEnvelope<{ supplierProduct: SupplierProduct }>>(
      endpoints.supplierProducts.bestOffer(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class Inventory {
  constructor(private readonly axios: AxiosInstance) {}

  async warehouses(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ warehouses: Warehouse[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ warehouses: Warehouse[] }>>(
      endpoints.warehouses.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async warehouse(organizationId: number, id: number): Promise<{ warehouse: Warehouse }> {
    const response = await this.axios.get<SuccessEnvelope<{ warehouse: Warehouse }>>(
      endpoints.warehouses.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createWarehouse(
    organizationId: number,
    request: CreateWarehouseRequest,
  ): Promise<{ warehouse: Warehouse }> {
    const response = await this.axios.post<SuccessEnvelope<{ warehouse: Warehouse }>>(
      endpoints.warehouses.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async updateWarehouse(
    organizationId: number,
    id: number,
    request: Partial<CreateWarehouseRequest>,
  ): Promise<{ warehouse: Warehouse }> {
    const response = await this.axios.put<SuccessEnvelope<{ warehouse: Warehouse }>>(
      endpoints.warehouses.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async deleteWarehouse(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.warehouses.delete(String(organizationId), String(id)));
  }

  async stockLocations(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ locations: StockLocation[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ locations: StockLocation[] }>>(
      endpoints.stockLocations.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async stockLocation(organizationId: number, id: number): Promise<{ location: StockLocation }> {
    const response = await this.axios.get<SuccessEnvelope<{ location: StockLocation }>>(
      endpoints.stockLocations.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createStockLocation(
    organizationId: number,
    request: CreateStockLocationRequest,
  ): Promise<{ location: StockLocation }> {
    const response = await this.axios.post<SuccessEnvelope<{ location: StockLocation }>>(
      endpoints.stockLocations.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async updateStockLocation(
    organizationId: number,
    id: number,
    request: Partial<CreateStockLocationRequest>,
  ): Promise<{ location: StockLocation }> {
    const response = await this.axios.put<SuccessEnvelope<{ location: StockLocation }>>(
      endpoints.stockLocations.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async deleteStockLocation(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.stockLocations.delete(String(organizationId), String(id)));
  }

  async onHand(organizationId: number, params?: ListQuery): Promise<{ onHand: OnHand }> {
    const response = await this.axios.get<SuccessEnvelope<{ onHand: OnHand }>>(
      endpoints.stock.onHand(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async availableToPromise(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ availableToPromise: AvailableToPromise }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ availableToPromise: AvailableToPromise }>
    >(endpoints.stock.availableToPromise(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async balances(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ balances: StockBalance[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ balances: StockBalance[] }>>(
      endpoints.stock.balances(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async rebuild(organizationId: number): Promise<void> {
    await this.axios.post(endpoints.stock.rebuild(String(organizationId)));
  }

  async stockMovements(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ movements: StockMovement[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ movements: StockMovement[] }>>(
      endpoints.stockMovements.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async stockMovement(organizationId: number, id: number): Promise<{ movement: StockMovement }> {
    const response = await this.axios.get<SuccessEnvelope<{ movement: StockMovement }>>(
      endpoints.stockMovements.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createStockMovement(
    organizationId: number,
    request: CreateStockMovementRequest,
  ): Promise<{ movement: StockMovement }> {
    const response = await this.axios.post<SuccessEnvelope<{ movement: StockMovement }>>(
      endpoints.stockMovements.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async deleteStockMovement(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.stockMovements.delete(String(organizationId), String(id)));
  }

  async receive(
    organizationId: number,
    id: number,
    request: ReceiveMoveRequest,
  ): Promise<{ layer: CostLayer }> {
    const response = await this.axios.post<SuccessEnvelope<{ layer: CostLayer }>>(
      endpoints.stockMovements.receive(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async ship(
    organizationId: number,
    id: number,
    request: ShipMoveRequest,
  ): Promise<{ layer: CostLayer }> {
    const response = await this.axios.post<SuccessEnvelope<{ layer: CostLayer }>>(
      endpoints.stockMovements.ship(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async stockShipments(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ shipments: Shipment[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ shipments: Shipment[] }>>(
      endpoints.shipments.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async stockShipment(organizationId: number, id: number): Promise<{ shipment: Shipment }> {
    const response = await this.axios.get<SuccessEnvelope<{ shipment: Shipment }>>(
      endpoints.shipments.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async stockReservations(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ reservations: StockHold[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ reservations: StockHold[] }>>(
      endpoints.stockHolds.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async createStockHold(
    organizationId: number,
    request: ReserveStockRequest,
  ): Promise<{ reservation: StockHold }> {
    const response = await this.axios.post<SuccessEnvelope<{ reservation: StockHold }>>(
      endpoints.stockHolds.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async releaseByMovement(
    organizationId: number,
    request: ReleaseByMovementRequest,
  ): Promise<void> {
    await this.axios.post(endpoints.stockHolds.releaseByMovement(String(organizationId)), request);
  }

  async deleteStockHold(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.stockHolds.delete(String(organizationId), String(id)));
  }

  async batches(organizationId: number, params?: ListQuery): Promise<{ batches: Batch[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ batches: Batch[] }>>(
      endpoints.batches.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async batch(organizationId: number, id: number): Promise<{ batch: Batch }> {
    const response = await this.axios.get<SuccessEnvelope<{ batch: Batch }>>(
      endpoints.batches.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createBatch(
    organizationId: number,
    request: CreateBatchRequest,
  ): Promise<{ batch: Batch }> {
    const response = await this.axios.post<SuccessEnvelope<{ batch: Batch }>>(
      endpoints.batches.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async updateBatch(
    organizationId: number,
    id: number,
    request: Partial<CreateBatchRequest>,
  ): Promise<{ batch: Batch }> {
    const response = await this.axios.put<SuccessEnvelope<{ batch: Batch }>>(
      endpoints.batches.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async reorderCandidates(
    organizationId: number,
  ): Promise<{ candidates: ReplenishmentCandidate[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ candidates: ReplenishmentCandidate[] }>
    >(endpoints.reorderRules.candidates(String(organizationId)));
    return withListMeta(response.data);
  }

  async reorderRules(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ rules: ReorderRule[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ rules: ReorderRule[] }>>(
      endpoints.reorderRules.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async reorderRule(organizationId: number, id: number): Promise<{ rule: ReorderRule }> {
    const response = await this.axios.get<SuccessEnvelope<{ rule: ReorderRule }>>(
      endpoints.reorderRules.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createReorderRule(
    organizationId: number,
    request: ReorderRuleRequest,
  ): Promise<{ rule: ReorderRule }> {
    const response = await this.axios.post<SuccessEnvelope<{ rule: ReorderRule }>>(
      endpoints.reorderRules.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async updateReorderRule(
    organizationId: number,
    id: number,
    request: ReorderRuleRequest,
  ): Promise<{ rule: ReorderRule }> {
    const response = await this.axios.put<SuccessEnvelope<{ rule: ReorderRule }>>(
      endpoints.reorderRules.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async deleteReorderRule(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.reorderRules.delete(String(organizationId), String(id)));
  }

  async inventoryCounts(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ counts: StockCount[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ counts: StockCount[] }>>(
      endpoints.stockCounts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async inventoryCount(organizationId: number, id: number): Promise<{ count: StockCount }> {
    const response = await this.axios.get<SuccessEnvelope<{ count: StockCount }>>(
      endpoints.stockCounts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createStockCount(
    organizationId: number,
    request: CreateStockCountRequest,
  ): Promise<{ count: StockCount }> {
    const response = await this.axios.post<SuccessEnvelope<{ count: StockCount }>>(
      endpoints.stockCounts.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async inventoryCountLines(
    organizationId: number,
    id: number,
  ): Promise<{ lines: StockCountLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ lines: StockCountLine[] }>>(
      endpoints.stockCounts.lines(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async postStockCount(
    organizationId: number,
    id: number,
    request: PostStockCountRequest,
  ): Promise<{ count: StockCount }> {
    const response = await this.axios.post<SuccessEnvelope<{ count: StockCount }>>(
      endpoints.stockCounts.post(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async deleteStockCount(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.stockCounts.delete(String(organizationId), String(id)));
  }

  async transferOrders(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ warehouseTransfers: WarehouseTransfer[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ warehouseTransfers: WarehouseTransfer[] }>
    >(endpoints.warehouseTransfers.list(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async transferOrder(
    organizationId: number,
    id: number,
  ): Promise<{ warehouseTransfer: WarehouseTransfer }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ warehouseTransfer: WarehouseTransfer }>
    >(endpoints.warehouseTransfers.get(String(organizationId), String(id)));
    return withListMeta(response.data);
  }

  async createWarehouseTransfer(
    organizationId: number,
    request: CreateWarehouseTransferRequest,
  ): Promise<{ warehouseTransfer: WarehouseTransfer }> {
    const response = await this.axios.post<
      SuccessEnvelope<{ warehouseTransfer: WarehouseTransfer }>
    >(endpoints.warehouseTransfers.create(String(organizationId)), request);
    return withListMeta(response.data);
  }

  async sendWarehouseTransfer(
    organizationId: number,
    id: number,
    request: TransferActionRequest,
  ): Promise<void> {
    await this.axios.post(
      endpoints.warehouseTransfers.send(String(organizationId), String(id)),
      request,
    );
  }

  async receiveWarehouseTransfer(
    organizationId: number,
    id: number,
    request: TransferActionRequest,
  ): Promise<void> {
    await this.axios.post(
      endpoints.warehouseTransfers.receive(String(organizationId), String(id)),
      request,
    );
  }

  async inboundCosts(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ costs: InboundCost[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ costs: InboundCost[] }>>(
      endpoints.inboundCosts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async inboundCost(organizationId: number, id: number): Promise<InboundCost> {
    const response = await this.axios.get<SuccessEnvelope<InboundCost>>(
      endpoints.inboundCosts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async createInboundCost(
    organizationId: number,
    request: CreateInboundCostRequest,
  ): Promise<InboundCost> {
    const response = await this.axios.post<SuccessEnvelope<InboundCost>>(
      endpoints.inboundCosts.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async inboundCostLines(
    organizationId: number,
    id: number,
  ): Promise<{ lines: InboundCostLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ lines: InboundCostLine[] }>>(
      endpoints.inboundCosts.lines(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async inboundCostAdjustments(
    organizationId: number,
    id: number,
  ): Promise<{ adjustments: InboundCostAdjustment[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ adjustments: InboundCostAdjustment[] }>
    >(endpoints.inboundCosts.adjustments(String(organizationId), String(id)));
    return withListMeta(response.data);
  }

  async postInboundCost(organizationId: number, id: number): Promise<InboundCost> {
    const response = await this.axios.post<SuccessEnvelope<InboundCost>>(
      endpoints.inboundCosts.post(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export type CreateShopTaskRequest = {
  productionStepId?: number | null;
  workCenterId: number;
  sequence?: number;
  plannedStart?: Date | null;
  plannedFinish?: Date | null;
  plannedMinutes?: number;
};

export class Boms {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ recipes: Recipe[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ recipes: Recipe[] }>>(
      endpoints.recipes.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ recipe: Recipe }> {
    const response = await this.axios.get<SuccessEnvelope<{ recipe: Recipe }>>(
      endpoints.recipes.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateBomRequest): Promise<{ recipe: Recipe }> {
    const response = await this.axios.post<SuccessEnvelope<{ recipe: Recipe }>>(
      endpoints.recipes.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateBomRequest,
  ): Promise<{ recipe: Recipe }> {
    const response = await this.axios.put<SuccessEnvelope<{ recipe: Recipe }>>(
      endpoints.recipes.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.recipes.delete(String(organizationId), String(id)));
  }

  async lines(organizationId: number, id: number): Promise<{ lines: BomComponentRequirement[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ lines: BomComponentRequirement[] }>>(
      endpoints.recipes.lines(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async explode(
    organizationId: number,
    id: number,
    body: ExplodeBomRequest,
  ): Promise<{ components: BomComponentRequirement[] }> {
    const response = await this.axios.post<
      SuccessEnvelope<{ components: BomComponentRequirement[] }>
    >(endpoints.recipes.explode(String(organizationId), String(id)), body);
    return withListMeta(response.data);
  }
}

export class ProductionOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ productionOrders: ProductionOrder[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ productionOrders: ProductionOrder[] }>>(
      endpoints.productionOrders.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateProductionOrderRequest,
  ): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.get<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async components(organizationId: number, id: number): Promise<{ components: MOComponent[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ components: MOComponent[] }>>(
      endpoints.productionOrders.components(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async confirm(organizationId: number, id: number): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.confirm(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async plan(organizationId: number, id: number): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.plan(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async start(organizationId: number, id: number): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.start(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async consume(
    organizationId: number,
    id: number,
    body: ConsumeMaterialRequest,
  ): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.consume(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async produce(
    organizationId: number,
    id: number,
    body: ProduceGoodsRequest,
  ): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.produce(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async settle(
    organizationId: number,
    id: number,
    body: SettleVarianceRequest,
  ): Promise<{ productionOrder: ProductionOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ productionOrder: ProductionOrder }>>(
      endpoints.productionOrders.settle(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  shopTasks = {
    list: async (
      organizationId: number,
      id: number,
      params?: ListQuery,
    ): Promise<{ shopTasks: ShopTask[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ shopTasks: ShopTask[] }>>(
        endpoints.productionOrders.shopTasks.list(String(organizationId), String(id)),
        {
          params,
        },
      );
      return withListMeta(response.data);
    },
    create: async (
      organizationId: number,
      id: number,
      body: CreateShopTaskRequest,
    ): Promise<{ shopTask: ShopTask }> => {
      const response = await this.axios.post<SuccessEnvelope<{ shopTask: ShopTask }>>(
        endpoints.productionOrders.shopTasks.create(String(organizationId), String(id)),
        body,
      );
      return withListMeta(response.data);
    },
    labor: async (
      organizationId: number,
      productionOrderId: number,
      shopTaskId: number,
      body: RecordLaborRequest,
    ): Promise<{ shopTask: ShopTask }> => {
      const response = await this.axios.post<SuccessEnvelope<{ shopTask: ShopTask }>>(
        endpoints.productionOrders.shopTasks.labor(
          String(organizationId),
          String(productionOrderId),
          String(shopTaskId),
        ),
        body,
      );
      return withListMeta(response.data);
    },
  };
}

export class Planning {
  constructor(private readonly axios: AxiosInstance) {}

  runs = {
    list: async (organizationId: number, params?: ListQuery): Promise<{ runs: PlanningRun[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ runs: PlanningRun[] }>>(
        endpoints.planning.runs.list(String(organizationId)),
        { params },
      );
      return withListMeta(response.data);
    },
    create: async (organizationId: number, body: RunMrpRequest): Promise<{ run: PlanningRun }> => {
      const response = await this.axios.post<SuccessEnvelope<{ run: PlanningRun }>>(
        endpoints.planning.runs.create(String(organizationId)),
        body,
      );
      return withListMeta(response.data);
    },
    get: async (organizationId: number, id: number): Promise<{ run: PlanningRun }> => {
      const response = await this.axios.get<SuccessEnvelope<{ run: PlanningRun }>>(
        endpoints.planning.runs.get(String(organizationId), String(id)),
      );
      return withListMeta(response.data);
    },
  };

  async confirmPlannedOrder(
    organizationId: number,
    id: number,
    body: ConfirmPlannedRequest,
  ): Promise<{ plannedOrder: PlannedSupply }> {
    const response = await this.axios.post<SuccessEnvelope<{ plannedOrder: PlannedSupply }>>(
      endpoints.planning.plannedOrders.confirm(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  forecasts = {
    list: async (
      organizationId: number,
      params?: ListQuery,
    ): Promise<{ forecasts: DemandPlan[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ forecasts: DemandPlan[] }>>(
        endpoints.planning.forecasts.list(String(organizationId)),
        { params },
      );
      return withListMeta(response.data);
    },
    create: async (
      organizationId: number,
      body: CreateForecastRequest,
    ): Promise<{ forecast: DemandPlan }> => {
      const response = await this.axios.post<SuccessEnvelope<{ forecast: DemandPlan }>>(
        endpoints.planning.forecasts.create(String(organizationId)),
        body,
      );
      return withListMeta(response.data);
    },
  };
}

export class OutsideProcessingOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async create(
    organizationId: number,
    body: CreateOutsideProcessingOrderRequest,
  ): Promise<{ outsideProcessingOrder: OutsideProcessingOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ outsideProcessingOrder: OutsideProcessingOrder }>>(
      endpoints.outsideProcessingOrders.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async send(
    organizationId: number,
    id: number,
    body: SendOutsideProcessingOrderRequest,
  ): Promise<{ outsideProcessingOrder: OutsideProcessingOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ outsideProcessingOrder: OutsideProcessingOrder }>>(
      endpoints.outsideProcessingOrders.send(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async receive(
    organizationId: number,
    id: number,
    body: ReceiveOutsideProcessingOrderRequest,
  ): Promise<{ outsideProcessingOrder: OutsideProcessingOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ outsideProcessingOrder: OutsideProcessingOrder }>>(
      endpoints.outsideProcessingOrders.receive(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async done(organizationId: number, id: number): Promise<{ outsideProcessingOrder: OutsideProcessingOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ outsideProcessingOrder: OutsideProcessingOrder }>>(
      endpoints.outsideProcessingOrders.done(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(
    organizationId: number,
    id: number,
  ): Promise<{ outsideProcessingOrder: OutsideProcessingOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ outsideProcessingOrder: OutsideProcessingOrder }>>(
      endpoints.outsideProcessingOrders.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class PurchaseRequests {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ requisitions: PurchaseRequest[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ requisitions: PurchaseRequest[] }>>(
      endpoints.purchaseRequests.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ request: PurchaseRequest }> {
    const response = await this.axios.get<SuccessEnvelope<{ request: PurchaseRequest }>>(
      endpoints.purchaseRequests.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreatePurchaseRequestRequest,
  ): Promise<{ request: PurchaseRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ request: PurchaseRequest }>>(
      endpoints.purchaseRequests.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async confirm(organizationId: number, id: number): Promise<{ request: PurchaseRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ request: PurchaseRequest }>>(
      endpoints.purchaseRequests.confirm(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async approve(organizationId: number, id: number): Promise<{ request: PurchaseRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ request: PurchaseRequest }>>(
      endpoints.purchaseRequests.approve(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ request: PurchaseRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ request: PurchaseRequest }>>(
      endpoints.purchaseRequests.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class PurchaseOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ orders: PurchaseOrder[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ orders: PurchaseOrder[] }>>(
      endpoints.purchaseOrders.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.get<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreatePurchaseOrderRequest,
  ): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: Partial<CreatePurchaseOrderRequest>,
  ): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.put<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async confirm(organizationId: number, id: number): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.confirm(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async receive(
    organizationId: number,
    id: number,
    request: ReceivePurchaseOrderRequest,
  ): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.purchaseOrders.receive(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async vendorBill(
    organizationId: number,
    id: number,
    request: BillPurchaseOrderRequest,
  ): Promise<{ invoice: InvoiceSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: InvoiceSummary }>>(
      endpoints.purchaseOrders.vendorBill(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async pay(
    organizationId: number,
    id: number,
    request: PayPurchaseOrderRequest,
  ): Promise<{ payment: PaymentSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ payment: PaymentSummary }>>(
      endpoints.purchaseOrders.pay(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class SupplierQuoteRequests {
  constructor(private readonly axios: AxiosInstance) {}

  quotes = {
    list: async (
      organizationId: number,
      id: number,
      params?: ListQuery,
    ): Promise<{ quotes: SupplierQuote[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ quotes: SupplierQuote[] }>>(
        endpoints.supplierQuoteRequests.quotes.list(String(organizationId), String(id)),
        {
          params,
        },
      );
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      id: number,
      request: SubmitSupplierQuoteRequest,
    ): Promise<{ quote: SupplierQuote }> => {
      const response = await this.axios.post<SuccessEnvelope<{ quote: SupplierQuote }>>(
        endpoints.supplierQuoteRequests.quotes.create(String(organizationId), String(id)),
        request,
      );
      return withListMeta(response.data);
    },

    accept: async (
      organizationId: number,
      id: number,
      quoteId: number,
    ): Promise<{ quote: SupplierQuote }> => {
      const response = await this.axios.post<SuccessEnvelope<{ quote: SupplierQuote }>>(
        endpoints.supplierQuoteRequests.quotes.accept(
          String(organizationId),
          String(id),
          String(quoteId),
        ),
      );
      return withListMeta(response.data);
    },
  };

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ quoteRequests: SupplierQuoteRequest[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ quoteRequests: SupplierQuoteRequest[] }>
    >(endpoints.supplierQuoteRequests.list(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ quoteRequest: SupplierQuoteRequest }> {
    const response = await this.axios.get<SuccessEnvelope<{ quoteRequest: SupplierQuoteRequest }>>(
      endpoints.supplierQuoteRequests.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateSupplierQuoteRequestRequest,
  ): Promise<{ quoteRequest: SupplierQuoteRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ quoteRequest: SupplierQuoteRequest }>>(
      endpoints.supplierQuoteRequests.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async fromRequisition(
    organizationId: number,
    request: CreateSupplierQuoteRequestFromRequisitionRequest,
  ): Promise<{ quoteRequest: SupplierQuoteRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ quoteRequest: SupplierQuoteRequest }>>(
      endpoints.supplierQuoteRequests.fromRequisition(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async send(organizationId: number, id: number): Promise<{ quoteRequest: SupplierQuoteRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ quoteRequest: SupplierQuoteRequest }>>(
      endpoints.supplierQuoteRequests.send(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(
    organizationId: number,
    id: number,
  ): Promise<{ quoteRequest: SupplierQuoteRequest }> {
    const response = await this.axios.post<SuccessEnvelope<{ quoteRequest: SupplierQuoteRequest }>>(
      endpoints.supplierQuoteRequests.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async lines(organizationId: number, id: number): Promise<{ lines: SupplierQuoteRequestLine[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ lines: SupplierQuoteRequestLine[] }>>(
      endpoints.supplierQuoteRequests.lines(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async purchaseOrder(organizationId: number, id: number): Promise<{ order: PurchaseOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PurchaseOrder }>>(
      endpoints.supplierQuoteRequests.purchaseOrder(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class QualityPoints {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ points: QualityPoint[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ points: QualityPoint[] }>>(
      endpoints.quality.points.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ point: QualityPoint }> {
    const response = await this.axios.get<SuccessEnvelope<{ point: QualityPoint }>>(
      endpoints.quality.points.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateQualityPointRequest,
  ): Promise<{ point: QualityPoint }> {
    const response = await this.axios.post<SuccessEnvelope<{ point: QualityPoint }>>(
      endpoints.quality.points.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class QualityChecks {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ checks: QualityCheck[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ checks: QualityCheck[] }>>(
      endpoints.quality.checks.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ check: QualityCheck }> {
    const response = await this.axios.get<SuccessEnvelope<{ check: QualityCheck }>>(
      endpoints.quality.checks.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async result(
    organizationId: number,
    id: number,
    request: RecordQualityCheckRequest,
  ): Promise<{ check: QualityCheck }> {
    const response = await this.axios.post<SuccessEnvelope<{ check: QualityCheck }>>(
      endpoints.quality.checks.result(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async scrap(
    organizationId: number,
    shipmentId: number,
    request: RouteToScrapRequest,
  ): Promise<{ scrappedItemIds: number[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ scrappedItemIds: number[] }>>(
      endpoints.quality.checks.scrap(String(organizationId), String(shipmentId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class QualityAlerts {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ alerts: QualityAlert[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ alerts: QualityAlert[] }>>(
      endpoints.quality.alerts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ alert: QualityAlert }> {
    const response = await this.axios.get<SuccessEnvelope<{ alert: QualityAlert }>>(
      endpoints.quality.alerts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async state(
    organizationId: number,
    id: number,
    request: UpdateQualityAlertRequest,
  ): Promise<{ alert: QualityAlert }> {
    const response = await this.axios.put<SuccessEnvelope<{ alert: QualityAlert }>>(
      endpoints.quality.alerts.state(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class AssetCategories {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ assetCategories: AssetCategory[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ assetCategories: AssetCategory[] }>>(
      endpoints.assetCategories.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ assetCategory: AssetCategory }> {
    const response = await this.axios.get<SuccessEnvelope<{ assetCategory: AssetCategory }>>(
      endpoints.assetCategories.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateAssetCategoryRequest,
  ): Promise<{ assetCategory: AssetCategory }> {
    const response = await this.axios.post<SuccessEnvelope<{ assetCategory: AssetCategory }>>(
      endpoints.assetCategories.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class FixedAssets {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ fixedAssets: FixedAsset[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ fixedAssets: FixedAsset[] }>>(
      endpoints.fixedAssets.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ fixedAsset: FixedAsset }> {
    const response = await this.axios.get<SuccessEnvelope<{ fixedAsset: FixedAsset }>>(
      endpoints.fixedAssets.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: RegisterAssetRequest,
  ): Promise<{ fixedAsset: FixedAsset }> {
    const response = await this.axios.post<SuccessEnvelope<{ fixedAsset: FixedAsset }>>(
      endpoints.fixedAssets.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async schedule(organizationId: number, id: number): Promise<{ lines: AssetDepreciationLine[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ lines: AssetDepreciationLine[] }>>(
      endpoints.fixedAssets.schedule(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async postDepreciation(
    organizationId: number,
    id: number,
    body: PostDepreciationRequest,
  ): Promise<{ fixedAsset: FixedAsset }> {
    const response = await this.axios.post<SuccessEnvelope<{ fixedAsset: FixedAsset }>>(
      endpoints.fixedAssets.postDepreciation(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async dispose(
    organizationId: number,
    id: number,
    body: DisposeAssetRequest,
  ): Promise<{ fixedAsset: FixedAsset }> {
    const response = await this.axios.post<SuccessEnvelope<{ fixedAsset: FixedAsset }>>(
      endpoints.fixedAssets.dispose(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }
}

export type GenerateMaintenanceOrdersRequest = {
  maintenancePlanId: number;
};

export class Equipments {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ equipments: Equipment[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ equipments: Equipment[] }>>(
      endpoints.equipments.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ equipment: Equipment }> {
    const response = await this.axios.get<SuccessEnvelope<{ equipment: Equipment }>>(
      endpoints.equipments.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateEquipmentRequest,
  ): Promise<{ equipment: Equipment }> {
    const response = await this.axios.post<SuccessEnvelope<{ equipment: Equipment }>>(
      endpoints.equipments.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class ServiceContracts {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ serviceContracts: ServiceContract[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ serviceContracts: ServiceContract[] }>>(
      endpoints.serviceContracts.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ serviceContract: ServiceContract }> {
    const response = await this.axios.get<SuccessEnvelope<{ serviceContract: ServiceContract }>>(
      endpoints.serviceContracts.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateServiceContractRequest,
  ): Promise<{ serviceContract: ServiceContract }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceContract: ServiceContract }>>(
      endpoints.serviceContracts.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async activate(
    organizationId: number,
    id: number,
  ): Promise<{ serviceContract: ServiceContract }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceContract: ServiceContract }>>(
      endpoints.serviceContracts.activate(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ serviceContract: ServiceContract }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceContract: ServiceContract }>>(
      endpoints.serviceContracts.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class ServiceOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ serviceOrders: ServiceOrder[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ serviceOrders: ServiceOrder[] }>>(
      endpoints.serviceOrders.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.get<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateServiceOrderRequest,
  ): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  lines = {
    list: async (
      organizationId: number,
      id: number,
      params?: ListQuery,
    ): Promise<{ serviceOrderLines: ServiceOrderLine[] }> => {
      const response = await this.axios.get<
        SuccessEnvelope<{ serviceOrderLines: ServiceOrderLine[] }>
      >(endpoints.serviceOrders.lines.list(String(organizationId), String(id)), { params });
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      id: number,
      body: AddServiceOrderLineRequest,
    ): Promise<{ serviceOrderLine: ServiceOrderLine }> => {
      const response = await this.axios.post<
        SuccessEnvelope<{ serviceOrderLine: ServiceOrderLine }>
      >(endpoints.serviceOrders.lines.create(String(organizationId), String(id)), body);
      return withListMeta(response.data);
    },
  };

  async schedule(
    organizationId: number,
    id: number,
    body: ScheduleServiceOrderRequest,
  ): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.schedule(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async start(organizationId: number, id: number): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.start(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async complete(
    organizationId: number,
    id: number,
    body: CompleteServiceOrderRequest,
  ): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.complete(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async bill(
    organizationId: number,
    id: number,
    body: BillServiceOrderRequest,
  ): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.bill(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ serviceOrder: ServiceOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrder: ServiceOrder }>>(
      endpoints.serviceOrders.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class MaintenancePlans {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ maintenancePlans: MaintenancePlan[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ maintenancePlans: MaintenancePlan[] }>>(
      endpoints.maintenancePlans.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ maintenancePlan: MaintenancePlan }> {
    const response = await this.axios.get<SuccessEnvelope<{ maintenancePlan: MaintenancePlan }>>(
      endpoints.maintenancePlans.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateMaintenancePlanRequest,
  ): Promise<{ maintenancePlan: MaintenancePlan }> {
    const response = await this.axios.post<SuccessEnvelope<{ maintenancePlan: MaintenancePlan }>>(
      endpoints.maintenancePlans.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async generateOrders(
    organizationId: number,
    body: GenerateMaintenanceOrdersRequest,
  ): Promise<{ serviceOrders: ServiceOrder[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ serviceOrders: ServiceOrder[] }>>(
      endpoints.maintenancePlans.generateOrders(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}
