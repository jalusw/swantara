import type { AxiosInstance } from "axios";
import { endpoints } from "./endpoints";
import type {
  AccrueCommissionRequest,
  AccrueFromInvoiceCommissionRequest,
  AdvanceStageRequest,
  ClosePosSessionRequest,
  CommissionAssignment,
  CommissionEntry,
  CommissionPlan,
  CommissionRule,
  Coupon,
  CreateCommissionAssignmentRequest,
  CreateCommissionPlanRequest,
  CreateCommissionRuleRequest,
  CreateCouponRequest,
  CreateCrmActivityRequest,
  CreateCrmLeadRequest,
  CreatePosConfigRequest,
  CreateSaleOrderRequest,
  CrmActivity,
  CrmLead,
  CrmOpportunity,
  CrmStage,
  DeliverSaleOrderRequest,
  ForfeitExpiredGiftCardsRequest,
  GiftCard,
  GiftCardTransaction,
  InvoicePosOrderRequest,
  InvoiceSaleOrderRequest,
  InvoiceSummary,
  IssueGiftCardRequest,
  ListQuery,
  LoseOpportunityRequest,
  OpenPosSessionRequest,
  PayCommissionRequest,
  PaymentSummary,
  PaySaleOrderRequest,
  PipelineForecast,
  PosConfig,
  PosOrder,
  PosSession,
  PromoteLeadRequest,
  RedeemCouponRequest,
  RedeemGiftCardRequest,
  RefundGiftCardRequest,
  RefundPosOrderRequest,
  SaleOrder,
  SalesGroup,
  SellPosOrderRequest,
  SuccessEnvelope,
  UpdateCommissionPlanRequest,
  UpdateCouponRequest,
  UpdateCrmActivityRequest,
  UpdateCrmLeadRequest,
  UpdateSaleOrderRequest,
} from "./types";
import { withListMeta } from "./types";

