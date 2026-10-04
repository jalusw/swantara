"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { MaintenancePlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function MaintenancePlanDetail({ planId }: { orgId: string; planId: string }) {
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
  const query = useOrgQuery<{ maintenancePlan: MaintenancePlan }>(
    "maintenancePlan",
    planId,
    (organizationId) => getSwantaraService().maintenancePlans.get(organizationId, Number(planId)),
  );

  const plan = query.data?.maintenancePlan;

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!plan) {
    return <p className="text-sm text-muted-foreground">{t("maintenancePlanNotFound")}</p>;
  }

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("plansTitle"), href: "/service-orders/plans" },
        { label: plan.name },
      ]}
      title={plan.name}
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("equipment")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.equipmentId ? `#${plan.equipmentId}` : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("intervalDays")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm tabular-nums">
                    {t("daysCount", { count: plan.intervalDays })}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("nextDue")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.nextDue ? formatDate(String(plan.nextDue)) : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("active")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.active ? tCommon("yes") : tCommon("no")}</p>
                </CardContent>
              </Card>
            </div>
          ),
        },
      ]}
    />
  );
}
