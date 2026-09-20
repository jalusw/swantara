const endpoints = {
  health: "/api/v1/health",

  auth: {
    login: "/api/v1/auth/login",
    register: "/api/v1/auth/register",
    checkEmail: "/api/v1/auth/email/check",
    refresh: "/api/v1/auth/refresh",
    logout: "/api/v1/auth/logout",
    requestEmailVerification: "/api/v1/auth/email-verification/request",
    verifyEmail: "/api/v1/auth/email-verification/verify",
    requestPasswordReset: "/api/v1/auth/password-reset/request",
    resetPassword: "/api/v1/auth/password-reset",
    sessions: "/api/v1/auth/sessions",
    revokeSession: (sessionId: string) => `/api/v1/auth/sessions/${sessionId}`,
  },

  me: {
    me: "/api/v1/me",
    organizations: "/api/v1/me/organizations",
    permissions: (organizationId: string) =>
      `/api/v1/me/organizations/${organizationId}/permissions`,
    avatar: "/api/v1/me/avatar",
  },

  users: {
    list: "/api/v1/users",
    get: (id: string) => `/api/v1/users/${id}`,
    create: "/api/v1/users",
    update: (id: string) => `/api/v1/users/${id}`,
    delete: (id: string) => `/api/v1/users/${id}`,
  },

  organizations: {
    list: "/api/v1/organizations",
    create: "/api/v1/organizations",
    quickCreate: "/api/v1/organizations/quick",
    get: (id: string) => `/api/v1/organizations/${id}`,
    update: (id: string) => `/api/v1/organizations/${id}`,
    delete: (id: string) => `/api/v1/organizations/${id}`,
    modules: (id: string) => `/api/v1/organizations/${id}/modules`,
    updateModule: (id: string) => `/api/v1/organizations/${id}/modules`,
  },

  accountingStandards: {
    list: "/api/v1/accounting-standards",
  },

  members: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/members`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/members`,
    delete: (organizationId: string, memberId: string) =>
      `/api/v1/organizations/${organizationId}/members/${memberId}`,
    roles: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/member-roles`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/member-roles`,
    },
    permissions: (organizationId: string) => `/api/v1/organizations/${organizationId}/permissions`,
  },

  fxRates: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/fx-rates`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/fx-rates`,
    resolve: (organizationId: string) => `/api/v1/organizations/${organizationId}/fx-rates/resolve`,
    get: (organizationId: string, fxRateId: string) =>
      `/api/v1/organizations/${organizationId}/fx-rates/${fxRateId}`,
    update: (organizationId: string, fxRateId: string) =>
      `/api/v1/organizations/${organizationId}/fx-rates/${fxRateId}`,
    delete: (organizationId: string, fxRateId: string) =>
      `/api/v1/organizations/${organizationId}/fx-rates/${fxRateId}`,
  },

  currencies: {
    list: "/api/v1/currencies",
    get: (code: string) => `/api/v1/currencies/${code}`,
  },

  unitGroups: {
    list: "/api/v1/unit-categories",
    create: "/api/v1/unit-categories",
    get: (id: string) => `/api/v1/unit-categories/${id}`,
    update: (id: string) => `/api/v1/unit-categories/${id}`,
    delete: (id: string) => `/api/v1/unit-categories/${id}`,
  },

  units: {
    list: "/api/v1/units",
    create: "/api/v1/units",
    convert: "/api/v1/units/convert",
    get: (id: string) => `/api/v1/units/${id}`,
    update: (id: string) => `/api/v1/units/${id}`,
    delete: (id: string) => `/api/v1/units/${id}`,
  },

  dimensions: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/dimensions`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/dimensions`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/dimensions/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/dimensions/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/dimensions/${id}`,
  },

  paymentTerms: {
    list: "/api/v1/payment-terms",
    create: "/api/v1/payment-terms",
    get: (id: string) => `/api/v1/payment-terms/${id}`,
    update: (id: string) => `/api/v1/payment-terms/${id}`,
    delete: (id: string) => `/api/v1/payment-terms/${id}`,
    splits: (id: string) => `/api/v1/payment-terms/${id}/splits`,
  },

  accounts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/accounts`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/accounts`,
    get: (organizationId: string, accountId: string) =>
      `/api/v1/organizations/${organizationId}/accounts/${accountId}`,
    update: (organizationId: string, accountId: string) =>
      `/api/v1/organizations/${organizationId}/accounts/${accountId}`,
    delete: (organizationId: string, accountId: string) =>
      `/api/v1/organizations/${organizationId}/accounts/${accountId}`,
  },

  journals: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/journals`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/journals`,
    get: (organizationId: string, journalId: string) =>
      `/api/v1/organizations/${organizationId}/journals/${journalId}`,
    update: (organizationId: string, journalId: string) =>
      `/api/v1/organizations/${organizationId}/journals/${journalId}`,
    delete: (organizationId: string, journalId: string) =>
      `/api/v1/organizations/${organizationId}/journals/${journalId}`,
  },

  taxes: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/taxes`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/taxes`,
    get: (organizationId: string, taxId: string) =>
      `/api/v1/organizations/${organizationId}/taxes/${taxId}`,
    update: (organizationId: string, taxId: string) =>
      `/api/v1/organizations/${organizationId}/taxes/${taxId}`,
    delete: (organizationId: string, taxId: string) =>
      `/api/v1/organizations/${organizationId}/taxes/${taxId}`,
  },

  taxYears: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-years`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-years`,
    get: (organizationId: string, taxYearId: string) =>
      `/api/v1/organizations/${organizationId}/tax-years/${taxYearId}`,
    update: (organizationId: string, taxYearId: string) =>
      `/api/v1/organizations/${organizationId}/tax-years/${taxYearId}`,
    delete: (organizationId: string, taxYearId: string) =>
      `/api/v1/organizations/${organizationId}/tax-years/${taxYearId}`,
  },

  carriers: {
    list: "/api/v1/carriers",
    create: "/api/v1/carriers",
    get: (id: string) => `/api/v1/carriers/${id}`,
    update: (id: string) => `/api/v1/carriers/${id}`,
    delete: (id: string) => `/api/v1/carriers/${id}`,
  },

  systemConfigs: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/system-configs`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/system-configs`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/system-configs/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/system-configs/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/system-configs/${id}`,
  },

  integrationEvents: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/integration-events`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/integration-events/${id}`,
    dispatch: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/integration-events/${id}/dispatch`,
  },

  auditLogs: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/audit-logs`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/audit-logs/${id}`,
  },

  contacts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/contacts`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/contacts`,
    get: (organizationId: string, contactId: string) =>
      `/api/v1/organizations/${organizationId}/contacts/${contactId}`,
    update: (organizationId: string, contactId: string) =>
      `/api/v1/organizations/${organizationId}/contacts/${contactId}`,
    delete: (organizationId: string, contactId: string) =>
      `/api/v1/organizations/${organizationId}/contacts/${contactId}`,
    addresses: {
      list: (organizationId: string, contactId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/addresses`,
      create: (organizationId: string, contactId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/addresses`,
      update: (organizationId: string, contactId: string, addressId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/addresses/${addressId}`,
      delete: (organizationId: string, contactId: string, addressId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/addresses/${addressId}`,
      setDefault: (organizationId: string, contactId: string, addressId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/addresses/${addressId}/default`,
    },
    bankAccounts: {
      list: (organizationId: string, contactId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/bank-accounts`,
      create: (organizationId: string, contactId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/bank-accounts`,
      update: (organizationId: string, contactId: string, accountId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/bank-accounts/${accountId}`,
      delete: (organizationId: string, contactId: string, accountId: string) =>
        `/api/v1/organizations/${organizationId}/contacts/${contactId}/bank-accounts/${accountId}`,
    },
    customer: (organizationId: string, contactId: string) =>
      `/api/v1/organizations/${organizationId}/contacts/${contactId}/customer`,
    supplier: (organizationId: string, contactId: string) =>
      `/api/v1/organizations/${organizationId}/contacts/${contactId}/supplier`,
  },

  itemCategories: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/item-categories`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/item-categories`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/item-categories/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/item-categories/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/item-categories/${id}`,
  },

  products: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/products`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/products`,
    get: (organizationId: string, itemId: string) =>
      `/api/v1/organizations/${organizationId}/products/${itemId}`,
    update: (organizationId: string, itemId: string) =>
      `/api/v1/organizations/${organizationId}/products/${itemId}`,
    delete: (organizationId: string, itemId: string) =>
      `/api/v1/organizations/${organizationId}/products/${itemId}`,
    variants: {
      list: (organizationId: string, itemId: string) =>
        `/api/v1/organizations/${organizationId}/products/${itemId}/variants`,
      create: (organizationId: string, itemId: string) =>
        `/api/v1/organizations/${organizationId}/products/${itemId}/variants`,
    },
  },

  priceBooks: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/price_books`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/price_books`,
    get: (organizationId: string, priceBookId: string) =>
      `/api/v1/organizations/${organizationId}/price_books/${priceBookId}`,
    update: (organizationId: string, priceBookId: string) =>
      `/api/v1/organizations/${organizationId}/price_books/${priceBookId}`,
    delete: (organizationId: string, priceBookId: string) =>
      `/api/v1/organizations/${organizationId}/price_books/${priceBookId}`,
    rules: {
      list: (organizationId: string, priceBookId: string) =>
        `/api/v1/organizations/${organizationId}/price_books/${priceBookId}/rules`,
      create: (organizationId: string, priceBookId: string) =>
        `/api/v1/organizations/${organizationId}/price_books/${priceBookId}/rules`,
    },
    resolve: (organizationId: string, priceBookId: string) =>
      `/api/v1/organizations/${organizationId}/price_books/${priceBookId}/resolve`,
  },

  supplierProducts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/supplier-products`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/supplier-products`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-products/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-products/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-products/${id}`,
    bestOffer: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/supplier-products/best-offer`,
  },

  recipes: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/recipes`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/recipes`,
    get: (organizationId: string, recipeId: string) =>
      `/api/v1/organizations/${organizationId}/recipes/${recipeId}`,
    update: (organizationId: string, recipeId: string) =>
      `/api/v1/organizations/${organizationId}/recipes/${recipeId}`,
    delete: (organizationId: string, recipeId: string) =>
      `/api/v1/organizations/${organizationId}/recipes/${recipeId}`,
    lines: (organizationId: string, recipeId: string) =>
      `/api/v1/organizations/${organizationId}/recipes/${recipeId}/lines`,
    explode: (organizationId: string, recipeId: string) =>
      `/api/v1/organizations/${organizationId}/recipes/${recipeId}/explode`,
  },

  productionOrders: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/production-orders`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/production-orders`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}`,
    components: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/components`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/confirm`,
    plan: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/plan`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/cancel`,
    start: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/start`,
    shopTasks: {
      list: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/production-orders/${id}/shop-tasks`,
      create: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/production-orders/${id}/shop-tasks`,
      labor: (organizationId: string, productionOrderId: string, shopTaskId: string) =>
        `/api/v1/organizations/${organizationId}/production-orders/${productionOrderId}/shop-tasks/${shopTaskId}/labor`,
    },
    consume: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/consume`,
    produce: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/produce`,
    settle: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/production-orders/${id}/settle`,
  },

  planning: {
    runs: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/planning/runs`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/planning/runs`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/planning/runs/${id}`,
    },
    plannedOrders: {
      confirm: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/planning/planned-orders/${id}/confirm`,
    },
    forecasts: {
      list: (organizationId: string) =>
        `/api/v1/organizations/${organizationId}/planning/forecasts`,
      create: (organizationId: string) =>
        `/api/v1/organizations/${organizationId}/planning/forecasts`,
    },
  },

  outsideProcessingOrders: {
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/outside-processing-orders`,
    send: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/outside-processing-orders/${id}/send`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/outside-processing-orders/${id}/receive`,
    done: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/outside-processing-orders/${id}/done`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/outside-processing-orders/${id}/cancel`,
  },

  journalEntries: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/journal-entries`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/journal-entries`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/journal-entries/${id}`,
    reverse: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/journal-entries/${id}/reverse`,
  },

  invoices: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/invoices`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/invoices`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/invoices/${id}`,
    creditNote: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/invoices/${id}/credit-note`,
  },

  supplierBills: {
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/supplier-bills`,
    creditNote: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-bills/${id}/credit-note`,
  },

  payments: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/payments`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/payments`,
    createOutbound: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/payments/outbound`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payments/${id}`,
  },

  taxPeriods: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-periods`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-periods`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-periods/${id}`,
    close: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-periods/${id}/close`,
    lock: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-periods/${id}/lock`,
    open: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-periods/${id}/open`,
  },

  bankStatements: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/bank-statements`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/bank-statements`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/bank-statements/${id}`,
    match: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/bank-statements/${id}/match`,
  },

  reconciliations: {
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/reconciliations`,
  },

  reminder: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/reminder/actions`,
    generate: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reminder/actions/generate`,
  },

  budgets: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/budgets`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/budgets`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/budgets/${id}`,
    variance: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/budgets/${id}/variance`,
  },

  taxRules: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-rules`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-rules`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-rules/${id}`,
    resolve: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-rules/${id}/resolve`,
  },

  withholdingTaxes: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/withholding-taxes`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/withholding-taxes`,
    apply: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/withholding-taxes/apply`,
  },

  taxReturns: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-returns`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/tax-returns`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-returns/${id}`,
    file: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-returns/${id}/file`,
    pay: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-returns/${id}/pay`,
    open: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/tax-returns/${id}/open`,
  },

  deferrals: {
    recognize: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/deferrals/recognize`,
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/deferrals`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/deferrals`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/deferrals/${id}`,
    lines: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/deferrals/${id}/lines`,
  },

  warehouses: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/warehouses`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/warehouses`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouses/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouses/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouses/${id}`,
  },

  stockLocations: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-locations`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-locations`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-locations/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-locations/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-locations/${id}`,
  },

  stock: {
    onHand: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock/on-hand`,
    availableToPromise: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/stock/available-to-promise`,
    balances: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock/balances`,
    rebuild: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock/rebuild`,
  },

  stockMovements: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-movements`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-movements`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-movements/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-movements/${id}`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-movements/${id}/receive`,
    ship: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-movements/${id}/ship`,
  },

  shipments: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/shipments`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/shipments/${id}`,
  },

  stockHolds: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-holds`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-holds`,
    releaseByMovement: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/stock-holds/release-by-move`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-holds/${id}`,
  },

  batches: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/batches`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/batches`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/batches/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/batches/${id}`,
  },

  reorderRules: {
    candidates: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reorder-rules/candidates`,
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/reorder-rules`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/reorder-rules`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/reorder-rules/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/reorder-rules/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/reorder-rules/${id}`,
  },

  stockCounts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-counts`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/stock-counts`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-counts/${id}`,
    lines: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-counts/${id}/lines`,
    post: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-counts/${id}/post`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/stock-counts/${id}`,
  },

  warehouseTransfers: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/warehouse-transfers`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/warehouse-transfers`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouse-transfers/${id}`,
    send: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouse-transfers/${id}/send`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/warehouse-transfers/${id}/receive`,
  },

  inboundCosts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/inbound-costs`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/inbound-costs`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/inbound-costs/${id}`,
    lines: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/inbound-costs/${id}/lines`,
    adjustments: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/inbound-costs/${id}/adjustments`,
    post: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/inbound-costs/${id}/post`,
  },

  crm: {
    leads: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/leads`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/leads`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/leads/${id}`,
      update: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/leads/${id}`,
      delete: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/leads/${id}`,
      promote: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/leads/${id}/promote`,
    },
    opportunities: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/opportunities`,
      create: (organizationId: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}`,
      update: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}`,
      delete: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}`,
      advanceStage: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}/advance-stage`,
      win: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}/win`,
      lose: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/opportunities/${id}/lose`,
    },
    activities: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/activities`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/activities`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/activities/${id}`,
      update: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/activities/${id}`,
      delete: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/activities/${id}`,
      done: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/crm/activities/${id}/done`,
    },
    pipeline: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/pipeline`,
    stages: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/stages`,
    teams: (organizationId: string) => `/api/v1/organizations/${organizationId}/crm/teams`,
  },

  saleOrders: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/sale-orders`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/sale-orders`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}`,
    send: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/send`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/confirm`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/cancel`,
    done: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/done`,
    recomputeStatuses: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/recompute-statuses`,
    deliver: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/deliver`,
    invoice: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/invoice`,
    pay: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/sale-orders/${id}/pay`,
  },

  pos: {
    configs: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/configs`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/configs`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/configs/${id}`,
    },
    sessions: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/sessions`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/sessions`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/sessions/${id}`,
      closing: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/sessions/${id}/closing`,
      close: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/sessions/${id}/close`,
    },
    orders: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/orders`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/pos/orders`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/orders/${id}`,
      invoice: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/orders/${id}/invoice`,
      refund: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/pos/orders/${id}/refund`,
    },
  },

  approvalRequests: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/approval-requests`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/approval-requests`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/approval-requests/${id}`,
    decide: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/approval-requests/${id}/decide`,
  },

  attachments: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/attachments`,
    upload: (organizationId: string) => `/api/v1/organizations/${organizationId}/attachments`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/attachments/${id}`,
    download: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/attachments/${id}/download`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/attachments/${id}`,
  },

  messages: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/messages`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/messages`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/messages/${id}`,
  },

  quality: {
    points: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/quality-points`,
      create: (organizationId: string) => `/api/v1/organizations/${organizationId}/quality-points`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/quality-points/${id}`,
    },
    checks: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/quality-checks`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/quality-checks/${id}`,
      result: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/quality-checks/${id}/result`,
      scrap: (organizationId: string, shipmentId: string) =>
        `/api/v1/organizations/${organizationId}/quality-checks/shipments/${shipmentId}/scrap`,
    },
    alerts: {
      list: (organizationId: string) => `/api/v1/organizations/${organizationId}/quality-alerts`,
      get: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/quality-alerts/${id}`,
      state: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/quality-alerts/${id}/state`,
    },
  },

  purchaseRequests: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/purchase-requests`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/purchase-requests`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-requests/${id}`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-requests/${id}/confirm`,
    approve: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-requests/${id}/approve`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-requests/${id}/cancel`,
  },

  purchaseOrders: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/purchase-orders`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/purchase-orders`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}/confirm`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}/cancel`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}/receive`,
    vendorBill: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}/supplier-bill`,
    pay: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/purchase-orders/${id}/pay`,
  },

  supplierQuoteRequests: {
    list: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests`,
    fromRequisition: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/from-request`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}`,
    send: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/send`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/cancel`,
    lines: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/lines`,
    quotes: {
      list: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/quotes`,
      create: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/quotes`,
      accept: (organizationId: string, id: string, quoteId: string) =>
        `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/quotes/${quoteId}/accept`,
    },
    purchaseOrder: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/supplier-quote-requests/${id}/purchase-order`,
  },

  rmas: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/rmas`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/rmas`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}/confirm`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}/receive`,
    refund: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}/refund`,
    done: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}/done`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/rmas/${id}/cancel`,
  },

  departments: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/departments`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/departments`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/departments/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/departments/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/departments/${id}`,
  },

  jobPositions: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/job-positions`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/job-positions`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/job-positions/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/job-positions/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/job-positions/${id}`,
  },

  leaveTypes: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/leave-types`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/leave-types`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-types/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-types/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-types/${id}`,
  },

  employees: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/employees`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/employees`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/employees/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/employees/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/employees/${id}`,
  },

  contracts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/contracts`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/contracts`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/contracts/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/contracts/${id}`,
    terminate: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/contracts/${id}/terminate`,
  },

  leaveRequests: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/leave-requests`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/leave-requests`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-requests/${id}`,
    submit: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-requests/${id}/submit`,
    approve: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-requests/${id}/approve`,
    refuse: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/leave-requests/${id}/refuse`,
    balance: (organizationId: string, employeeId: string, leaveTypeId: string) =>
      `/api/v1/organizations/${organizationId}/leave-requests/balance/${employeeId}/${leaveTypeId}`,
  },

  attendances: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/attendances`,
    checkIn: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/attendances/check-in`,
    checkOut: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/attendances/${id}/check-out`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/attendances/${id}`,
  },

  timesheets: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/timesheets`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/timesheets`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/timesheets/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/timesheets/${id}`,
  },

  salaryRules: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/salary-rules`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/salary-rules`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/salary-rules/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/salary-rules/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/salary-rules/${id}`,
  },

  payrollRuns: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/payroll-runs`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/payroll-runs`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payroll-runs/${id}`,
    confirm: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payroll-runs/${id}/confirm`,
    pay: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payroll-runs/${id}/pay`,
    close: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payroll-runs/${id}/close`,
  },

  payslips: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/payslips`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/payslips/${id}`,
  },

  projects: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/projects`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/projects`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/projects/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/projects/${id}`,
    state: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/projects/${id}/state`,
    summary: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/projects/${id}/summary`,
    bill: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/projects/${id}/bill`,
    tasks: {
      list: (organizationId: string, projectId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/tasks`,
      create: (organizationId: string, projectId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/tasks`,
      update: (organizationId: string, projectId: string, taskId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/tasks/${taskId}`,
    },
    milestones: {
      list: (organizationId: string, projectId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/milestones`,
      create: (organizationId: string, projectId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/milestones`,
      reached: (organizationId: string, projectId: string, milestoneId: string) =>
        `/api/v1/organizations/${organizationId}/projects/${projectId}/milestones/${milestoneId}/reached`,
    },
  },

  expenseCategories: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/expense-categories`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/expense-categories`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-categories/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-categories/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-categories/${id}`,
  },

  expenseReports: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/expense-reports`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/expense-reports`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}`,
    submit: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/submit`,
    approve: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/approve`,
    refuse: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/refuse`,
    post: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/post`,
    reimburse: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/reimburse`,
    bill: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/expense-reports/${id}/bill`,
  },

  assetCategories: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/asset-categories`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/asset-categories`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/asset-categories/${id}`,
  },

  fixedAssets: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/fixed-assets`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/fixed-assets`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/fixed-assets/${id}`,
    schedule: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/fixed-assets/${id}/schedule`,
    postDepreciation: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/fixed-assets/${id}/post-depreciation`,
    dispose: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/fixed-assets/${id}/dispose`,
  },

  subscriptionPlans: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/subscription-plans`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/subscription-plans`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscription-plans/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscription-plans/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscription-plans/${id}`,
  },

  subscriptions: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/subscriptions`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/subscriptions`,
    metrics: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/metrics`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}`,
    activate: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}/activate`,
    pause: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}/pause`,
    resume: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}/resume`,
    churn: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}/churn`,
    close: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/subscriptions/${id}/close`,
  },

  commissionPlans: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/commission-plans`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/commission-plans`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/commission-plans/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/commission-plans/${id}`,
    rules: {
      list: (organizationId: string, planId: string) =>
        `/api/v1/organizations/${organizationId}/commission-plans/${planId}/rules`,
      create: (organizationId: string, planId: string) =>
        `/api/v1/organizations/${organizationId}/commission-plans/${planId}/rules`,
    },
    assignments: {
      list: (organizationId: string, planId: string) =>
        `/api/v1/organizations/${organizationId}/commission-plans/${planId}/assignments`,
      create: (organizationId: string, planId: string) =>
        `/api/v1/organizations/${organizationId}/commission-plans/${planId}/assignments`,
    },
  },

  commissionEntries: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/commission-entries`,
    accrue: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/commission-entries/accrue`,
    accrueFromInvoice: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/commission-entries/accrue-from-invoice`,
    pay: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/commission-entries/${id}/pay`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/commission-entries/${id}/cancel`,
  },

  giftCards: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/gift-cards`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/gift-cards`,
    forfeitExpired: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/gift-cards/forfeit-expired`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/gift-cards/${id}`,
    redeem: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/gift-cards/${id}/redeem`,
    refund: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/gift-cards/${id}/refund`,
    transactions: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/gift-cards/${id}/transactions`,
  },

  coupons: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/coupons`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/coupons`,
    redeem: (organizationId: string) => `/api/v1/organizations/${organizationId}/coupons/redeem`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/coupons/${id}`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/coupons/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/coupons/${id}`,
  },

  equipments: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/equipments`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/equipments`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/equipments/${id}`,
  },

  serviceContracts: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/service-contracts`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/service-contracts`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-contracts/${id}`,
    activate: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-contracts/${id}/activate`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-contracts/${id}/cancel`,
  },

  serviceOrders: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/service-orders`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/service-orders`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}`,
    lines: {
      list: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/service-orders/${id}/lines`,
      create: (organizationId: string, id: string) =>
        `/api/v1/organizations/${organizationId}/service-orders/${id}/lines`,
    },
    schedule: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}/schedule`,
    start: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}/start`,
    complete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}/complete`,
    bill: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}/bill`,
    cancel: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/service-orders/${id}/cancel`,
  },

  maintenancePlans: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/maintenance-plans`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/maintenance-plans`,
    generateOrders: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/maintenance-plans/generate-orders`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/maintenance-plans/${id}`,
  },

  dropshipOrders: {
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/dropship-orders`,
    receive: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/dropship-orders/${id}/receive`,
  },

  interorganizationRules: {
    list: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-rules`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-rules`,
    update: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-rules/${id}`,
    delete: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-rules/${id}`,
  },

  interorganizationTransactions: {
    list: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-transactions`,
    mirrorSaleOrder: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/interorganization-transactions/mirror-sale-order/${id}`,
  },

  consolidationRuns: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/consolidation-runs`,
    create: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/consolidation-runs`,
    get: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/consolidation-runs/${id}`,
    run: (organizationId: string, id: string) =>
      `/api/v1/organizations/${organizationId}/consolidation-runs/${id}/run`,
  },

  reports: {
    trialBalance: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reports/trial-balance`,
    aging: (organizationId: string) => `/api/v1/organizations/${organizationId}/reports/aging`,
    inventoryValuation: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reports/inventory-valuation`,
    profitAndLoss: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reports/profit-and-loss`,
    balanceSheet: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reports/balance-sheet`,
    cashFlow: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/reports/cash-flow`,
  },

  kpis: {
    sales: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/sales`,
    pipeline: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/pipeline`,
    inventory: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/inventory`,
    subscription: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/kpis/subscription`,
    projects: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/projects`,
    payroll: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/payroll`,
    finance: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/finance`,
    procurement: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/kpis/procurement`,
    manufacturing: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/kpis/manufacturing`,
    arAp: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/ar-ap`,
    cash: (organizationId: string) => `/api/v1/organizations/${organizationId}/kpis/cash`,
    inventoryRatio: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/kpis/inventory-ratio`,
  },

  fxRevaluations: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/fx-revaluations`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/fx-revaluations`,
  },

  accruals: {
    list: (organizationId: string) => `/api/v1/organizations/${organizationId}/accruals`,
    create: (organizationId: string) => `/api/v1/organizations/${organizationId}/accruals`,
    reverseDue: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/accruals/reverse-due`,
  },

  periodClose: {
    close: (organizationId: string) => `/api/v1/organizations/${organizationId}/period-close`,
    yearEndRoll: (organizationId: string) =>
      `/api/v1/organizations/${organizationId}/period-close/year-end-roll`,
  },
};

export { endpoints };
