export type ContactRole = "customer" | "supplier";

export type StubContactAddress = {
  id: string;
  type: "billing" | "shipping" | "physical" | "other";
  line1: string;
  line2: string;
  city: string;
  state: string;
  postalCode: string;
  countryCode: string;
  isDefault: boolean;
};

export type StubContactBankAccount = {
  id: string;
  accountHolder: string;
  bankName: string;
  iban: string;
  swiftBic: string;
  accountNumber: string;
  routingNumber: string;
  currencyCode: string;
};

export type StubContact = {
  id: string;
  name: string;
  displayName: string;
  isOrganization: boolean;
  email: string;
  phone: string;
  mobile: string;
  website: string;
  taxId: string;
  industry: string;
  currencyCode: string;
  lang: string;
  active: boolean;
  roles: Partial<Record<ContactRole, boolean>>;
  customer?: {
    paymentTermId: string;
    creditLimit: string;
    receivableAccountId: string;
  };
  supplier?: {
    paymentTermId: string;
    payableAccountId: string;
  };
  addresses: StubContactAddress[];
  bankAccounts: StubContactBankAccount[];
};

export const paymentTermOptions = [
  { id: "pt1", name: "Net 30" },
  { id: "pt2", name: "50/50 split" },
  { id: "pt3", name: "Due on receipt" },
];

export const accountOptions = [
  { id: "acc-1100", name: "1100 · Accounts receivable" },
  { id: "acc-1200", name: "1200 · Trade receivables" },
  { id: "acc-2100", name: "2100 · Accounts payable" },
];

export const currencyOptions = ["USD", "IDR", "EUR", "SGD", "GBP"];

export const countryCodeOptions = [
  { code: "US", name: "United States" },
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapore" },
  { code: "GB", name: "United Kingdom" },
  { code: "DE", name: "Germany" },
];

export const initialContacts: StubContact[] = [
  {
    id: "p1",
    name: "Bluebird Trading Pte. Ltd.",
    displayName: "Bluebird Trading",
    isOrganization: true,
    email: "billing@bluebird.sg",
    phone: "+65 6123 4567",
    mobile: "+65 8123 4567",
    website: "https://bluebird.sg",
    taxId: "202012345K",
    industry: "retail",
    currencyCode: "SGD",
    lang: "en",
    active: true,
    roles: { customer: true, supplier: true },
    customer: {
      paymentTermId: "pt1",
      creditLimit: "50000",
      receivableAccountId: "acc-1100",
    },
    supplier: {
      paymentTermId: "pt3",
      payableAccountId: "acc-2100",
    },
    addresses: [
      {
        id: "pa1-1",
        type: "billing",
        line1: "10 Anson Road",
        line2: "#26-08 International Plaza",
        city: "Singapore",
        state: "",
        postalCode: "079903",
        countryCode: "SG",
        isDefault: true,
      },
      {
        id: "pa1-2",
        type: "shipping",
        line1: "51 Changi South Ave 2",
        line2: "Warehouse Block B",
        city: "Singapore",
        state: "",
        postalCode: "486103",
        countryCode: "SG",
        isDefault: false,
      },
    ],
    bankAccounts: [
      {
        id: "pb1-1",
        accountHolder: "Bluebird Trading Pte. Ltd.",
        bankName: "DBS Bank",
        iban: "",
        swiftBic: "DBSSSGSG",
        accountNumber: "012-345678-9",
        routingNumber: "",
        currencyCode: "SGD",
      },
    ],
  },
  {
    id: "p2",
    name: "PT Nusantara Logistics",
    displayName: "Nusantara Logistics",
    isOrganization: true,
    email: "finance@nusantaralog.id",
    phone: "+62 21 5550 1188",
    mobile: "+62 811 5550 1188",
    website: "https://nusantaralog.id",
    taxId: "02.345.678.9-002.000",
    industry: "logistics",
    currencyCode: "IDR",
    lang: "id",
    active: true,
    roles: { supplier: true },
    supplier: {
      paymentTermId: "pt1",
      payableAccountId: "acc-2100",
    },
    addresses: [
      {
        id: "pa2-1",
        type: "physical",
        line1: "Jl. Raya Cakung Cilincing No. 88",
        line2: "Kawasan Berikat Nusantara",
        city: "Jakarta Utara",
        state: "DKI Jakarta",
        postalCode: "14140",
        countryCode: "ID",
        isDefault: true,
      },
    ],
    bankAccounts: [
      {
        id: "pb2-1",
        accountHolder: "PT Nusantara Logistics",
        bankName: "Bank Mandiri",
        iban: "",
        swiftBic: "BMRIIDJA",
        accountNumber: "123-00-4567890-1",
        routingNumber: "",
        currencyCode: "IDR",
      },
    ],
  },
  {
    id: "p3",
    name: "Klima Foods GmbH",
    displayName: "Klima Foods",
    isOrganization: true,
    email: "team@klimafoods.de",
    phone: "+49 30 1204 550",
    mobile: "",
    website: "https://klimafoods.de",
    taxId: "DE312345678",
    industry: "food",
    currencyCode: "EUR",
    lang: "de",
    active: false,
    roles: { customer: true },
    customer: {
      paymentTermId: "pt2",
      creditLimit: "15000",
      receivableAccountId: "acc-1200",
    },
    addresses: [
      {
        id: "pa3-1",
        type: "billing",
        line1: "Warschauer Str. 58",
        line2: "",
        city: "Berlin",
        state: "",
        postalCode: "10243",
        countryCode: "DE",
        isDefault: true,
      },
    ],
    bankAccounts: [],
  },
  {
    id: "p4",
    name: "Aria Chen",
    displayName: "Aria Chen",
    isOrganization: false,
    email: "aria.chen@example.com",
    phone: "+1 415 555 0134",
    mobile: "+1 415 555 0134",
    website: "",
    taxId: "",
    industry: "",
    currencyCode: "USD",
    lang: "en",
    active: true,
    roles: { customer: true },
    customer: {
      paymentTermId: "pt3",
      creditLimit: "2000",
      receivableAccountId: "acc-1100",
    },
    addresses: [
      {
        id: "pa4-1",
        type: "shipping",
        line1: "2200 Harrison St",
        line2: "Apt 4",
        city: "San Francisco",
        state: "CA",
        postalCode: "94110",
        countryCode: "US",
        isDefault: true,
      },
    ],
    bankAccounts: [],
  },
];

export function findContact(id: string): StubContact | undefined {
  return structuredClone(initialContacts.find((contact) => contact.id === id));
}
