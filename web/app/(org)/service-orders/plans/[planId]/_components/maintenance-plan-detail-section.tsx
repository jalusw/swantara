"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { MaintenancePlan } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function MaintenancePlanDetail({ planId }: { orgId: string; planId: string }) {
  const query = useOrgQuery<{ maintenancePlan: MaintenancePlan }>(
    "maintenancePlan",
    planId,
    (organizationId) => getSwantaraService().maintenancePlans.get(organizationId, Number(planId)),
  );

  const plan = query.data?.maintenancePlan;

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!plan) {
    return <p className="text-sm text-muted-foreground">{"Maintenance plan not found."}</p>;
  }

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Maintenance plans", href: "/service-orders/plans" },
        { label: plan.name },
      ]}
      title={plan.name}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Equipment"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.equipmentId ? `#${plan.equipmentId}` : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Interval (days)"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm tabular-nums">{plan.intervalDays} days</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Next due"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.nextDue ? formatDate(String(plan.nextDue)) : "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Active"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{plan.active ? "Yes" : "No"}</p>
                </CardContent>
              </Card>
            </div>
          ),
        },
      ]}
    />
  );
}
