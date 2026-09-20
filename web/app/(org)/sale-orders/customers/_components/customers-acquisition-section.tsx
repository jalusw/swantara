"use client";

import { BarChart } from "@/components/bar-chart";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Contact } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

const MONTH_NAMES = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];

function computeAcquisitionTrend(contacts: Contact[]) {
  const now = new Date();
  const months: Array<{ label: string; value: number }> = [];

  for (let i = 6; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    const year = d.getFullYear();
    const month = d.getMonth();
    const count = contacts.filter((p) => {
      if (!p.createdAt) return false;
      const created = new Date(p.createdAt);
      return created.getFullYear() === year && created.getMonth() === month;
    }).length;
    months.push({ label: MONTH_NAMES[month]!, value: count });
  }

  return months;
}

export function CustomersAcquisition() {
  const query = useOrgListQuery<{ contacts: Contact[] }, Record<string, never>>(
    "contacts",
    (organizationId) => getSwantaraService().contacts.list(organizationId),
  );

  const contacts = query.data?.contacts ?? [];
  const acquisitionTrend = computeAcquisitionTrend(contacts);
  const total = acquisitionTrend.reduce((sum, month) => sum + month.value, 0);

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Customer acquisition"}</CardTitle>
        <CardDescription>{"New customers added over the last six months."}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {query.isLoading ? (
          <output
            className="flex h-44 items-end gap-2"
            aria-busy="true"
            aria-label={"Customer acquisition"}
          >
            {acquisitionTrend.map((month) => (
              <div
                key={month.label}
                className="h-full flex-1 animate-pulse rounded-md bg-muted"
                aria-hidden
              />
            ))}
          </output>
        ) : query.isError ? (
          <div className="flex flex-col items-start gap-2 py-4">
            <p className="text-sm text-muted-foreground">{"Something went wrong."}</p>
            <Button variant="outline" size="sm" onClick={() => void query.refetch()}>
              {"Retry"}
            </Button>
          </div>
        ) : (
          <>
            <BarChart
              data={acquisitionTrend}
              ariaLabel={"Customer acquisition"}
              valueFormatter={(value) => String(value)}
            />
            {total === 0 ? (
              <p className="text-sm text-muted-foreground">
                {"Add your first customer to get started."}
              </p>
            ) : null}
          </>
        )}
      </CardContent>
    </Card>
  );
}
