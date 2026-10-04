"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Currency } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export type CurrencyRow = {
  code: string;
  name: string;
  symbol: string;
  decimalPlaces: number;
  rounding: string;
};

function toCurrencyRow(currency: Currency): CurrencyRow {
  return {
    code: currency.code,
    name: currency.name,
    symbol: currency.symbol ?? "",
    decimalPlaces: currency.decimalPlaces,
    rounding: String(currency.rounding),
  };
}

export function CurrenciesSection(_props: { orgId: string }) {
  const t = useTranslations("Reference");
  const query = useOrgListQuery<{ currencies: Currency[] }, Record<string, never>>(
    "currencies",
    () => getSwantaraService().currencies.list(),
  );

  const currencies = (query.data?.currencies ?? []).map(toCurrencyRow);

  const columns: ColumnDef<CurrencyRow>[] = [
    {
      accessorKey: "code",
      header: t("tableCode"),
      cell: ({ row }) => <span className="font-mono text-xs">{row.original.code}</span>,
    },
    {
      accessorKey: "name",
      header: t("tableName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "symbol",
      header: t("tableSymbol"),
      cell: ({ row }) => <span className="text-muted-foreground">{row.original.symbol}</span>,
    },
    {
      accessorKey: "decimalPlaces",
      header: t("tableDecimals"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{row.original.decimalPlaces}</span>
      ),
    },
    {
      accessorKey: "rounding",
      header: t("tableRounding"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">{row.original.rounding}</span>
      ),
    },
  ];

  return (
    <InteractiveEntityTable
      columns={columns}
      data={currencies}
      getRowId={(row) => row.code}
      searchKeys={["name", "code"]}
      searchPlaceholder={t("searchCurrenciesPlaceholder")}
      filterLabel=""
      statusOptions={[]}
      allLabel=""
      ariaLabel={t("currenciesTitle")}
      emptyTitle={t("currenciesEmpty")}
      status={
        query.isLoading
          ? { type: "loading" }
          : query.isError
            ? {
                type: "error",
                message: query.error.message,
                onRetry: () => void query.refetch(),
              }
            : undefined
      }
    />
  );
}
