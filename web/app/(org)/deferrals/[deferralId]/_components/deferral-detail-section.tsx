"use client";

import { useTranslations } from "next-intl";
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
import { canRecognize, deferralStateTone } from "../../_components/deferral-utils";

export function DeferralDetail({ orgId, deferralId }: { orgId: string; deferralId: string }) {
  const t = useTranslations("Deferrals");
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
    return <p className="text-sm text-muted-foreground">{t("loading")}</p>;
  }

  if (!schedule) {
    return <p className="text-sm text-muted-foreground">{t("deferralNotFound")}</p>;
  }

  const tone = deferralStateTone(schedule.state);

  function handleRecognize() {
    void toast.promise(
      getSwantaraService().deferrals.recognize(Number(orgId), {
        asOf: recognizeDate,
      }),
      {
        loading: t("processing"),
        success: () => {
          void scheduleQuery.refetch();
          void linesQuery.refetch();
          return t("toastRecognized");
        },
        error: t("toastActionFailed"),
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
            {t("recognize")}
          </Button>
        </div>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("title"), href: "/deferrals" },
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
          {(t as unknown as (k: string) => string)(`deferralState_${schedule.state}`)}
        </Badge>
      }
      actions={stateActions}
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("fieldTotalAmount")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(schedule.totalAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("recognizedAmount")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold tabular-nums">
                    {formatNumber(schedule.recognizedAmount)}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("fieldMethod")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <Badge variant="outline">
                    {(t as unknown as (k: string) => string)(`deferMethod_${schedule.method}`)}
                  </Badge>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("detailCard")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("fieldType")}</dt>
                      <dd className="text-sm">
                        {(t as unknown as (k: string) => string)(`deferralType_${schedule.type}`)}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("fieldSource")}</dt>
                      <dd className="text-sm">
                        {schedule.sourceType} #{schedule.sourceId}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("fieldStartDate")}</dt>
                      <dd className="text-sm">
                        {schedule.dateStart ? formatDate(String(schedule.dateStart)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">
                        {t("fieldBalanceSheetAccount")}
                      </dt>
                      <dd className="text-sm">
                        {schedule.balanceSheetAccountId
                          ? `#${schedule.balanceSheetAccountId}`
                          : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-muted-foreground text-sm">{t("fieldPlAccount")}</dt>
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
          label: t("tabLines"),
          content: (
            <div className="space-y-4">
              {lines.length === 0 ? (
                <p className="text-sm text-muted-foreground">{t("linesEmpty")}</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b text-muted-foreground">
                        <th className="pb-2 text-left">{"#"}</th>
                        <th className="pb-2 text-left">{t("colDate")}</th>
                        <th className="pb-2 text-right">{t("colAmount")}</th>
                        <th className="pb-2 text-right">{t("colPosted")}</th>
                        <th className="pb-2 text-right">{t("colMove")}</th>
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