export class SaleOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ orders: SaleOrder[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ orders: SaleOrder[] }>>(
      endpoints.saleOrders.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.get<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateSaleOrderRequest,
  ): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateSaleOrderRequest,
  ): Promise<{ order: SaleOrder }> {
    const response = await this.axios.put<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.saleOrders.delete(String(organizationId), String(id)));
  }

  async send(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.send(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async confirm(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.confirm(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async done(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.done(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async recomputeStatuses(organizationId: number, id: number): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.recomputeStatuses(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async deliver(
    organizationId: number,
    id: number,
    request: DeliverSaleOrderRequest,
  ): Promise<{ order: SaleOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: SaleOrder }>>(
      endpoints.saleOrders.deliver(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async invoice(
    organizationId: number,
    id: number,
    request: InvoiceSaleOrderRequest,
  ): Promise<{ invoice: InvoiceSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoice: InvoiceSummary }>>(
      endpoints.saleOrders.invoice(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async pay(
    organizationId: number,
    id: number,
    request: PaySaleOrderRequest,
  ): Promise<{ payment: PaymentSummary }> {
    const response = await this.axios.post<SuccessEnvelope<{ payment: PaymentSummary }>>(
      endpoints.saleOrders.pay(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export type CreateCrmOpportunityRequest = CreateCrmLeadRequest;

export class CrmLeads {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ leads: CrmLead[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ leads: CrmLead[] }>>(
      endpoints.crm.leads.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ prospect: CrmLead }> {
    const response = await this.axios.get<SuccessEnvelope<{ prospect: CrmLead }>>(
      endpoints.crm.leads.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateCrmLeadRequest,
  ): Promise<{ prospect: CrmLead }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmLead }>>(
      endpoints.crm.leads.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateCrmLeadRequest,
  ): Promise<{ prospect: CrmLead }> {
    const response = await this.axios.put<SuccessEnvelope<{ prospect: CrmLead }>>(
      endpoints.crm.leads.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.crm.leads.delete(String(organizationId), String(id)));
  }

  async promote(
    organizationId: number,
    id: number,
    request: PromoteLeadRequest,
  ): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.leads.promote(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class CrmOpportunities {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ opportunities: CrmOpportunity[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ opportunities: CrmOpportunity[] }>>(
      endpoints.crm.opportunities.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.get<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateCrmOpportunityRequest,
  ): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateCrmLeadRequest,
  ): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.put<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.crm.opportunities.delete(String(organizationId), String(id)));
  }

  async advanceStage(
    organizationId: number,
    id: number,
    request: AdvanceStageRequest,
  ): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.advanceStage(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async win(organizationId: number, id: number): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.win(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async lose(
    organizationId: number,
    id: number,
    request: LoseOpportunityRequest,
  ): Promise<{ prospect: CrmOpportunity }> {
    const response = await this.axios.post<SuccessEnvelope<{ prospect: CrmOpportunity }>>(
      endpoints.crm.opportunities.lose(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class CrmActivities {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ activities: CrmActivity[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ activities: CrmActivity[] }>>(
      endpoints.crm.activities.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ activity: CrmActivity }> {
    const response = await this.axios.get<SuccessEnvelope<{ activity: CrmActivity }>>(
      endpoints.crm.activities.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreateCrmActivityRequest,
  ): Promise<{ activity: CrmActivity }> {
    const response = await this.axios.post<SuccessEnvelope<{ activity: CrmActivity }>>(
      endpoints.crm.activities.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    request: UpdateCrmActivityRequest,
  ): Promise<{ activity: CrmActivity }> {
    const response = await this.axios.put<SuccessEnvelope<{ activity: CrmActivity }>>(
      endpoints.crm.activities.update(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.crm.activities.delete(String(organizationId), String(id)));
  }

  async done(organizationId: number, id: number): Promise<{ activity: CrmActivity }> {
    const response = await this.axios.post<SuccessEnvelope<{ activity: CrmActivity }>>(
      endpoints.crm.activities.done(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}

export class CrmPipeline {
  constructor(private readonly axios: AxiosInstance) {}

  async get(organizationId: number): Promise<{ pipeline: PipelineForecast }> {
    const response = await this.axios.get<SuccessEnvelope<{ pipeline: PipelineForecast }>>(
      endpoints.crm.pipeline(String(organizationId)),
    );
    return withListMeta(response.data);
  }
}

export class CrmStages {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ stages: CrmStage[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ stages: CrmStage[] }>>(
      endpoints.crm.stages(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }
}

export class SalesGroups {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ teams: SalesGroup[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ teams: SalesGroup[] }>>(
      endpoints.crm.teams(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }
}

export class PosConfigs {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ configs: PosConfig[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ configs: PosConfig[] }>>(
      endpoints.pos.configs.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ config: PosConfig }> {
    const response = await this.axios.get<SuccessEnvelope<{ config: PosConfig }>>(
      endpoints.pos.configs.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: CreatePosConfigRequest,
  ): Promise<{ config: PosConfig }> {
    const response = await this.axios.post<SuccessEnvelope<{ config: PosConfig }>>(
      endpoints.pos.configs.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class PosSessions {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ sessions: PosSession[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ sessions: PosSession[] }>>(
      endpoints.pos.sessions.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ session: PosSession }> {
    const response = await this.axios.get<SuccessEnvelope<{ session: PosSession }>>(
      endpoints.pos.sessions.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    request: OpenPosSessionRequest,
  ): Promise<{ session: PosSession }> {
    const response = await this.axios.post<SuccessEnvelope<{ session: PosSession }>>(
      endpoints.pos.sessions.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async closing(organizationId: number, id: number): Promise<{ session: PosSession }> {
    const response = await this.axios.post<SuccessEnvelope<{ session: PosSession }>>(
      endpoints.pos.sessions.closing(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async close(
    organizationId: number,
    id: number,
    request: ClosePosSessionRequest,
  ): Promise<{ session: PosSession }> {
    const response = await this.axios.post<SuccessEnvelope<{ session: PosSession }>>(
      endpoints.pos.sessions.close(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class PosOrders {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ orders: PosOrder[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ orders: PosOrder[] }>>(
      endpoints.pos.orders.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ order: PosOrder }> {
    const response = await this.axios.get<SuccessEnvelope<{ order: PosOrder }>>(
      endpoints.pos.orders.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, request: SellPosOrderRequest): Promise<{ order: PosOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PosOrder }>>(
      endpoints.pos.orders.create(String(organizationId)),
      request,
    );
    return withListMeta(response.data);
  }

  async invoice(
    organizationId: number,
    id: number,
    request: InvoicePosOrderRequest,
  ): Promise<{ invoiceId: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ invoiceId: number }>>(
      endpoints.pos.orders.invoice(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }

  async refund(
    organizationId: number,
    id: number,
    request: RefundPosOrderRequest,
  ): Promise<{ order: PosOrder }> {
    const response = await this.axios.post<SuccessEnvelope<{ order: PosOrder }>>(
      endpoints.pos.orders.refund(String(organizationId), String(id)),
      request,
    );
    return withListMeta(response.data);
  }
}

export class GiftCards {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ giftCards: GiftCard[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ giftCards: GiftCard[] }>>(
      endpoints.giftCards.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ giftCard: GiftCard }> {
    const response = await this.axios.get<SuccessEnvelope<{ giftCard: GiftCard }>>(
      endpoints.giftCards.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: IssueGiftCardRequest,
  ): Promise<{ giftCard: GiftCard }> {
    const response = await this.axios.post<SuccessEnvelope<{ giftCard: GiftCard }>>(
      endpoints.giftCards.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async forfeitExpired(
    organizationId: number,
    body: ForfeitExpiredGiftCardsRequest,
  ): Promise<{ giftCards: GiftCard[] }> {
    const response = await this.axios.post<SuccessEnvelope<{ giftCards: GiftCard[] }>>(
      endpoints.giftCards.forfeitExpired(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async redeem(
    organizationId: number,
    id: number,
    body: RedeemGiftCardRequest,
  ): Promise<{ giftCard: GiftCard }> {
    const response = await this.axios.post<SuccessEnvelope<{ giftCard: GiftCard }>>(
      endpoints.giftCards.redeem(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async refund(
    organizationId: number,
    id: number,
    body: RefundGiftCardRequest,
  ): Promise<{ giftCard: GiftCard }> {
    const response = await this.axios.post<SuccessEnvelope<{ giftCard: GiftCard }>>(
      endpoints.giftCards.refund(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async transactions(
    organizationId: number,
    id: number,
  ): Promise<{ giftCardTransactions: GiftCardTransaction[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ giftCardTransactions: GiftCardTransaction[] }>
    >(endpoints.giftCards.transactions(String(organizationId), String(id)));
    return withListMeta(response.data);
  }
}

export class Coupons {
  constructor(private readonly axios: AxiosInstance) {}

  async list(organizationId: number, params?: ListQuery): Promise<{ coupons: Coupon[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ coupons: Coupon[] }>>(
      endpoints.coupons.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ coupon: Coupon }> {
    const response = await this.axios.get<SuccessEnvelope<{ coupon: Coupon }>>(
      endpoints.coupons.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(organizationId: number, body: CreateCouponRequest): Promise<{ coupon: Coupon }> {
    const response = await this.axios.post<SuccessEnvelope<{ coupon: Coupon }>>(
      endpoints.coupons.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateCouponRequest,
  ): Promise<{ coupon: Coupon }> {
    const response = await this.axios.put<SuccessEnvelope<{ coupon: Coupon }>>(
      endpoints.coupons.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async delete(organizationId: number, id: number): Promise<void> {
    await this.axios.delete(endpoints.coupons.delete(String(organizationId), String(id)));
  }

  async redeem(
    organizationId: number,
    body: RedeemCouponRequest,
  ): Promise<{ coupon: Coupon; discount: number }> {
    const response = await this.axios.post<SuccessEnvelope<{ coupon: Coupon; discount: number }>>(
      endpoints.coupons.redeem(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }
}

export class CommissionPlans {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ commissionPlans: CommissionPlan[] }> {
    const response = await this.axios.get<SuccessEnvelope<{ commissionPlans: CommissionPlan[] }>>(
      endpoints.commissionPlans.list(String(organizationId)),
      { params },
    );
    return withListMeta(response.data);
  }

  async get(organizationId: number, id: number): Promise<{ commissionPlan: CommissionPlan }> {
    const response = await this.axios.get<SuccessEnvelope<{ commissionPlan: CommissionPlan }>>(
      endpoints.commissionPlans.get(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }

  async create(
    organizationId: number,
    body: CreateCommissionPlanRequest,
  ): Promise<{ commissionPlan: CommissionPlan }> {
    const response = await this.axios.post<SuccessEnvelope<{ commissionPlan: CommissionPlan }>>(
      endpoints.commissionPlans.create(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async update(
    organizationId: number,
    id: number,
    body: UpdateCommissionPlanRequest,
  ): Promise<{ commissionPlan: CommissionPlan }> {
    const response = await this.axios.put<SuccessEnvelope<{ commissionPlan: CommissionPlan }>>(
      endpoints.commissionPlans.update(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  rules = {
    list: async (
      organizationId: number,
      planId: number,
      params?: ListQuery,
    ): Promise<{ commissionRules: CommissionRule[] }> => {
      const response = await this.axios.get<SuccessEnvelope<{ commissionRules: CommissionRule[] }>>(
        endpoints.commissionPlans.rules.list(String(organizationId), String(planId)),
        { params },
      );
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      planId: number,
      body: CreateCommissionRuleRequest,
    ): Promise<{ commissionRule: CommissionRule }> => {
      const response = await this.axios.post<SuccessEnvelope<{ commissionRule: CommissionRule }>>(
        endpoints.commissionPlans.rules.create(String(organizationId), String(planId)),
        body,
      );
      return withListMeta(response.data);
    },
  };

  assignments = {
    list: async (
      organizationId: number,
      planId: number,
      params?: ListQuery,
    ): Promise<{ commissionAssignments: CommissionAssignment[] }> => {
      const response = await this.axios.get<
        SuccessEnvelope<{ commissionAssignments: CommissionAssignment[] }>
      >(endpoints.commissionPlans.assignments.list(String(organizationId), String(planId)), {
        params,
      });
      return withListMeta(response.data);
    },

    create: async (
      organizationId: number,
      planId: number,
      body: CreateCommissionAssignmentRequest,
    ): Promise<{ commissionAssignment: CommissionAssignment }> => {
      const response = await this.axios.post<
        SuccessEnvelope<{ commissionAssignment: CommissionAssignment }>
      >(endpoints.commissionPlans.assignments.create(String(organizationId), String(planId)), body);
      return withListMeta(response.data);
    },
  };
}

export class CommissionEntries {
  constructor(private readonly axios: AxiosInstance) {}

  async list(
    organizationId: number,
    params?: ListQuery,
  ): Promise<{ commissionEntries: CommissionEntry[] }> {
    const response = await this.axios.get<
      SuccessEnvelope<{ commissionEntries: CommissionEntry[] }>
    >(endpoints.commissionEntries.list(String(organizationId)), { params });
    return withListMeta(response.data);
  }

  async accrue(
    organizationId: number,
    body: AccrueCommissionRequest,
  ): Promise<{ commissionEntry: CommissionEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ commissionEntry: CommissionEntry }>>(
      endpoints.commissionEntries.accrue(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async accrueFromInvoice(
    organizationId: number,
    body: AccrueFromInvoiceCommissionRequest,
  ): Promise<{ commissionEntry: CommissionEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ commissionEntry: CommissionEntry }>>(
      endpoints.commissionEntries.accrueFromInvoice(String(organizationId)),
      body,
    );
    return withListMeta(response.data);
  }

  async pay(
    organizationId: number,
    id: number,
    body: PayCommissionRequest,
  ): Promise<{ commissionEntry: CommissionEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ commissionEntry: CommissionEntry }>>(
      endpoints.commissionEntries.pay(String(organizationId), String(id)),
      body,
    );
    return withListMeta(response.data);
  }

  async cancel(organizationId: number, id: number): Promise<{ commissionEntry: CommissionEntry }> {
    const response = await this.axios.post<SuccessEnvelope<{ commissionEntry: CommissionEntry }>>(
      endpoints.commissionEntries.cancel(String(organizationId), String(id)),
    );
    return withListMeta(response.data);
  }
}
