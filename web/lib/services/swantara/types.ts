export type Link = {
  rel: string;
  href: string;
  method?: string;
};

export type SuccessEnvelope<T> = {
  success: boolean;
  message: string;
  data: T;
  meta?: {
    timestamp: string;
    requestId?: string;
    version?: string;
    elapsed?: string;
    sort?: { field: string; direction: string };
    filter?: Record<string, string>;
    pagination?: {
      page: number;
      perPage: number;
      total: number;
      totalPages: number;
    };
  };
  links?: Link[];
};

export type ListQuery = {
  page?: number;
  size?: number;
  sort?: string;
  filter?: string[];
  format?: "json" | "xml" | "csv";
};

export type ListResult<T> = T & {
  meta?: SuccessEnvelope<unknown>["meta"];
};

export function withListMeta<T extends object>(envelope: SuccessEnvelope<T>): ListResult<T> {
  if (envelope.meta === undefined) return { ...envelope.data };
  return { ...envelope.data, meta: envelope.meta };
}

export type Permission = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
  code: string;
  description: string | null;
  action: string;
  resource: string;
};

export type SystemRole = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
  description: string | null;
  code: string;
  permissions?: Permission[];
};

export type PublicUser = {
  id: number;
  username: string;
  firstName: string;
  lastName: string | null;
  avatar: string | null;
  bio: string | null;
};

export type User = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  username: string;
  firstName: string;
  lastName: string | null;
  email: string;
  phone: string | null;
  avatar: string | null;
  bio: string | null;
  birthday: Date | null;
  active: boolean;
  private: boolean;
  sex: string | null;
  address: string | null;
  city: string | null;
  postalCode: string | null;
  systemRoleId: number;
  emailVerifiedAt: Date | null;
  phoneVerifiedAt: Date | null;
  systemRole?: SystemRole | null;
};

export type Profile = {
  id: number;
  username: string;
  firstName: string;
  lastName: string | null;
  email: string;
  phone: string | null;
  avatar: string | null;
  bio: string | null;
  birthday: Date | null;
  active: boolean;
  sex: string | null;
  address: string | null;
  city: string | null;
  postalCode: string | null;
  systemRoleId: number;
  emailVerifiedAt: Date | null;
  phoneVerifiedAt: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type UserSession = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  deviceName?: string | null;
  os?: string | null;
  browser?: string | null;
  userAgent?: string | null;
  ipAddress?: string | null;
  country?: string | null;
  city?: string | null;
  expiresAt?: Date | null;
  revokedAt?: Date | null;
  lastUsedAt?: Date | null;
  userId: number;
  user?: User | null;
};

export type Organization = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
  logo: string | null;
  legalName: string | null;
  parentId: number | null;
  baseCurrency: string;
  countryCode: string | null;
  taxId: string | null;
  timezone: string;
  taxYearStartMonth: number;
};

export type Member = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  userId: number;
  organizationId: number;
  position: string | null;
  user?: User | null;
  organization?: Organization | null;
  roles?: MemberRole[];
};

export type MemberRole = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number;
  name: string;
  code: string;
  description: string | null;
  permissions?: Permission[];
};

export type FxRate = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  currencyCode: string;
  organizationId: number | null;
  rate: number;
  rateType: "spot" | "avg" | "closing";
  validFrom: Date;
};

export type FxRateResponse = {
  id: number;
  currencyCode: string;
  organizationId: number | null;
  rate: number;
  rateType: "spot" | "avg" | "closing";
  validFrom: Date;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateFxRateRequest = {
  currencyCode: string;
  organizationId?: number | null;
  rate: number;
  rateType?: "spot" | "avg" | "closing";
  validFrom: string;
};

export type UpdateFxRateRequest = {
  rate: number;
  rateType?: "spot" | "avg" | "closing";
  validFrom: string;
};

export type ResolveFxRateRequest = {
  currencyCode: string;
  organizationId: number;
  rateType?: "spot" | "avg" | "closing";
  date: string;
};

export type ResolveFxRateResponse = {
  rate: number;
};

export type Currency = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  code: string;
  name: string;
  symbol: string | null;
  decimalPlaces: number;
  rounding: number;
};

export type CurrencyResponse = {
  code: string;
  name: string;
  symbol: string | null;
  decimalPlaces: number;
  rounding: number;
  createdAt: Date;
  updatedAt: Date;
};

export type AccountingStandard = {
  standardCode: string;
  standardName: string;
};

export type UnitCategory = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
};

export type UnitCategoryResponse = {
  id: number;
  name: string;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateUnitCategoryRequest = {
  name: string;
};

export type Unit = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  categoryId: number;
  name: string;
  factor: number;
  uomType: string;
  rounding: number;
};

export type UnitResponse = {
  id: number;
  categoryId: number;
  name: string;
  factor: number;
  uomType: string;
  rounding: number;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateUnitRequest = {
  categoryId: number;
  name: string;
  factor: number;
  uomType?: string;
  rounding?: number;
};

export type UpdateUnitRequest = {
  name: string;
  factor: number;
  uomType?: string;
  rounding?: number;
};

export type ConvertUnitRequest = {
  fromId: number;
  toId: number;
  qty: string;
};

export type ConvertUnitResponse = {
  value: number;
};

export type Dimension = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number | null;
  name: string;
  code: string | null;
  kind: string | null;
  parentId: number | null;
  active: boolean | null;
};

export type DimensionResponse = {
  id: number;
  organizationId: number | null;
  name: string;
  code: string | null;
  kind: string | null;
  parentId: number | null;
  active: boolean | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateDimensionRequest = {
  organizationId?: number | null;
  name: string;
  code?: string | null;
  kind?: string | null;
  parentId?: number | null;
  active?: boolean | null;
};

export type UpdateDimensionRequest = {
  name: string;
  code?: string | null;
  kind?: string | null;
  parentId?: number | null;
  active?: boolean | null;
};

export type PaymentTermLine = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  paymentTermId: number;
  sequence: number;
  valueType: "percent" | "fixed" | "balance";
  value: number;
  daysAfter: number;
  dayOfMonth: number | null;
  discountPct: number | null;
  discountDays: number | null;
};

export type PaymentTermLineResponse = {
  id: number;
  sequence: number;
  valueType: "percent" | "fixed" | "balance";
  value: number;
  daysAfter: number;
  dayOfMonth: number | null;
  discountPct: number | null;
  discountDays: number | null;
};

export type PaymentTerm = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
  note: string | null;
};

export type PaymentTermResponse = {
  id: number;
  name: string;
  note: string | null;
  lines: PaymentTermLineResponse[];
  createdAt: Date;
  updatedAt: Date;
};

export type CreatePaymentTermLineRequest = {
  sequence?: number;
  valueType: "percent" | "fixed" | "balance";
  value?: number;
  daysAfter?: number;
  dayOfMonth?: number | null;
  discountPct?: number | null;
  discountDays?: number | null;
};

export type CreatePaymentTermRequest = {
  name: string;
  note?: string | null;
  lines: CreatePaymentTermLineRequest[];
};

export type UpdatePaymentTermRequest = {
  name: string;
  note?: string | null;
  lines: CreatePaymentTermLineRequest[];
};

export type SplitPaymentTermRequest = {
  total: string;
  date: string;
};

export type PaymentSplit = {
  sequence: number;
  valueType: "percent" | "fixed" | "balance";
  amount: number;
  dueDate: Date;
  daysAfter: number;
  dayOfMonth?: number | null;
  discountPct?: number | null;
  discountDays?: number | null;
};

export type SplitPaymentTermResponse = {
  splits: PaymentSplit[];
};

export type Account = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number;
  code: string;
  name: string;
  type:
    | "asset"
    | "liability"
    | "equity"
    | "income"
    | "expense"
    | "receivable"
    | "payable"
    | "bank"
    | "cash"
    | "cogs"
    | "tax"
    | "current_asset"
    | "fixed_asset"
    | "depreciation";
  reconcilable: boolean;
  currencyCode: string | null;
  parentId: number | null;
  active: boolean;
};

export type AccountResponse = {
  id: number;
  organizationId: number;
  code: string;
  name: string;
  type: Account["type"];
  reconcilable: boolean;
  currencyCode: string | null;
  parentId: number | null;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateAccountRequest = {
  organizationId?: number | null;
  code: string;
  name: string;
  type: string;
  reconcilable?: boolean;
  currencyCode?: string | null;
  parentId?: number | null;
  active?: boolean | null;
};

export type UpdateAccountRequest = {
  code: string;
  name: string;
  type: string;
  reconcilable?: boolean;
  currencyCode?: string | null;
  parentId?: number | null;
  active?: boolean | null;
};

export type Journal = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number;
  name: string;
  code: string | null;
  type: "sale" | "purchase" | "bank" | "cash" | "general";
  defaultAccountId: number | null;
  currencyCode: string | null;
  bankAccountId: number | null;
  sequenceId: number | null;
};

export type JournalResponse = {
  id: number;
  organizationId: number;
  code: string | null;
  name: string;
  type: "sale" | "purchase" | "bank" | "cash" | "general";
  defaultAccountId: number | null;
  bankAccountId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateJournalRequest = {
  organizationId?: number | null;
  code?: string | null;
  name: string;
  type: string;
  defaultAccountId?: number | null;
  bankAccountId?: number | null;
};

export type UpdateJournalRequest = {
  code?: string | null;
  name: string;
  type: string;
  defaultAccountId?: number | null;
  bankAccountId?: number | null;
};

export type Tax = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number | null;
  name: string;
  amount: number | null;
  type: "percent" | "fixed" | "group";
  scope: "sale" | "purchase" | "none";
  priceInclude: boolean;
  taxAccountId: number | null;
  refundTaxAccountId: number | null;
  active: boolean;
};

export type TaxResponse = {
  id: number;
  organizationId: number | null;
  name: string;
  amount: number | null;
  type: "percent" | "fixed" | "group";
  scope: "sale" | "purchase" | "none";
  priceInclude: boolean;
  taxAccountId: number | null;
  refundTaxAccountId: number | null;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateTaxRequest = {
  organizationId?: number | null;
  name: string;
  amount?: number | null;
  type: string;
  scope: string;
  priceInclude?: boolean;
  taxAccountId?: number | null;
  refundTaxAccountId?: number | null;
};

export type UpdateTaxRequest = {
  name: string;
  amount?: number | null;
  type: string;
  scope: string;
  priceInclude?: boolean;
  taxAccountId?: number | null;
  refundTaxAccountId?: number | null;
  active?: boolean;
};

export type TaxYear = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  organizationId: number | null;
  name: string;
  dateStart: Date | null;
  dateEnd: Date | null;
  state: string | null;
};

