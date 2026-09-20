"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Input } from "@/components/input";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { DeferralLine, DeferralSchedule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatNumber, getLocalDateString } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { canRecognize, deferralStateTone } from "../../_components/deferral-utils";

export function DeferralDetail({ orgId, deferralId }: { orgId: string; deferralId: string }) {
  const [recognizeDate, setRecognizeDate] = useState(getLocalDateString());

  const scheduleQuery = useOrgListQuery<{ schedule: DeferralSchedule }, Record<string, never>>(
    "deferral",
    (organizationId) => getSwantaraService().deferrals.get(organizationId, Number(deferralId)),
  );

  const linesQuery = useOrgListQuery<{ lines: DeferralLine[] }, Record<string, never>>(
    "deferralLines",
    (organizationId) => getSwantaraService().deferrals.lines(organizationId, Number(deferralId)),
  );

  const schedule = scheduleQuery.data?.schedule;
  const lines = linesQuery.data?.lines ?? [];

  if (scheduleQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!schedule) {
    return <p className="text-sm text-muted-foreground">{"Deferral schedule not found."}</p>;
  }

  const tone = deferralStateTone(schedule.state);

  function handleRecognize() {
    void toast.promise(
      getSwantaraService().deferrals.recognize(Number(orgId), {
        asOf: recognizeDate,
      }),
      {
        loading: "Processing…",
        success: () => {
          void scheduleQuery.refetch();
          void linesQuery.refetch();
          return "Deferrals recognized";
        },
        error: "Action failed",
      },
    );
  }

  const stateActions = (
    <div className="flex items-center gap-2">
      {canRecognize(schedule.state) ? (
        <div className="flex items-center gap-2">
          <Input
            type="date"
            value={recognizeDate}
            onChange={(e) => setRecognizeDate(e.target.value)}
            className="h-8 w-[160px]"
          />
          <Button size="sm" onClick={handleRecognize}>
            {"Recognize"}
          </Button>
        </div>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Deferrals", href: "/deferrals" },
        { label: `${schedule.sourceType} #${schedule.sourceId}` },
      ]}
      title={`${schedule.sourceType} #${schedule.sourceId}`}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "info"
                  ? "border-info text-info"
                  : tone === "danger"
                    ? "border-destructive text-destructive"
                    : ""
          }
        >
          {humanizeKey(String(schedule.state))}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Total amount"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(schedule.totalAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Recognized amount"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(schedule.recognizedAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Method"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <Badge variant="outline">{humanizeKey(String(schedule.method))}</Badge>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{"Details"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Type"}</dt>
                      <dd className="text-sm">{humanizeKey(String(schedule.type))}</dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Source"}</dt>
                      <dd className="text-sm">
                        {schedule.sourceType} #{schedule.sourceId}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Start date"}</dt>
                      <dd className="text-sm">
                        {schedule.dateStart ? formatDate(String(schedule.dateStart)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"Balance sheet account"}</dt>
                      <dd className="text-sm">
                        {schedule.balanceSheetAccountId
                          ? `#${schedule.balanceSheetAccountId}`
                          : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{"P&L account"}</dt>
                      <dd className="text-sm">
                        {schedule.plAccountId ? `#${schedule.plAccountId}` : "—"}
                      </dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
        {
          id: "lines",
          label: "Recognition lines",
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{"No recognition lines."}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"#"}</th>
                        <th className="pb-2 text-left">{"Date"}</th>
                        <th className="pb-2 text-right">{"Amount"}</th>
                        <th className="pb-2 text-right">{"Posted"}</th>
                        <th className="pb-2 text-right">{"Move"}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {lines.map((line) => (
                        <tr key={line.id} className="border-b">
                          <td className="py-2">{line.sequence}</td>
                          <td className="py-2 text-muted-foreground">
                            {line.recognitionDate ? formatDate(String(line.recognitionDate)) : "—"}
                          </td>
                          <td className="py-2 text-right tabular-nums ">
                            {formatNumber(line.amount)}
                          </td>
                          <td className="py-2 text-right">{line.posted ? "✓" : "—"}</td>
                          <td className="py-2 text-right">
                            {line.entryId ? `#${line.entryId}` : "—"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ),
        },
      ]}
    />
  );
}
