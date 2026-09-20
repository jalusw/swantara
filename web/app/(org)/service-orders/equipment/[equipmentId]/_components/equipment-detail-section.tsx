"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Equipment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function EquipmentDetail({ equipmentId }: { orgId: string; equipmentId: string }) {
  const query = useOrgQuery<{ equipment: Equipment }>("equipment", equipmentId, (organizationId) =>
    getSwantaraService().equipments.get(organizationId, Number(equipmentId)),
  );

  const equipment = query.data?.equipment;

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{"Loading..."}</p>;
  }

  if (!equipment) {
    return <p className="text-sm text-muted-foreground">{"Equipment not found."}</p>;
  }

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: "Equipments", href: "/service-orders/equipment" },
        { label: equipment.name },
      ]}
      title={equipment.name}
      tabs={[
        {
          id: "overview",
          label: "Overview",
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{"Category"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{equipment.category || "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Location"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{equipment.location || "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{"Install date"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {equipment.installDate ? formatDate(String(equipment.installDate)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{"Details"}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Warranty end"}</dt>
                      <dd className="text-sm">
                        {equipment.warrantyEnd ? formatDate(String(equipment.warrantyEnd)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Owner"}</dt>
                      <dd className="text-sm">
                        {equipment.ownerContactId ? `#${equipment.ownerContactId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Item"}</dt>
                      <dd className="text-sm">{equipment.itemId ? `#${equipment.itemId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{"Fixed asset"}</dt>
                      <dd className="text-sm">
                        {equipment.fixedAssetId ? `#${equipment.fixedAssetId}` : "—"}
                      </dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
      ]}
    />
  );
}
