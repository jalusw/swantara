"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/table";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { formatMoney } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import type { getMockEntries } from "../_utils";

function Tag({ tone }: { tone: "income" | "expense" }) {
  const color =
    tone === "income" ? "bg-success/10 text-success" : "bg-destructive/10 text-destructive";
  return (
    <span className={cn("rounded-full px-2 py-0.5 text-[10px] uppercase tracking-wide", color)}>
      {tone === "income" ? "IN" : "EX"}
    </span>
  );
}

export function RecentEntries({ entries }: { entries: ReturnType<typeof getMockEntries> }) {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Accounting");
  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("recentEntriesTitle")}</CardTitle>
        <CardDescription>{t("recentEntriesDescription")}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("colDate")}</TableHead>
              <TableHead>{t("colDescription")}</TableHead>
              <TableHead className="text-right">{t("colDebit")}</TableHead>
              <TableHead className="text-right">{t("colCredit")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.map((entry) => (
              <TableRow key={`${entry.date}-${entry.description}`}>
                <TableCell className="text-muted-foreground">{entry.date}</TableCell>
                <TableCell>
                  <div className="flex items-center gap-2">
                    <span className="">{entry.description}</span>
                    <Tag tone={entry.tag === "income" ? "income" : "expense"} />
                  </div>
                </TableCell>
                <TableCell className="text-right text-destructive tabular-nums">
                  {formatMoney(entry.debit, { currency: DEFAULT_CURRENCY, nullFallback: "—" })}
                </TableCell>
                <TableCell className="text-right text-success tabular-nums">
                  {formatMoney(entry.credit, { currency: DEFAULT_CURRENCY, nullFallback: "—" })}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