export type TaxYearResponse = {
  id: number;
  organizationId: number | null;
  name: string;
  dateStart: string | null;
  dateEnd: string | null;
  state: string | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateTaxYearRequest = {
  organizationId?: number | null;
  name: string;
  dateStart: string;
  dateEnd: string;
};

export type UpdateTaxYearRequest = {
  name: string;
  dateStart: string;
  dateEnd: string;
  state?: string | null;
};

export type Carrier = {
  id: number;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
  name: string;
  trackingUrlTpl: string | null;
  deliveryItemId: number | null;
};

export type CarrierResponse = {
  id: number;
  name: string;
  trackingUrlTpl: string | null;
  deliveryItemId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateCarrierRequest = {
  name: string;
  trackingUrlTpl?: string | null;
  deliveryItemId?: number | null;
};

export type UpdateCarrierRequest = {
  name: string;
  trackingUrlTpl?: string | null;
  deliveryItemId?: number | null;
};

export type SystemConfig = {
  id: number;
  organizationId: number | null;
  key: string;
  value: unknown;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateSystemConfigRequest = {
  key: string;
  value: unknown;
};

export type UpdateSystemConfigRequest = {
  value: unknown;
};

export type IntegrationEvent = {
  id: number;
  topic: string;
  payload: unknown;
  status: "pending" | "sent" | "failed";
  retries: number;
  createdAt: Date;
  updatedAt: Date;
};

export type AuditLog = {
  id: number;
  tableName: string;
  recordId: number;
  action: "insert" | "update" | "delete";
  changedBy: number;
  changedAt: Date;
  diff: Record<string, unknown>;
  createdAt: Date;
  updatedAt: Date;
};

export type Contact = {
  id: number;
  organizationId: number | null;
  name: string;
  displayName: string | null;
  isOrganization: boolean;
  parentId: number | null;
  email: string | null;
  phone: string | null;
  mobile: string | null;
  website: string | null;
  taxId: string | null;
  industry: string | null;
  currencyCode: string | null;
  lang: string;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type ContactAddressRequest = {
  type: "billing" | "shipping" | "other" | null;
  line1: string | null;
  line2: string | null;
  city: string | null;
  state: string | null;
  postalCode: string | null;
  countryCode: string | null;
  isDefault: boolean;
};

export type ContactBankAccountRequest = {
  accountHolder: string | null;
  bankName: string | null;
  iban: string | null;
  swiftBic: string | null;
  accountNumber: string | null;
  routingNumber: string | null;
  currencyCode: string | null;
};

export type CustomerProfileRequest = {
  customerPaymentTermId: number | null;
  creditLimit: number | null;
  receivableAccountId: number | null;
  active: boolean;
};

export type SupplierProfileRequest = {
  vendorPaymentTermId: number | null;
  payableAccountId: number | null;
  active: boolean;
};

export type CreateContactRequest = {
  organizationId: number | null;
  name: string;
  displayName: string | null;
  isOrganization: boolean;
  parentId: number | null;
  email: string | null;
  phone: string | null;
  mobile: string | null;
  website: string | null;
  taxId: string | null;
  industry: string | null;
  currencyCode: string | null;
  lang: string;
  active: boolean;
  addresses: ContactAddressRequest[];
  bankAccounts: ContactBankAccountRequest[];
  customer: CustomerProfileRequest | null;
  supplier: SupplierProfileRequest | null;
};

export type UpdateContactRequest = {
  organizationId: number | null;
  name: string;
  displayName: string | null;
  isOrganization: boolean;
  parentId: number | null;
  email: string | null;
  phone: string | null;
  mobile: string | null;
  website: string | null;
  taxId: string | null;
  industry: string | null;
  currencyCode: string | null;
  lang: string;
  active: boolean;
};

export type ContactAddress = {
  id: number;
  contactId: number;
  type: "billing" | "shipping" | "other" | null;
  line1: string | null;
  line2: string | null;
  city: string | null;
  state: string | null;
  postalCode: string | null;
  countryCode: string | null;
  isDefault: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type ContactBankAccount = {
  id: number;
  contactId: number;
  accountHolder: string | null;
  bankName: string | null;
  iban: string | null;
  swiftBic: string | null;
  accountNumber: string | null;
  routingNumber: string | null;
  currencyCode: string | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CustomerProfile = {
  contactId: number;
  customerPaymentTermId: number | null;
  creditLimit: number | null;
  receivableAccountId: number | null;
  active: boolean;
};

export type SupplierProfile = {
  contactId: number;
  vendorPaymentTermId: number | null;
  payableAccountId: number | null;
  active: boolean;
};

export type ItemCategory = {
  id: number;
  name: string;
  parentId: number | null;
  incomeAccountId: number | null;
  expenseAccountId: number | null;
  stockCostAccountId: number | null;
  stockInputAccountId: number | null;
  stockOutputAccountId: number | null;
  cogsAccountId: number | null;
  costMethod: "standard" | "fifo" | "average" | null;
  valuation: "manual" | "automated" | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateItemCategoryRequest = {
  name: string;
  parentId: number | null;
  incomeAccountId: number | null;
  expenseAccountId: number | null;
  stockCostAccountId: number | null;
  stockInputAccountId: number | null;
  stockOutputAccountId: number | null;
  cogsAccountId: number | null;
  costMethod: "standard" | "fifo" | "average" | null;
  valuation: "manual" | "automated" | null;
};

export type Item = {
  id: number;
  organizationId: number | null;
  name: string;
  categoryId: number | null;
  type: "stockable" | "consumable" | "service" | "digital";
  unitId: number | null;
  purchaseUnitId: number | null;
  listPrice: number;
  standardCost: number;
  isPurchasable: boolean;
  isSellable: boolean;
  isManufactured: boolean;
  tracking: "none" | "batch" | "serial";
  weight: number;
  volume: number;
  hsCode: string | null;
  descriptionSale: string | null;
  descriptionPurchase: string | null;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type ItemVariant = {
  id: number;
  templateId: number;
  sku: string | null;
  barcode: string | null;
  attributeJson: Record<string, unknown>;
  extraCost: number;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type AttributeOption = {
  name: string;
  values: string[];
};

export type CreateVariantRequest = {
  sku: string | null;
  barcode: string | null;
  attributeJson: Record<string, unknown>;
  extraCost: number;
  active: boolean | null;
};

export type CreateProductRequest = {
  organizationId: number | null;
  name: string;
  categoryId: number | null;
  type: "stockable" | "consumable" | "service" | "digital";
  unitId: number | null;
  purchaseUnitId: number | null;
  listPrice: number;
  standardCost: number;
  isPurchasable: boolean | null;
  isSellable: boolean | null;
  isManufactured: boolean;
  tracking: "none" | "batch" | "serial";
  weight: number;
  volume: number;
  hsCode: string | null;
  descriptionSale: string | null;
  descriptionPurchase: string | null;
  active: boolean | null;
  variants: CreateVariantRequest[];
  attributeMatrix: AttributeOption[];
};

export type UpdateProductRequest = {
  name: string;
  categoryId: number | null;
  type: "stockable" | "consumable" | "service" | "digital";
  unitId: number | null;
  purchaseUnitId: number | null;
  listPrice: number;
  standardCost: number;
  isPurchasable: boolean | null;
  isSellable: boolean | null;
  isManufactured: boolean;
  tracking: "none" | "batch" | "serial";
  weight: number;
  volume: number;
  hsCode: string | null;
  descriptionSale: string | null;
  descriptionPurchase: string | null;
  active: boolean | null;
};

export type PriceBook = {
  id: number;
  name: string;
  currencyCode: string | null;
  organizationId: number | null;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type CreatePriceBookRequest = {
  name: string;
  currencyCode: string | null;
  organizationId: number | null;
  active: boolean | null;
};

export type PriceRule = {
  id: number;
  priceBookId: number;
  appliesTo: "all" | "category" | "item" | "variant";
  itemId: number | null;
  categoryId: number | null;
  minQty: number;
  computeType: "fixed" | "percent" | "formula";
  fixedPrice: number | null;
  discountPct: number | null;
  dateStart: Date | null;
  dateEnd: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreatePriceRuleRequest = {
  appliesTo: "all" | "category" | "item" | "variant";
  itemId: number | null;
  categoryId: number | null;
  minQty: number;
  computeType: "fixed" | "percent" | "formula";
  fixedPrice: number | null;
  discountPct: number | null;
  dateStart: Date | null;
  dateEnd: Date | null;
};

export type ResolvePriceRequest = {
  variantId: number;
  qty: string;
  date: string;
};

export type ResolvePriceResponse = {
  price: number;
  basePrice: number;
  templateId: number;
  variantId: number;
  ruleId?: number;
  appliesTo?: "all" | "category" | "item" | "variant";
  computeType?: "fixed" | "percent" | "formula";
};

export type SupplierProduct = {
  id: number;
  itemId: number;
  supplierId: number;
  vendorSku: string | null;
  vendorProductName: string | null;
  minQty: number;
  price: number | null;
  currencyCode: string | null;
  leadTimeDays: number | null;
  priority: number;
  validFrom: Date | null;
  validTo: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateSupplierProductRequest = {
  itemId: number;
  supplierId: number;
  vendorSku: string | null;
  vendorProductName: string | null;
  minQty: number;
  price: number | null;
  currencyCode: string | null;
  leadTimeDays: number | null;
  priority: number;
  validFrom: Date | null;
  validTo: Date | null;
};

export type UpdateSupplierProductRequest = {
  supplierId: number;
  vendorSku: string | null;
  vendorProductName: string | null;
  minQty: number;
  price: number | null;
  currencyCode: string | null;
  leadTimeDays: number | null;
  priority: number;
  validFrom: Date | null;
  validTo: Date | null;
};

export type BestSupplierOfferRequest = {
  itemId: number;
  qty: string;
  date: string;
};

export type Recipe = {
  id: number;
  organizationId: number | null;
  itemId: number;
  code: string | null;
  qty: number;
  unitId: number | null;
  type: "manufacture" | "kit" | "subcontract";
  version: number;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type BomLine = {
  id: number;
  bomId: number;
  componentId: number;
  qty: number;
  unitId: number | null;
  scrapPct: number;
  productionStepId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type BomComponentRequirement = {
  componentId: number;
  qty: number;
  unitId?: number | null;
  scrapPct: number;
};

export type CreateBomLineRequest = {
  componentId: number;
  qty: number;
  unitId?: number | null;
  scrapPct?: number;
  productionStepId?: number | null;
};

export type CreateBomRequest = {
  organizationId?: number | null;
  itemId: number;
  code?: string | null;
  qty?: number;
  unitId?: number | null;
  type: "manufacture" | "kit" | "subcontract";
  version?: number;
  lines: CreateBomLineRequest[];
};

export type UpdateBomRequest = {
  organizationId?: number | null;
  itemId: number;
  code?: string | null;
  qty?: number;
  unitId?: number | null;
  type: "manufacture" | "kit" | "subcontract";
  version?: number;
  active?: boolean | null;
};

export type ExplodeBomRequest = {
  qty: string;
};

export type ProductionOrder = {
  id: number;
  organizationId: number | null;
  name: string | null;
  itemId: number;
  bomId: number | null;
  qtyToProduce: number;
  qtyProduced: number;
  unitId: number | null;
  srcLocationId: number | null;
  dstLocationId: number | null;
  state: "draft" | "confirmed" | "planned" | "in_progress" | "done" | "cancelled";
  datePlannedStart: Date | null;
  datePlannedFinish: Date | null;
  dateStart: Date | null;
  dateFinished: Date | null;
  origin: "manual" | "sale_order" | "reorder" | null;
  priority: number;
  createdAt: Date;
  updatedAt: Date;
};

export type MOComponent = {
  id: number;
  productionOrderId: number;
  itemId: number;
  qtyPlanned: number;
  qtyConsumed: number;
  unitId: number | null;
  stockMovementId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateProductionOrderRequest = {
  organizationId?: number | null;
  itemId: number;
  bomId: number;
  qtyToProduce: number;
  unitId?: number | null;
  srcLocationId: number;
  dstLocationId: number;
  datePlannedStart?: Date | null;
  datePlannedFinish?: Date | null;
  origin?: "manual" | "sale_order" | "reorder" | null;
  priority?: number;
};

export type ConsumeMaterialRequest = {
  componentId: number;
  qty: number;
  journalId: number;
  wipAccountId: number;
  date?: string;
};

export type ProduceGoodsRequest = {
  qty: number;
  journalId: number;
  wipAccountId: number;
  date?: string;
};

export type RecordLaborRequest = {
  minutes: number;
  journalId: number;
  wipAccountId: number;
  appliedLaborAccountId: number;
  date?: string;
};

export type SettleVarianceRequest = {
  journalId: number;
  wipAccountId: number;
  varianceAccountId: number;
  date?: string;
};

export type ShopTask = {
  id: number;
  organizationId: number | null;
  productionOrderId: number;
  productionStepId: number | null;
  workCenterId: number;
  name: string | null;
  state: "draft" | "confirmed" | "planned" | "in_progress" | "done" | "cancelled";
  sequence: number;
  plannedStart: Date | null;
  plannedFinish: Date | null;
  dateStart: Date | null;
  dateFinished: Date | null;
  plannedMinutes: number;
  actualMinutes: number;
  createdAt: Date;
  updatedAt: Date;
};

export type PlanningRun = {
  id: number;
  organizationId: number | null;
  runDate: Date | null;
  horizonDays: number;
  state: "running" | "done" | "failed";
  demands?: PlanningNeed[];
  plannedOrders?: PlannedSupply[];
  createdAt: Date;
  updatedAt: Date;
};

export type PlanningNeed = {
  id: number;
  planningRunId: number;
  itemId: number;
  warehouseId: number | null;
  sourceType: "sale_order" | "forecast" | "reorder" | "recipe";
  sourceId: number;
  qty: number;
  requiredDate: Date | null;
};

export type PlannedSupply = {
  id: number;
  planningRunId: number;
  itemId: number;
  warehouseId: number | null;
  type: "purchase" | "manufacture" | "transfer";
  qty: number;
  orderDate: Date | null;
  dueDate: Date | null;
  peggedDemandId: number | null;
  confirmed: boolean;
  generatedDocType: string | null;
  generatedDocId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type RunMrpRequest = {
  horizonDays: number;
  runDate?: string;
};

export type ConfirmPlannedRequest = {
  supplierId: number;
};

export type DemandPlan = {
  id: number;
  organizationId: number | null;
  itemId: number;
  warehouseId: number | null;
  periodStart: Date | null;
  periodEnd: Date | null;
  forecastQty: number;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateForecastRequest = {
  itemId: number;
  warehouseId?: number | null;
  periodStart: string;
  periodEnd: string;
  forecastQty: number;
};

export type OutsideProcessingOrder = {
  id: number;
  productionOrderId: number;
  supplierId: number;
  purchaseOrderId: number | null;
  state: "draft" | "sent" | "received" | "done" | "cancelled";
  createdAt: Date;
  updatedAt: Date;
};

export type CreateOutsideProcessingOrderRequest = {
  productionOrderId: number;
  supplierId: number;
};

export type SendOutsideProcessingOrderRequest = {
  journalId: number;
  wipAccountId: number;
  date?: string | null;
};

export type ReceiveOutsideProcessingOrderRequest = {
  journalId: number;
  wipAccountId: number;
  apPayableAccountId: number;
  date?: string | null;
};

export type JournalEntry = {
  id: number;
  organizationId: number;
  journalId: number;
  name: string | null;
  date: Date;
  ref: string | null;
  state: "draft" | "posted" | "cancelled";
  currencyCode: string | null;
  originType: string | null;
  originId: number | null;
  reversedMoveId: number | null;
  postedAt: Date | null;
  postedBy: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type JournalLine = {
  id: number;
  entryId: number;
  accountId: number;
  contactId: number | null;
  name: string | null;
  debit: number;
  credit: number;
  currencyCode: string | null;
  amountCurrency: number;
  dimensionId: number | null;
  taxId: number | null;
  reconciled: boolean;
  fullReconcileId: number | null;
  dueDate: Date | null;
};

export type PostingLineRequest = {
  accountId: number;
  dimensionId: number | null;
  name: string;
  debit: number;
  credit: number;
};

export type PostJournalEntryRequest = {
  organizationId: number | null;
  journalId: number;
  date: string;
  ref: string;
  originType: string;
  originId: number;
  description: string;
  lines: PostingLineRequest[];
};

export type ReverseJournalEntryRequest = {
  journalId: number;
  date: string;
  ref: string;
  description: string;
};

export type Invoice = {
  id: number;
  organizationId: number | null;
  entryId: number | null;
  type: "customer_invoice" | "customer_credit_note" | "vendor_bill" | "vendor_credit_note";
  contactId: number;
  name: string | null;
  reference: string | null;
  invoiceDate: Date | null;
  dueDate: Date | null;
  currencyCode: string | null;
  journalId: number | null;
  paymentTermId: number | null;
  state: "draft" | "posted" | "cancelled";
  paymentState: "not_paid" | "in_payment" | "partial" | "paid" | "reversed";
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
  amountResidual: number;
  createdAt: Date;
  updatedAt: Date;
};

export type InvoiceLine = {
  id: number;
  invoiceId: number;
  sequence: number;
  itemId: number | null;
  description: string | null;
  qty: number;
  unitId: number | null;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  accountId: number | null;
  dimensionId: number | null;
  priceSubtotal: number;
};

export type InvoiceTax = {
  id: number;
  invoiceId: number;
  taxId: number | null;
  base: number;
  amount: number;
  accountId: number | null;
};

export type InvoiceLineRequest = {
  itemId: number | null;
  description: string;
  qty: number;
  unitId: number | null;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  accountId: number;
  dimensionId: number | null;
};

export type CreateInvoiceRequest = {
  organizationId: number | null;
  journalId: number;
  contactId: number;
  date: string | null;
  dueDate: string | null;
  reference: string;
  lines: InvoiceLineRequest[];
};

export type CreateCreditNoteRequest = {
  organizationId: number | null;
  journalId: number;
  date: string | null;
  dueDate: string | null;
  reference: string;
};

export type Payment = {
  id: number;
  organizationId: number | null;
  name: string | null;
  contactId: number;
  type: "inbound" | "outbound";
  journalId: number | null;
  paymentMethod: string | null;
  amount: number;
  currencyCode: string | null;
  date: Date;
  reference: string | null;
  entryId: number | null;
  contactBankAccountId: number | null;
  state: "draft" | "posted" | "reconciled" | "cancelled";
  createdAt: Date;
  updatedAt: Date;
};

export type PaymentAllocation = {
  id: number;
  paymentId: number;
  invoiceId: number;
  amount: number;
};

export type CreatePaymentRequest = {
  organizationId: number | null;
  contactId: number;
  journalId: number;
  amount: number;
  date: string | null;
  reference: string;
  invoiceIds: number[];
};

export type TaxPeriod = {
  id: number;
  organizationId: number;
  taxYearId: number;
  name: string;
  dateStart: string | null;
  dateEnd: string | null;
  state: "open" | "closed" | "locked";
  createdAt: Date;
  updatedAt: Date;
};

export type CreateTaxPeriodRequest = {
  organizationId: number | null;
  taxYearId: number;
  name: string;
  dateStart: string;
  dateEnd: string;
  state: string;
};

export type BankStatement = {
  id: number;
  journalId: number | null;
  name: string | null;
  date: string | null;
  balanceStart: number;
  balanceEnd: number;
  state: "draft" | "open" | "reconciled" | "cancelled";
  createdAt: Date;
  updatedAt: Date;
};

export type BankStatementLine = {
  id: number;
  statementId: number;
  date: string | null;
  amount: number;
  contactId: number | null;
  ref: string | null;
  narration: string | null;
  reconciled: boolean;
  paymentId: number | null;
  moveLineId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateBankStatementLineRequest = {
  date: string | null;
  amount: number;
  contactId: number | null;
  ref: string;
  narration: string;
};

export type CreateBankStatementRequest = {
  organizationId: number | null;
  journalId: number;
  date: string | null;
  balanceStart: number;
  balanceEnd: number;
  lines: CreateBankStatementLineRequest[];
};

export type ReconcileRequest = {
  organizationId: number | null;
  debitLineId: number;
  creditLineId: number;
  amount: number;
};

export type AccountPartialReconcile = {
  id: number;
  debitLineId: number;
  creditLineId: number;
  amount: number;
  fullReconcileId: number | null;
};

export type ReminderAction = {
  id: number;
  contactId: number;
  invoiceId: number;
  levelId: number;
  sentAt: string | null;
  channel: string | null;
};

export type GenerateReminderRequest = {
  organizationId: number | null;
  asOf: string | null;
};

export type Budget = {
  id: number;
  organizationId: number | null;
  name: string | null;
  dateStart: string | null;
  dateEnd: string | null;
  state: "draft" | "approved" | "confirmed" | "cancelled";
  createdAt: Date;
  updatedAt: Date;
};

export type BudgetLine = {
  id: number;
  budgetId: number;
  accountId: number;
  dimensionId: number | null;
  plannedAmount: number;
  practicalAmount: number;
};

export type CreateBudgetLineRequest = {
  accountId: number;
  dimensionId: number | null;
  plannedAmount: number;
};

export type CreateBudgetRequest = {
  organizationId: number | null;
  name: string;
  dateStart: string;
  dateEnd: string;
  lines: CreateBudgetLineRequest[];
};

export type BudgetVarianceLine = {
  accountId: number;
  dimensionId: number | null;
  plannedAmount: number;
  practicalAmount: number;
  variance: number;
};

export type TaxRule = {
  id: number;
  organizationId: number | null;
  name: string | null;
  countryCode: string | null;
  autoApply: boolean;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type TaxRuleTaxMap = {
  id: number;
  taxRuleId: number;
  srcTaxId: number;
  destTaxId: number | null;
};

export type TaxRuleAccountMap = {
  id: number;
  taxRuleId: number;
  srcAccountId: number;
  destAccountId: number;
};

export type CreateTaxRuleTaxMapRequest = {
  srcTaxId: number;
  destTaxId: number | null;
};

export type CreateTaxRuleAccountMapRequest = {
  srcAccountId: number;
  destAccountId: number;
};

export type CreateTaxRuleRequest = {
  organizationId: number | null;
  name: string;
  countryCode: string;
  autoApply: boolean;
  taxMaps: CreateTaxRuleTaxMapRequest[];
  accountMaps: CreateTaxRuleAccountMapRequest[];
};

export type ResolveTaxRuleRequest = {
  taxId: number | null;
  accountId: number;
};

export type ResolveTaxRuleData = {
  taxId: number | null;
  accountId: number | null;
};

export type WithholdingTax = {
  id: number;
  organizationId: number | null;
  name: string | null;
  ratePct: number;
  accountId: number | null;
  scope: "sale" | "purchase";
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateWithholdingTaxRequest = {
  organizationId: number | null;
  name: string;
  ratePct: number;
  accountId: number;
  scope: "sale" | "purchase";
};

export type WithholdRequest = {
  organizationId: number | null;
  journalId: number;
  contactId: number;
  amount: number;
  date: string | null;
  ref: string;
  scope: "sale" | "purchase";
  withholdingTaxId: number;
  bankAccountId: number;
  payableAccountId: number;
  description: string;
};

export type TaxReturn = {
  id: number;
  organizationId: number | null;
  periodId: number;
  type: "sale" | "purchase";
  outputTax: number;
  inputTax: number;
  netPayable: number;
  state: "draft" | "filed" | "paid";
  filedAt: string | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateTaxReturnRequest = {
  organizationId: number | null;
  periodId: number;
  type: "sale" | "purchase";
};

export type DeferralSchedule = {
  id: number;
  organizationId: number | null;
  type: "deferred_revenue" | "deferred_expense" | "prepaid";
  sourceType: string;
  sourceId: number;
  totalAmount: number;
  balanceSheetAccountId: number | null;
  plAccountId: number | null;
  method: "linear" | "manual" | "milestone";
  dateStart: Date | null;
  state: "draft" | "running" | "done" | "cancelled";
  recognizedAmount: number;
};

export type DeferralLine = {
  id: number;
  scheduleId: number;
  sequence: number;
  recognitionDate: Date | null;
  amount: number;
  posted: boolean;
  entryId: number | null;
};

export type RecognizeDeferralsRequest = {
  asOf: string;
};

export type CreateDeferralLineRequest = {
  recognitionDate: Date | null;
  amount: number;
};

export type CreateDeferralScheduleRequest = {
  type: "deferred_revenue" | "deferred_expense" | "prepaid";
  sourceType: string;
  sourceId: number;
  contactId: number | null;
  itemId: number | null;
  totalAmount: number;
  balanceSheetAccountId: number;
  plAccountId: number;
  dimensionId: number | null;
  method: "linear" | "manual" | "milestone";
  dateStart: Date | null;
  dateEnd: Date | null;
  periods: number;
  lines: CreateDeferralLineRequest[];
};

export type Warehouse = {
  id: number;
  organizationId: number | null;
  name: string;
  code: string | null;
  line1: string | null;
  line2: string | null;
  city: string | null;
  state: string | null;
  postalCode: string | null;
  countryCode: string | null;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
};

export type CreateWarehouseRequest = {
  organizationId: number | null;
  name: string;
  code: string | null;
  line1: string | null;
  line2: string | null;
  city: string | null;
  state: string | null;
  postalCode: string | null;
  countryCode: string | null;
};

export type StockLocation = {
  id: number;
  warehouseId: number | null;
  name: string;
  code: string | null;
  parentId: number | null;
  usage: string;
  barcode: string | null;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
};

export type CreateStockLocationRequest = {
  organizationId: number | null;
  warehouseId: number | null;
  name: string;
  code: string | null;
  parentId: number | null;
  usage: string;
  barcode: string | null;
};

export type OnHand = {
  itemId: number;
  locationId: number;
  onHand: number;
};

export type AvailableToPromise = {
  itemId: number;
  locationId: number;
  available: number;
};

export type StockBalance = {
  id: number;
  itemId: number;
  locationId: number;
  batchId: number | null;
  quantity: number;
  reservedQty: number;
  createdAt: Date;
  updatedAt: Date;
};

export type StockMovement = {
  id: number;
  organizationId: number | null;
  shipmentId: number | null;
  itemId: number;
  qty: number;
  unitId: number | null;
  srcLocationId: number;
  dstLocationId: number;
  batchId: number | null;
  state: "draft" | "confirmed" | "assigned" | "done" | "cancelled";
  unitCost: number | null;
  originType: string | null;
  originId: number | null;
  scheduledDate: Date | null;
  dateDone: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateStockMovementRequest = {
  organizationId: number | null;
  itemId: number;
  qty: string;
  unitId: number | null;
  srcLocationId: number;
  dstLocationId: number;
  batchId: number | null;
  unitCost: number | null;
  originType: string | null;
  originId: number | null;
  scheduledDate: string | null;
};

export type ReceiveMoveRequest = {
  unitCost: string;
  journalId: number;
  date: string;
};

export type ShipMoveRequest = {
  journalId: number;
  date: string;
};

export type CostLayer = {
  id: number;
  movementId: number | null;
  itemId: number;
  quantity: number;
  unitCost: number | null;
  value: number;
  remainingQty: number;
  remainingValue: number;
  journalEntryId: number | null;
  description: string | null;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date | null;
};

export type Shipment = {
  id: number;
  organizationId: number | null;
  name: string | null;
  type: "incoming" | "outgoing" | "internal";
  contactId: number | null;
  srcLocationId: number | null;
  dstLocationId: number | null;
  state: "draft" | "waiting" | "confirmed" | "assigned" | "done" | "cancelled";
  scheduledDate: Date | null;
  dateDone: Date | null;
  origin: string | null;
  carrierId: number | null;
  trackingRef: string | null;
  createdAt: Date;
  updatedAt: Date;
};

export type StockHold = {
  id: number;
  movementId: number | null;
  balanceId: number;
  qty: number;
  createdAt: Date;
  updatedAt: Date;
};

export type ReserveStockRequest = {
  itemId: number;
  locationId: number;
  batchId: number | null;
  qty: string;
  movementId: number | null;
};

export type ReleaseByMovementRequest = {
  entryId: number;
};

export type Batch = {
  id: number;
  itemId: number;
  name: string;
  ref: string | null;
  expiryDate: Date | null;
  bestBeforeDate: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateBatchRequest = {
  itemId: number;
  name: string;
  ref: string | null;
  expiryDate: Date | null;
  bestBeforeDate: Date | null;
};

export type ReorderRule = {
  id: number;
  itemId: number;
  warehouseId: number | null;
  locationId: number | null;
  minQty: number;
  maxQty: number;
  qtyMultiple: number;
  leadTimeDays: number | null;
  active: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type ReorderRuleRequest = {
  itemId: number;
  warehouseId: number | null;
  locationId: number | null;
  minQty: number;
  maxQty: number;
  qtyMultiple: number;
  leadTimeDays: number | null;
  active: boolean | null;
};

export type ReplenishmentCandidate = {
  ruleId: number;
  itemId: number;
  locationId: number;
  onHand: number;
  minQty: number;
  maxQty: number;
  recommendedQty: number;
};

export type StockCount = {
  id: number;
  organizationId: number | null;
  name: string | null;
  locationId: number | null;
  state: "draft" | "posted";
  countDate: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CountLineRequest = {
  itemId: number;
  batchId: number | null;
  countedQty: number;
};

export type CreateStockCountRequest = {
  organizationId: number | null;
  name: string | null;
  locationId: number | null;
  countDate: string | null;
  lines: CountLineRequest[];
};

export type PostStockCountRequest = {
  journalId: number;
  gainLossAccountId: number;
  date: string;
};

export type StockCountLine = {
  id: number;
  countId: number;
  itemId: number;
  batchId: number | null;
  theoreticalQty: number;
  countedQty: number;
  diffQty: number;
};

export type WarehouseTransfer = {
  id: number;
  organizationId: number | null;
  name: string | null;
  srcWarehouseId: number;
  dstWarehouseId: number;
  state: "draft" | "sent" | "in_transit" | "received" | "cancelled";
  outShipmentId: number | null;
  inShipmentId: number | null;
  isInterorganization: boolean;
  scheduledDate: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type TransferLine = {
  itemId: number;
  qty: number;
  batchId: number | null;
  srcLocationId: number | null;
  dstLocationId: number | null;
};

export type CreateWarehouseTransferRequest = {
  organizationId: number | null;
  name: string | null;
  srcWarehouseId: number;
  dstWarehouseId: number;
  scheduledDate: string | null;
  lines: TransferLine[];
};

export type TransferActionRequest = {
  journalId: number;
  transitAccountId: number;
  date: string;
};

export type InboundCost = {
  id: number;
  name: string;
  date: Date | null;
  state: "draft" | "posted" | "cancelled";
  targetShipmentIds: number[];
  movementId: number | null;
};

export type CreateInboundCostLineRequest = {
  itemId: number;
  description: string;
  amount: number;
  vendorBillLineId: number | null;
  splitMethod: "by_quantity" | "by_weight" | "by_volume" | "by_value" | "equal";
  accountId: number | null;
};

export type CreateInboundCostRequest = {
  name: string;
  date: string | null;
  targetShipmentIds: number[];
  lines: CreateInboundCostLineRequest[];
};

export type InboundCostLine = {
  id: number;
  itemId: number;
  description: string | null;
  amount: number;
  vendorBillLineId: number | null;
  splitMethod: "by_quantity" | "by_weight" | "by_volume" | "by_value" | "equal";
  accountId: number | null;
};

export type InboundCostAdjustment = {
  id: number;
  stockMovementId: number;
  itemId: number;
  additionalCost: number;
  valuationLayerId: number;
};

export type CrmLead = {
  id: number;
  organizationId: number | null;
  name: string;
  type: "lead" | "opportunity";
  contactId: number | null;
  contactName: string | null;
  email: string | null;
  phone: string | null;
  jobPosition: string | null;
  stageId: number | null;
  stageName?: string | null;
  expectedRevenue: number;
  probability: number;
  priority: number;
  salespersonId: number | null;
  salesGroupId: number | null;
  source: string | null;
  medium: string | null;
  campaign: string | null;
  lostReason: string | null;
  expectedClose: Date | null;
  closedAt: Date | null;
  isWon: boolean;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateCrmLeadRequest = {
  organizationId: number | null;
  name: string;
  type: "lead" | "opportunity";
  contactId: number | null;
  contactName: string | null;
  email: string | null;
  phone: string | null;
  jobPosition: string | null;
  stageId: number | null;
  expectedRevenue: number;
  probability: number;
  priority: number;
  salespersonId: number | null;
  salesGroupId: number | null;
  source: string | null;
  medium: string | null;
  campaign: string | null;
  expectedClose: string | null;
};

export type UpdateCrmLeadRequest = CreateCrmLeadRequest;

export type PromoteLeadRequest = {
  stageId: number;
  salespersonId: number | null;
  salesGroupId: number | null;
  expectedRevenue: number;
  probability: number | null;
  priority: number;
  expectedClose: string | null;
};

export type CrmOpportunity = CrmLead;

export type AdvanceStageRequest = { stageId: number };
export type LoseOpportunityRequest = { lostReason: string };

export type CrmActivity = {
  id: number;
  prospectId: number | null;
  contactId: number | null;
  type: "call" | "meeting" | "email" | "note" | "task";
  summary: string;
  note: string | null;
  dueDate: Date | null;
  done: boolean;
  doneAt: Date | null;
  userId: number | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateCrmActivityRequest = {
  prospectId: number | null;
  contactId: number | null;
  type: "call" | "meeting" | "email" | "note" | "task";
  summary: string;
  note: string | null;
  dueDate: string | null;
  done: boolean | null;
};

export type UpdateCrmActivityRequest = CreateCrmActivityRequest;

export type PipelineStage = {
  stageId: number;
  stageName: string;
  probability: number;
  opportunityCount: number;
  expectedRevenue: number;
  weightedRevenue: number;
};

export type PipelineForecast = {
  totalExpectedRevenue: number;
  weightedPipeline: number;
  winRate: number;
  stages: PipelineStage[];
};

export type CrmStage = {
  id: number;
  name: string;
  sequence: number;
  isWon: boolean;
  probability: number;
};

export type SalesGroup = {
  id: number;
  name: string;
  leaderId: number | null;
  organizationId: number | null;
};

export type SaleOrder = {
  id: number;
  organizationId: number | null;
  name: string | null;
  contactId: number;
  shipAddressId: number | null;
  billAddressId: number | null;
  priceBookId: number | null;
  currencyCode: string | null;
  salespersonId: number | null;
  salesGroupId: number | null;
  crmLeadId: number | null;
  warehouseId: number | null;
  state: "draft" | "sent" | "confirmed" | "done" | "cancelled";
  orderDate: Date | null;
  expectedDate: Date | null;
  validityDate: Date | null;
  paymentTermId: number | null;
  incoterm: string | null;
  customerPoRef: string | null;
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
  invoiceStatus: "no" | "to_invoice" | "invoiced";
  deliveryStatus: "pending" | "partial" | "done";
  note: string | null;
  lines?: SaleOrderLine[];
  createdAt: Date;
  updatedAt: Date;
};

export type SaleOrderLine = {
  id: number;
  orderId: number;
  sequence: number;
  itemId: number | null;
  description: string | null;
  qtyOrdered: number;
  qtyDelivered: number;
  qtyInvoiced: number;
  qtyReturns: number;
  unitId: number | null;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  dimensionId: number | null;
  priceSubtotal: number;
  priceTax: number;
  priceTotal: number;
};

export type SaleOrderLineRequest = {
  sequence: number;
  itemId: number | null;
  description: string | null;
  qtyOrdered: number;
  unitId: number | null;
  discountPct: number;
  taxIds: number[];
  dimensionId: number | null;
};

export type CreateSaleOrderRequest = {
  organizationId: number | null;
  contactId: number;
  shipAddressId: number | null;
  billAddressId: number | null;
  priceBookId: number | null;
  salespersonId: number | null;
  salesGroupId: number | null;
  crmLeadId: number | null;
  warehouseId: number | null;
  orderDate: string | null;
  expectedDate: string | null;
  validityDate: string | null;
  paymentTermId: number | null;
  incoterm: string | null;
  customerPoRef: string | null;
  note: string | null;
  lines: SaleOrderLineRequest[];
};

export type UpdateSaleOrderRequest = CreateSaleOrderRequest;

export type DeliverSaleOrderRequest = {
  journalId: number;
  date: string | null;
};

export type InvoiceSaleOrderRequest = {
  journalId: number;
  date: string | null;
};

export type PaySaleOrderRequest = {
  journalId: number;
  amount: number;
  date: string | null;
};

export type InvoiceSummary = {
  id: number;
  type: string;
  contactId: number;
  name: string | null;
  state: string;
  paymentState: string;
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
  amountResidual: number;
  invoiceDate: Date | null;
};

export type PaymentSummary = {
  id: number;
  name: string | null;
  contactId: number;
  type: string;
  amount: number;
  date: Date;
  state: string;
};

export type PosConfig = {
  id: number;
  organizationId: number | null;
  name: string;
  warehouseId: number | null;
  journalId: number | null;
  priceBookId: number | null;
};

export type CreatePosConfigRequest = {
  organizationId: number | null;
  name: string;
  warehouseId: number | null;
  journalId: number | null;
  priceBookId: number | null;
};

export type PosSession = {
  id: number;
  configId: number;
  cashierId: number;
  openedAt: Date | null;
  closedAt: Date | null;
  openingBalance: number;
  closingBalance: number | null;
  state: "opened" | "closing" | "closed";
};

export type PosPaymentMethod = { method: string; amount: number };

export type OpenPosSessionRequest = {
  configId: number;
  cashierId: number;
  openingBalance: number;
};

export type ClosePosSessionRequest = { closingBalance: number };

export type PosOrder = {
  id: number;
  sessionId: number;
  contactId: number | null;
  name: string | null;
  amountTotal: number;
  amountTax: number;
  state: "done" | "refunded";
  invoiceId: number | null;
  orderTime: Date | null;
  lines?: PosOrderLine[];
  payments?: PosPayment[];
};

export type PosOrderLine = {
  id: number;
  orderId: number;
  itemId: number | null;
  qty: number;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  priceSubtotal: number;
  priceTax: number;
  priceTotal: number;
};

export type PosPayment = {
  id: number;
  orderId: number;
  method: string;
  amount: number;
};

export type PosOrderLineRequest = {
  itemId: number;
  qty: number;
  discountPct: number;
  taxIds: number[];
};

export type PosPaymentRequest = { method: string; amount: number };

export type SellPosOrderRequest = {
  sessionId: number;
  contactId: number | null;
  lines: PosOrderLineRequest[];
  payments: PosPaymentRequest[];
};

export type InvoicePosOrderRequest = { journalId: number; date: string };
export type RefundPosOrderRequest = { journalId: number; date: string };

export type ApprovalRequest = {
  id: number;
  ownerType: string;
  ownerId: number;
  requestedBy: number;
  state: "pending" | "approved" | "refused";
  steps?: ApprovalStep[];
  createdAt: Date;
  updatedAt: Date;
};

export type ApprovalStep = {
  id: number;
  approverId: number;
  sequence: number;
  decision: "pending" | "approved" | "refused";
  decidedAt: Date | null;
  comment: string | null;
};

export type CreateApprovalRequestRequest = {
  ownerType: string;
  ownerId: number;
  requestedBy: number;
  approverIds: number[];
};

export type DecideApprovalRequestRequest = {
  stepId: number;
  approve: boolean;
  comment: string;
};

export type Attachment = {
  id: number;
  ownerType: string;
  ownerId: number;
  filename: string;
  mimeType: string;
  byteSize: number;
  storageUrl: string;
  checksum: string;
  uploadedBy: number;
  uploadedAt: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type UploadAttachmentRequest = {
  ownerType: string;
  ownerId: number;
  file: File;
};

export type Message = {
  id: number;
  ownerType: string;
  ownerId: number;
  authorId: number;
  body: string;
  messageType: "note" | "system";
  createdAt: Date;
  updatedAt: Date;
};

export type CreateMessageRequest = {
  ownerType: string;
  ownerId: number;
  body: string;
  messageType: "note" | "system";
};

export type QualityPoint = {
  id: number;
  organizationId: number | null;
  itemId: number | null;
  operation: string | null;
  testType: "pass_fail" | "measure" | "instruction";
  normMin: number | null;
  normMax: number | null;
  unitId: number | null;
};

export type CreateQualityPointRequest = {
  organizationId: number | null;
  itemId: number | null;
  operation: string | null;
  testType: "pass_fail" | "measure" | "instruction";
  normMin: number | null;
  normMax: number | null;
  unitId: number | null;
};

export type QualityCheck = {
  id: number;
  pointId: number | null;
  itemId: number | null;
  batchId: number | null;
  shipmentId: number | null;
  productionOrderId: number | null;
  measuredValue: number | null;
  result: "pending" | "pass" | "fail";
  checkedBy: number | null;
  checkedAt: Date | null;
};

export type RecordQualityCheckRequest = {
  pass: boolean;
  measuredValue: number | null;
  checkedBy: number | null;
};

export type RouteToScrapRequest = {
  journalId: number;
  expenseAccountId: number;
};

export type QualityAlert = {
  id: number;
  itemId: number | null;
  batchId: number | null;
  checkId: number | null;
  title: string | null;
  description: string | null;
  severity: string | null;
  state: "open" | "in_progress" | "solved" | "cancelled";
  assignedTo: number | null;
};

export type UpdateQualityAlertRequest = {
  state: "open" | "in_progress" | "solved" | "cancelled";
};

export type PurchaseRequest = {
  id: number;
  organizationId: number | null;
  name: string | null;
  requesterId: number;
  departmentId: number | null;
  state: "draft" | "confirmed" | "approved" | "done" | "cancelled";
  neededBy: Date | null;
  lines?: PurchaseRequestLine[];
};

export type PurchaseRequestLine = {
  id: number;
  requestId: number;
  itemId: number | null;
  description: string | null;
  qty: number;
  unitId: number | null;
  neededBy: Date | null;
};

export type PurchaseRequestLineRequest = {
  itemId: number | null;
  description: string | null;
  qty: number;
  unitId: number | null;
  neededBy: string | null;
};

export type CreatePurchaseRequestRequest = {
  organizationId: number | null;
  requesterId: number;
  departmentId: number | null;
  neededBy: string | null;
  lines: PurchaseRequestLineRequest[];
};

export type PurchaseOrder = {
  id: number;
  organizationId: number | null;
  name: string | null;
  supplierId: number;
  vendorRef: string | null;
  currencyCode: string | null;
  warehouseId: number | null;
  destLocationId: number | null;
  state: "draft" | "sent" | "confirmed" | "done" | "cancelled";
  orderDate: Date | null;
  expectedDate: Date | null;
  paymentTermId: number | null;
  incoterm: string | null;
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
  invoiceStatus: "no" | "to_invoice" | "invoiced";
  receiptStatus: "pending" | "partial" | "done";
  lines?: PurchaseOrderLine[];
  createdAt: Date;
  updatedAt: Date;
};

export type PurchaseOrderLine = {
  id: number;
  orderId: number;
  sequence: number;
  itemId: number | null;
  description: string | null;
  qtyOrdered: number;
  qtyReceived: number;
  qtyBilled: number;
  qtyReturns: number;
  unitId: number | null;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  dimensionId: number | null;
  priceSubtotal: number;
};

export type PurchaseOrderLineRequest = {
  sequence: number;
  itemId: number | null;
  description: string | null;
  qtyOrdered: number;
  unitId: number | null;
  unitPrice: number;
  discountPct: number;
  taxIds: number[];
  dimensionId: number | null;
};

export type CreatePurchaseOrderRequest = {
  organizationId: number | null;
  requestId: number | null;
  supplierId: number;
  vendorRef: string | null;
  currencyCode: string | null;
  warehouseId: number | null;
  destLocationId: number | null;
  orderDate: string | null;
  expectedDate: string | null;
  paymentTermId: number | null;
  incoterm: string | null;
  lines: PurchaseOrderLineRequest[];
};

export type ConfirmPurchaseOrderRequest = {
  byUserId: number | null;
};

export type ReceivePurchaseOrderRequest = {
  journalId: number;
  date: string | null;
};

export type BillPurchaseOrderRequest = {
  journalId: number;
  date: string | null;
  override: boolean;
};

export type PayPurchaseOrderRequest = {
  journalId: number;
  date: string | null;
};

export type PurchaseInvoiceSummary = {
  id: number;
  type: "customer_invoice" | "customer_credit_note" | "vendor_bill" | "vendor_credit_note";
  contactId: number;
  name: string | null;
  state: "draft" | "posted" | "cancelled";
  paymentState: "not_paid" | "in_payment" | "partial" | "paid" | "reversed";
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
  amountResidual: number;
  invoiceDate: Date | null;
};

export type PurchasePaymentSummary = {
  id: number;
  name: string | null;
  contactId: number;
  type: "inbound" | "outbound";
  amount: number;
  date: Date;
  state: "draft" | "posted" | "reconciled" | "cancelled";
};

export type SupplierQuoteRequest = {
  id: number;
  organizationId: number | null;
  name: string | null;
  requesterId: number;
  supplierId: number | null;
  currencyCode: string | null;
  state: "draft" | "sent" | "done" | "cancelled";
  orderDate: Date | null;
  quoteDeadline: Date | null;
  notes: string | null;
};

export type SupplierQuoteRequestLine = {
  id: number;
  quoteRequestId: number;
  itemId: number | null;
  description: string | null;
  qty: number;
  unitId: number | null;
  neededBy: Date | null;
};

export type SupplierQuoteRequestLineRequest = {
  itemId: number | null;
  description: string | null;
  qty: number;
  unitId: number | null;
  neededBy: Date | null;
};

export type CreateSupplierQuoteRequestRequest = {
  organizationId: number | null;
  requesterId: number;
  supplierId: number | null;
  currencyCode: string | null;
  orderDate: Date | null;
  quoteDeadline: Date | null;
  notes: string | null;
  lines: SupplierQuoteRequestLineRequest[];
};

export type CreateSupplierQuoteRequestFromRequisitionRequest = {
  requestId: number;
};

export type SupplierQuote = {
  id: number;
  quoteRequestId: number;
  supplierId: number;
  currencyCode: string | null;
  state: "draft" | "submitted" | "accepted" | "rejected";
  quoteDate: Date | null;
  validUntil: Date | null;
  notes: string | null;
  amountUntaxed: number;
  amountTax: number;
  amountTotal: number;
};

export type SupplierQuoteLineRequest = {
  quoteRequestLineId: number;
  itemId: number | null;
  description: string | null;
  qty: number;
  unitPrice: number;
  discountPct: number;
};

export type SubmitSupplierQuoteRequest = {
  supplierId: number;
  currencyCode: string | null;
  validUntil: Date | null;
  notes: string | null;
  lines: SupplierQuoteLineRequest[];
};

export type CreatePurchaseOrderFromQuoteRequestRequest = {
  supplierId: number | null;
  currencyCode: string | null;
  warehouseId: number | null;
  vendorRef: string | null;
  orderDate: Date | null;
  expectedDate: Date | null;
  paymentTermId: number | null;
  incoterm: string | null;
};

export type Rma = {
  id: number;
  organizationId: number | null;
  name: string | null;
  type: "customer_return" | "vendor_return";
  contactId: number;
  originOrderType: string | null;
  originOrderId: number | null;
  reason: string | null;
  state: "draft" | "confirmed" | "received" | "refunded" | "done" | "cancelled";
  createdAt: Date;
  updatedAt: Date;
};

export type RmaLine = {
  id: number;
  rmaId: number;
  itemId: number;
  qty: number;
  batchId: number | null;
  disposition: "restock" | "scrap" | "repair" | "replace";
  stockMovementId: number | null;
  creditNoteId: number | null;
};

export type CreateRmaLineRequest = {
  itemId: number;
  qty: number;
  batchId: number | null;
  disposition: "restock" | "scrap" | "repair" | "replace";
};

export type CreateRmaRequest = {
  organizationId: number | null;
  type: "customer_return" | "vendor_return";
  contactId: number;
  originOrderType: string;
  originOrderId: number;
  reason: string;
  lines: CreateRmaLineRequest[];
};

export type ReceiveRmaRequest = {
  journalId: number;
  date: string;
};

export type RefundRmaRequest = {
  journalId: number;
  date: string;
  reference: string;
};

export type Department = {
  id: number;
  organizationId: number | null;
  name: string;
  description: string | null;
  parentId: number | null;
  managerId: number | null;
  dimensionId: number | null;
};

export type CreateDepartmentRequest = {
  organizationId: number | null;
  name: string;
  description: string | null;
  parentId: number | null;
  managerId: number | null;
  dimensionId: number | null;
};

export type JobPosition = {
  id: number;
  organizationId: number | null;
  name: string;
  departmentId: number | null;
};

export type CreateJobPositionRequest = {
  organizationId: number | null;
  name: string;
  departmentId: number | null;
};

export type LeaveType = {
  id: number;
  organizationId: number | null;
  name: string;
  paid: boolean;
  allocationDays: number | null;
};

export type CreateLeaveTypeRequest = {
  organizationId: number | null;
  name: string;
  paid: boolean;
  allocationDays: number | null;
};

export type Employee = {
  id: number;
  organizationId: number | null;
  contactId: number;
  userId: number | null;
  employeeNumber: string;
  departmentId: number | null;
  jobPositionId: number | null;
  managerId: number | null;
  hireDate: string | null;
  terminationDate: string | null;
  employmentType: "full_time" | "part_time" | "contract";
  workLocation: string | null;
  active: boolean;
};

export type CreateEmployeeRequest = {
  organizationId: number | null;
  name: string;
  email: string | null;
  phone: string | null;
  userId: number | null;
  employeeNumber: string;
  departmentId: number | null;
  jobPositionId: number | null;
  managerId: number | null;
  hireDate: string | null;
  employmentType: "full_time" | "part_time" | "contract";
  workLocation: string | null;
  wage: number;
  wageType: "monthly" | "hourly";
  currencyCode: string;
};

export type UpdateEmployeeRequest = {
  name: string;
  employeeNumber: string;
  userId: number | null;
  departmentId: number | null;
  jobPositionId: number | null;
  managerId: number | null;
  hireDate: string | null;
  terminationDate: string | null;
  employmentType: "full_time" | "part_time" | "contract";
  workLocation: string | null;
  active: boolean;
};

export type EmploymentContract = {
  id: number;
  employeeId: number;
  dateStart: string;
  dateEnd: string | null;
  wage: number;
  wageType: "monthly" | "hourly";
  currencyCode: string;
  state: "active" | "closed";
};

export type CreateContractRequest = {
  employeeId: number;
  dateStart: string | null;
  dateEnd: string | null;
  wage: number;
  wageType: "monthly" | "hourly";
  currencyCode: string;
};

export type UpdateContractRequest = {
  dateStart: string | null;
  dateEnd: string | null;
  wage: number;
  wageType: "monthly" | "hourly";
  currencyCode: string;
  state: "active" | "closed";
};

export type LeaveRequest = {
  id: number;
  employeeId: number;
  leaveTypeId: number;
  dateFrom: string;
  dateTo: string;
  days: number;
  state: "draft" | "submitted" | "approved" | "refused";
};

export type CreateLeaveRequestRequest = {
  employeeId: number;
  leaveTypeId: number;
  dateFrom: string;
  dateTo: string;
  days: number;
};

export type LeaveBalance = {
  employeeId: number;
  leaveTypeId: number;
  balance: number;
};

export type Attendance = {
  id: number;
  employeeId: number;
  checkIn: string | null;
  checkOut: string | null;
  workedHours: number;
};

export type CheckInRequest = {
  employeeId: number;
  checkIn: string;
};

export type CheckOutRequest = {
  checkOut: string;
};

export type Timesheet = {
  id: number;
  employeeId: number;
  date: string;
  projectId: number | null;
  taskId: number | null;
  dimensionId: number | null;
  hours: number;
  description: string | null;
};

export type CreateTimesheetRequest = {
  employeeId: number;
  date: string;
  projectId: number | null;
  taskId: number | null;
  dimensionId: number | null;
  hours: number;
  description: string | null;
};

export type SalaryRule = {
  id: number;
  organizationId: number | null;
  code: string;
  name: string;
  category: "earning" | "deduction" | null;
  computeType: "fixed" | "percent" | "formula" | null;
  amount: number | null;
  formula: string | null;
  accountDebitId: number | null;
  accountCreditId: number | null;
};

export type CreateSalaryRuleRequest = {
  organizationId: number | null;
  code: string;
  name: string;
  category: "earning" | "deduction";
  computeType: "fixed" | "percent" | "formula";
  amount: number | null;
  formula: string | null;
  accountDebitId: number | null;
  accountCreditId: number | null;
};

export type UpdateSalaryRuleRequest = {
  code: string;
  name: string;
  category: "earning" | "deduction";
  computeType: "fixed" | "percent" | "formula";
  amount: number | null;
  formula: string | null;
  accountDebitId: number | null;
  accountCreditId: number | null;
};

export type PayrollRun = {
  id: number;
  organizationId: number;
  name: string | null;
  periodStart: string;
  periodEnd: string;
  state: "draft" | "confirmed" | "paid" | "closed";
};

export type CreatePayrollRunRequest = {
  organizationId: number;
  periodStart: string;
  periodEnd: string;
};

export type ConfirmPayrollRunRequest = {
  journalId: number;
  date: string;
};

export type PayPayrollRunRequest = {
  journalId: number;
  date: string;
};

export type Payslip = {
  id: number;
  runId: number;
  employeeId: number;
  contractId: number;
  gross: number;
  net: number;
  entryId: number | null;
  state: "draft" | "posted";
  lines: PayslipLine[];
};

export type PayslipLine = {
  id: number;
  payslipId: number;
  ruleId: number;
  code: string;
  name: string;
  category: "earning" | "deduction";
  amount: number;
};

export type PayrollRunDetail = {
  run: PayrollRun;
  payslips: Payslip[];
};

export type Project = {
  id: number;
  organizationId: number;
  name: string;
  contactId: number;
  managerId: number | null;
  dimensionId: number | null;
  saleOrderId: number | null;
  billingType: "fixed" | "time_material" | "milestone";
  billableRate: number;
  dateStart: Date | null;
  dateEnd: Date | null;
  state: "draft" | "open" | "closed" | "cancelled";
};

export type CreateProjectRequest = {
  organizationId: number | null;
  name: string;
  contactId: number;
  managerId: number | null;
  dimensionId: number | null;
  saleOrderId: number | null;
  billingType: "fixed" | "time_material" | "milestone";
  billableRate: number;
  dateStart: Date | null;
  dateEnd: Date | null;
};

export type UpdateProjectRequest = {
  name: string;
  contactId: number;
  managerId: number | null;
  dimensionId: number | null;
  saleOrderId: number | null;
  billingType: "fixed" | "time_material" | "milestone";
  billableRate: number;
  dateStart: Date | null;
  dateEnd: Date | null;
};

export type SetProjectStateRequest = {
  state: "draft" | "open" | "closed" | "cancelled";
};

export type ProjectSummary = {
  projectId: number;
  plannedHours: number;
  effectiveHours: number;
  utilization: number;
  billedAmount: number;
  unbilledAmount: number;
  billableAmount: number;
  costAmount: number;
  marginAmount: number;
};

export type BillTimeMaterialRequest = {
  journalId: number;
  date: string;
  taxIds: number[];
};

export type ProjectTask = {
  id: number;
  projectId: number;
  name: string;
  assigneeId: number | null;
  stage: "backlog" | "todo" | "in_progress" | "done";
  plannedHours: number;
  effectiveHours: number;
  parentTaskId: number | null;
  deadline: Date | null;
  priority: number | null;
};

export type CreateProjectTaskRequest = {
  name: string;
  assigneeId: number | null;
  stage: "backlog" | "todo" | "in_progress" | "done";
  plannedHours: number;
  parentTaskId: number | null;
  deadline: Date | null;
  priority: number | null;
};

export type UpdateProjectTaskRequest = CreateProjectTaskRequest;

export type ProjectMilestone = {
  id: number;
  projectId: number;
  name: string;
  deadline: Date | null;
  reached: boolean;
  saleLineId: number | null;
};

export type CreateProjectMilestoneRequest = {
  name: string;
  deadline: Date | null;
  saleLineId: number | null;
};

export type SetMilestoneReachedRequest = {
  reached: boolean;
};

export type ExpenseCategory = {
  id: number;
  organizationId: number | null;
  name: string;
  expenseAccountId: number | null;
  defaultTaxIds: number[];
};

export type CreateExpenseCategoryRequest = {
  organizationId: number | null;
  name: string;
  expenseAccountId: number | null;
  defaultTaxIds: number[];
};

export type UpdateExpenseCategoryRequest = {
  name: string;
  expenseAccountId: number | null;
  defaultTaxIds: number[];
};

export type ExpenseReport = {
  id: number;
  name: string;
  employeeId: number;
  state: "draft" | "submitted" | "approved" | "refused" | "posted" | "reimbursed";
  paymentMode: "own_account" | "organization_account";
  totalAmount: number;
  entryId: number | null;
  submittedAt: Date | null;
  approvedBy: number | null;
  lines?: ExpenseLine[];
};

export type ExpenseLine = {
  id: number;
  categoryId: number | null;
  itemId: number | null;
  description: string | null;
  expenseDate: Date | null;
  quantity: number;
  unitPrice: number;
  amount: number;
  taxIds: number[];
  dimensionId: number | null;
  projectId: number | null;
  reimbursable: boolean;
  receiptAttachmentId: number | null;
};

export type CreateExpenseLineRequest = {
  categoryId: number | null;
  itemId: number | null;
  description: string;
  expenseDate: string;
  quantity: number;
  unitPrice: number;
  taxIds: number[];
  currencyCode: string;
  dimensionId: number | null;
  projectId: number | null;
  reimbursable: boolean | null;
  receiptAttachmentId: number | null;
};

export type CreateExpenseReportRequest = {
  name: string;
  employeeId: number;
  paymentMode: "own_account" | "organization_account";
  lines: CreateExpenseLineRequest[];
};

export type AssetCategory = {
  id: number;
  organizationId: number | null;
  name: string;
  assetAccountId: number | null;
  depreciationAccountId: number | null;
  expenseAccountId: number | null;
  gainAccountId: number | null;
  lossAccountId: number | null;
  method: "linear" | "declining" | "declining_then_linear" | null;
  methodNumber: number | null;
  methodPeriod: "month" | "year" | null;
};

export type CreateAssetCategoryRequest = {
  organizationId: number | null;
  name: string;
  assetAccountId: number | null;
  depreciationAccountId: number | null;
  expenseAccountId: number | null;
  gainAccountId: number | null;
  lossAccountId: number | null;
  method: "linear" | "declining" | "declining_then_linear";
  methodNumber: number | null;
  methodPeriod: "month" | "year";
};

export type FixedAsset = {
  id: number;
  organizationId: number;
  name: string;
  categoryId: number;
  purchaseValue: number;
  salvageValue: number;
  acquisitionDate: Date | null;
  inServiceDate: Date | null;
  originalEntryId: number | null;
  invoiceLineId: number | null;
  state: "draft" | "running" | "disposed" | "sold";
  disposalDate: Date | null;
};

export type RegisterAssetRequest = {
  organizationId: number | null;
  name: string;
  categoryId: number;
  purchaseValue: number;
  salvageValue: number;
  acquisitionDate: string;
  inServiceDate: string;
  invoiceLineId: number;
};

export type AssetDepreciationLine = {
  id: number;
  assetId: number;
  sequence: number;
  depreciationDate: Date;
  amount: number;
  accumulated: number;
  remainingValue: number;
  entryId: number | null;
  posted: boolean;
};

export type PostDepreciationRequest = {
  journalId: number;
  date: string;
};

export type DisposeAssetRequest = {
  journalId: number;
  date: string;
  state: "disposed" | "sold";
  proceedsAmount: number;
  proceedsAccountId: number | null;
};

export type SubscriptionPlan = {
  id: number;
  name: string;
  recurringInterval: string;
  recurringCount: number;
};

export type CreateSubscriptionPlanRequest = {
  name: string;
  recurringInterval: string;
  recurringCount: number;
};

export type UpdateSubscriptionPlanRequest = CreateSubscriptionPlanRequest;

export type Subscription = {
  id: number;
  organizationId: number | null;
  name: string;
  contactId: number | null;
  planId: number | null;
  priceBookId: number | null;
  currencyCode: string | null;
  dateStart: Date | null;
  nextInvoiceDate: Date | null;
  dateEnd: Date | null;
  state: "draft" | "active" | "paused" | "churned" | "closed";
  mrr: number;
  lines?: SubscriptionLine[];
};

export type SubscriptionLine = {
  id: number;
  itemId: number | null;
  qty: number;
  unitPrice: number;
  discountPct: number;
};

export type CreateSubscriptionLineRequest = {
  itemId: number;
  qty: number;
  unitPrice: number;
  discountPct: number;
};

export type CreateSubscriptionRequest = {
  name: string;
  contactId: number;
  planId: number;
  priceBookId: number;
  currencyCode: string;
  lines: CreateSubscriptionLineRequest[];
};

export type SubscriptionMetrics = {
  mrr: number;
  arr: number;
  churned: number;
  churnRate: number;
  ltv: number;
};

export type CommissionPlan = {
  id: number;
  organizationId: number | null;
  name: string;
  basis: "revenue" | "margin" | "collected";
  active: boolean;
};

export type CreateCommissionPlanRequest = {
  organizationId?: number | null;
  name: string;
  basis: "revenue" | "margin" | "collected";
};

export type UpdateCommissionPlanRequest = {
  active?: boolean | null;
};

export type CommissionRule = {
  id: number;
  planId: number;
  itemCategoryId: number | null;
  minAmount: number;
  maxAmount: number;
  ratePct: number;
  fixedAmount: number;
};

export type CreateCommissionRuleRequest = {
  itemCategoryId?: number | null;
  minAmount: number;
  maxAmount: number;
  ratePct: number;
  fixedAmount: number;
};

export type CommissionAssignment = {
  id: number;
  planId: number;
  salespersonId: number;
  dateStart: Date | null;
  dateEnd: Date | null;
};

export type CreateCommissionAssignmentRequest = {
  salespersonId: number;
  dateStart: string;
  dateEnd?: string | null;
};

export type CommissionEntry = {
  id: number;
  salespersonId: number;
  planId: number;
  sourceType: string;
  sourceId: number;
  baseAmount: number;
  commissionAmount: number;
  state: "draft" | "confirmed" | "paid" | "cancelled";
  periodId: number | null;
  payslipId: number | null;
};

export type AccrueCommissionRequest = {
  salespersonId: number;
  planId: number;
  sourceType: string;
  sourceId: number;
  itemCategoryId?: number | null;
  baseAmount: number;
  journalId: number;
  expenseAccountId: number;
  payableAccountId: number;
  date?: string;
  periodId?: number | null;
};

export type AccrueFromInvoiceCommissionRequest = {
  invoiceId: number;
  salespersonId: number;
  journalId: number;
  expenseAccountId: number;
  payableAccountId: number;
  date?: string;
  periodId?: number | null;
};

export type PayCommissionRequest = { payslipId?: number | null };

export type GiftCard = {
  id: number;
  organizationId: number | null;
  code: string;
  contactId: number | null;
  initialAmount: number;
  balance: number;
  currencyCode: string;
  expiryDate: Date | null;
  state: "active" | "used" | "expired" | "cancelled";
  issuedFromOrderId: number | null;
};

export type IssueGiftCardRequest = {
  organizationId?: number | null;
  code?: string;
  contactId?: number | null;
  amount: number;
  currencyCode: string;
  expiryDate?: string | null;
  issuedFromOrderId?: number | null;
  journalId: number;
  cashAccountId: number;
  liabilityAccountId: number;
  date?: string;
};

export type RedeemGiftCardRequest = {
  amount: number;
  orderType?: string;
  orderId?: number;
  journalId: number;
  revenueAccountId: number;
  liabilityAccountId: number;
  date?: string;
};

export type RefundGiftCardRequest = {
  amount: number;
  orderType?: string;
  orderId?: number;
  journalId: number;
  refundAccountId: number;
  liabilityAccountId: number;
  date?: string;
};

export type ForfeitExpiredGiftCardsRequest = {
  journalId: number;
  liabilityAccountId: number;
  incomeAccountId: number;
  asOf?: string;
  date?: string;
};

export type GiftCardTransaction = {
  id: number;
  giftCardId: number;
  type: "issue" | "redeem" | "refund" | "adjust" | "forfeit";
  amount: number;
  orderType: string;
  orderId: number;
  entryId: number | null;
};

export type Coupon = {
  id: number;
  organizationId: number | null;
  code: string;
  discountType: "percent" | "fixed";
  discountValue: number;
  price_bookRuleId: number | null;
  usageLimit: number | null;
  usedCount: number;
  expiryDate: Date | null;
  createdAt: Date;
  updatedAt: Date;
};

export type CreateCouponRequest = {
  organizationId?: number | null;
  code: string;
  discountType: "percent" | "fixed";
  discountValue?: number;
  price_bookRuleId?: number | null;
  usageLimit?: number | null;
  expiryDate?: string | null;
};

export type UpdateCouponRequest = {
  code: string;
  discountType: "percent" | "fixed";
  discountValue?: number;
  price_bookRuleId?: number | null;
  usageLimit?: number | null;
  expiryDate?: string | null;
};

export type RedeemCouponRequest = {
  code: string;
  subtotal?: number;
  date?: string | null;
};

export type Equipment = {
  id: number;
  organizationId: number | null;
  name: string;
  itemId: number | null;
  serialBatchId: number | null;
  ownerContactId: number | null;
  fixedAssetId: number | null;
  location: string;
  installDate: Date | null;
  warrantyEnd: Date | null;
  category: string;
};

export type CreateEquipmentRequest = {
  organizationId?: number | null;
  name: string;
  itemId?: number | null;
  serialBatchId?: number | null;
  ownerContactId?: number | null;
  fixedAssetId?: number | null;
  location?: string;
  installDate?: Date | null;
  warrantyEnd?: Date | null;
  category?: string;
};

export type ServiceContract = {
  id: number;
  organizationId: number | null;
  name: string;
  contactId: number | null;
  equipmentId: number | null;
  subscriptionId: number | null;
  coverage: string;
  slaResponseHours: number | null;
  dateStart: Date | null;
  dateEnd: Date | null;
  state: "draft" | "active" | "cancelled";
};

export type CreateServiceContractRequest = {
  organizationId?: number | null;
  name: string;
  contactId?: number | null;
  equipmentId?: number | null;
  subscriptionId?: number | null;
  coverage?: string;
  slaResponseHours?: number | null;
  dateStart?: Date | null;
  dateEnd?: Date | null;
};

export type ServiceOrder = {
  id: number;
  organizationId: number | null;
  name: string;
  contactId: number | null;
  equipmentId: number | null;
  contractId: number | null;
  type: "repair" | "maintenance" | "installation" | "inspection";
  priority: number;
  state: "new" | "scheduled" | "in_progress" | "done" | "invoiced" | "cancelled";
  scheduledDate: Date | null;
  technicianId: number | null;
  invoiceId: number | null;
  dimensionId: number | null;
  reportedIssue: string;
  resolution: string;
};

export type ServiceOrderLine = {
  id: number;
  serviceOrderId: number;
  type: "part" | "labor" | "expense";
  itemId: number | null;
  description: string;
  qty: number;
  unitId: number | null;
  unitCost: number;
  unitPrice: number;
  stockMovementId: number | null;
  billable: boolean;
  coveredByWarranty: boolean;
};

export type ServiceOrderLineRequest = {
  type: "part" | "labor" | "expense";
  itemId?: number | null;
  description?: string;
  qty: number;
  unitId?: number | null;
  unitCost?: number;
  unitPrice?: number;
  billable?: boolean;
  coveredByWarranty?: boolean;
};

export type CreateServiceOrderRequest = {
  organizationId?: number | null;
  name: string;
  contactId?: number | null;
  equipmentId?: number | null;
  contractId?: number | null;
  type: "repair" | "maintenance" | "installation" | "inspection";
  priority?: number;
  scheduledDate?: Date | null;
  technicianId?: number | null;
  dimensionId?: number | null;
  reportedIssue?: string;
  lines: ServiceOrderLineRequest[];
};

export type AddServiceOrderLineRequest = ServiceOrderLineRequest;

export type ScheduleServiceOrderRequest = {
  scheduledDate: string;
  technicianId?: number | null;
};

export type CompleteServiceOrderRequest = {
  journalId: number;
  date: string;
  cogsAccountId: number;
  stockCostAccountId: number;
  resolution?: string;
};

export type BillServiceOrderRequest = {
  journalId: number;
  date?: string;
  revenueAccountId: number;
  deferredAccount?: number | null;
};

export type MaintenancePlan = {
  id: number;
  equipmentId: number | null;
  name: string;
  intervalDays: number;
  nextDue: Date | null;
  active: boolean;
};

export type CreateMaintenancePlanRequest = {
  equipmentId: number;
  name: string;
  intervalDays: number;
  nextDue: string;
};

export type DropshipOrder = {
  id: number;
  name: string | null;
  organizationId: number | null;
  supplierId: number;
  destLocationId: number | null;
  state: string;
  receiptStatus: string;
  amountTotal: number;
};

export type DropshipLink = {
  id: number;
  saleOrderLineId: number;
  purchaseOrderLineId: number;
  stockMovementId: number | null;
};

export type CreateDropshipOrderRequest = {
  saleOrderId: number;
  supplierId: number;
  destLocationId?: number | null;
  date?: string;
};

export type ReceiveDropshipOrderRequest = {
  journalId: number;
  date?: string;
};

export type InterorganizationRule = {
  id: number;
  fromOrganizationId: number | null;
  toOrganizationId: number | null;
  autoMirror: boolean;
  vendorContactId: number | null;
  customerContactId: number | null;
};

export type UpsertInterorganizationRuleRequest = {
  fromOrganizationId?: number | null;
  toOrganizationId?: number | null;
  autoMirror?: boolean | null;
  vendorContactId?: number | null;
  customerContactId?: number | null;
};

export type InterorganizationTransaction = {
  id: number;
  sourceOrganizationId: number | null;
  sourceType: string;
  sourceId: number | null;
  mirrorOrganizationId: number | null;
  mirrorType: string;
  mirrorId: number | null;
  amount: number;
  state: "done" | "cancelled";
};

export type MirrorSaleOrderRequest = { toOrganizationId: number };

export type ConsolidationRun = {
  id: number;
  groupOrganizationId: number | null;
  periodId: number | null;
  reportingCurrency: string;
  state: "draft" | "done";
};

export type ConsolidationElimination = {
  id: number;
  accountId: number;
  counterpartyOrganizationId: number | null;
  amount: number;
  description: string;
};

export type CreateConsolidationRunRequest = {
  periodId: number;
  reportingCurrency: string;
};

export type ConsolidatedBalance = {
  organizationId: number;
  accountId: number;
  amount: number;
};

export type TrialBalance = {
  organizationId: number;
  periodId: number;
  start: Date;
  end: Date;
  rows: TrialBalanceRow[];
};

export type TrialBalanceRow = {
  accountId: number;
  code: string;
  name: string;
  accountType: string;
  openingDebit: number;
  openingCredit: number;
  periodDebit: number;
  periodCredit: number;
  closingDebit: number;
  closingCredit: number;
};

export type AgingReport = {
  asOf: Date;
  rows: AgingRow[];
};

export type AgingRow = {
  type: string;
  currencyCode: string;
  bucket: "current" | "1-30" | "31-60" | "61-90" | "90+";
  amount: number;
};

export type InventoryValuation = {
  rows: InventoryValueRow[];
};

export type InventoryValueRow = {
  itemId: number;
  productName: string;
  quantity: number;
  value: number;
};

export type ProfitAndLoss = {
  organizationId: number;
  periodId: number;
  start: Date;
  end: Date;
  rows: ProfitAndLossRow[];
  revenue: number;
  cogs: number;
  grossProfit: number;
  expenses: number;
  netIncome: number;
};

export type ProfitAndLossRow = {
  accountId: number;
  code: string;
  name: string;
  accountType: string;
  amount: number;
};

export type BalanceSheetAccount = {
  accountId: number;
  code: string;
  name: string;
  accountType: string;
  balance: number;
};

export type BalanceSheet = {
  organizationId: number;
  asOf: Date;
  assets: BalanceSheetAccount[];
  liabilities: BalanceSheetAccount[];
  equity: BalanceSheetAccount[];
  currentEarnings: number;
  totalAssets: number;
  totalLiabilities: number;
  totalEquity: number;
};

export type CashFlow = {
  organizationId: number;
  start: Date;
  end: Date;
  operating: number;
  investing: number;
  financing: number;
  netChange: number;
  openingCash: number;
  closingCash: number;
};

export type SalesKpi = {
  bookings: number;
  revenue: number;
  cogs: number;
  grossMargin: number;
  grossMarginPct: number;
  winRate: number;
};

export type PipelineKpi = {
  totalExpectedRevenue: number;
  weightedPipeline: number;
  winRate: number;
  stages: number;
};

export type InventoryKpi = {
  onHandValue: number;
  onHandQuantity: number;
  productCount: number;
};

export type SubscriptionKpi = {
  mrr: number;
  arr: number;
  churned: number;
  churnRate: number;
  ltv: number;
};

export type ProjectKpi = {
  projectCount: number;
  totalMargin: number;
  totalCost: number;
  totalBilled: number;
  utilization: number;
};

export type PayrollKpi = {
  grossCost: number;
  netCost: number;
};

export type FinanceKpi = {
  revenue: number;
  expenses: number;
  grossMarginPct: number;
  netMarginPct: number;
  ebitda: number;
  currentRatio: number;
};

export type ProcurementKpi = {
  avgCycleDays: number;
  onTimeDeliveryPct: number;
  priceVariancePct: number;
  purchaseCount: number;
};

export type ManufacturingKpi = {
  oeePct: number;
  yieldPct: number;
  scrapPct: number;
  costVariancePct: number;
  orderCount: number;
};

export type ArApKpi = {
  dso: number;
  dpo: number;
  overdueArPct: number;
  overdueApPct: number;
};

export type CashKpi = {
  position: number;
  burn: number;
  forecast: number;
};

export type InventoryRatioKpi = {
  turnover: number;
  daysOnHand: number;
  stockoutCount: number;
};

export type FxRevaluation = {
  id: number;
  organizationId: number;
  periodId: number;
  currencyCode: string;
  rate: number;
  revaluationDate: Date | null;
  state: string;
  lines?: RevaluationLine[];
};

export type RevalueRequest = {
  periodId: number;
  currencyCode: string;
  rate: number;
  revaluationDate: string;
};

export type RevaluationLine = {
  id: number;
  accountId: number;
  amount: number;
  fxGain: number;
  fxLoss: number;
  entryId: number | null;
};

export type RevaluationResult = {
  revaluation: FxRevaluation;
  lines: RevaluationLine[];
};

export type Accrual = {
  id: number;
  organizationId: number;
  type: "accrued_revenue" | "accrued_expense";
  name: string | null;
  description: string | null;
  lines?: AccrualLine[];
  state: string;
};

export type AccrualLine = {
  id: number;
  accrualId: number;
  itemId: number | null;
  qty: number;
  amount: number;
  accountId: number | null;
  recognitionDate: Date | null;
  posted: boolean;
};

export type CreateAccrualLineRequest = {
  itemId: number | null;
  qty: number;
  amount: number;
  accountId: number | null;
  recognitionDate: string;
};

export type CreateAccrualRequest = {
  type: "accrued_revenue" | "accrued_expense";
  name: string;
  description: string;
  contactId: number;
  journalId: number;
  expenseAccountId: number;
  deferredAccountId: number;
  lines: CreateAccrualLineRequest[];
};

export type AccrualResult = {
  accrual: Accrual;
  lines: AccrualLine[];
};

export type ReverseAccrualsRequest = {
  asOf: string;
};

export type ClosePeriodRequest = {
  periodId: number;
};

export type CloseResult = {
  periodId: number;
  postedMoves: number;
};

export type YearEndRollRequest = {
  taxYearId: number;
  periodId: number;
  retainedEarningsAccountId: number;
};

export type YearEndRollResult = {
  taxYearId: number;
  entryId: number;
  accounts: number;
};
