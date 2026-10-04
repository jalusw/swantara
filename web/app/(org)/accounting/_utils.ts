type AccountRow = {
  name: string;
  balance: number;
  movement: "up" | "down";
  percent: string;
};

type EntryRow = {
  date: string;
  description: string;
  debit: number | null;
  credit: number | null;
  tag: "income" | "expense";
};

export function getMockAccounts(): AccountRow[] {
  return [
    { name: "Current assets", balance: 58200, movement: "up", percent: "+4.2%" },
    { name: "Accounts payable", balance: 12800, movement: "down", percent: "-1.1%" },
    { name: "Accounts receivable", balance: 41300, movement: "up", percent: "+6.5%" },
    { name: "Persediaan", balance: 23900, movement: "up", percent: "+0.8%" },
    { name: "Ekuitas", balance: 96400, movement: "up", percent: "+2.9%" },
  ];
}

export function getMockEntries(): EntryRow[] {
  return [
    {
      date: "2026-08-05",
      description: "Payment received",
      debit: null,
      credit: 2400,
      tag: "income",
    },
    {
      date: "2026-08-04",
      description: "Payroll run",
      debit: 1940,
      credit: null,
      tag: "expense",
    },
    {
      date: "2026-08-03",
      description: "Supplier payment",
      debit: 880,
      credit: null,
      tag: "expense",
    },
    {
      date: "2026-08-01",
      description: "Invoice issued",
      debit: null,
      credit: 3150,
      tag: "income",
    },
    { date: "2026-07-29", description: "Rent", debit: 1200, credit: null, tag: "expense" },
    {
      date: "2026-07-25",
      description: "Software subscription",
      debit: 120,
      credit: null,
      tag: "expense",
    },
    {
      date: "2026-07-22",
      description: "Bank interest",
      debit: 60,
      credit: null,
      tag: "expense",
    },
    {
      date: "2026-07-18",
      description: "Cash deposit",
      debit: null,
      credit: 5000,
      tag: "income",
    },
    {
      date: "2026-07-15",
      description: "Tax refund",
      debit: null,
      credit: 730,
      tag: "income",
    },
    { date: "2026-07-10", description: "Fuel expense", debit: 340, credit: null, tag: "expense" },
  ];
}
