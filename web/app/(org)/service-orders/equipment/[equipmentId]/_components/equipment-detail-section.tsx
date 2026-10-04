"use client";

import { useTranslations } from "next-intl";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Equipment } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

export function EquipmentDetail({ equipmentId }: { orgId: string; equipmentId: string }) {
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
  const query = useOrgQuery<{ equipment: Equipment }>("equipment", equipmentId, (organizationId) =>
    getSwantaraService().equipments.get(organizationId, Number(equipmentId)),
  );

  const equipment = query.data?.equipment;

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!equipment) {
    return <p className="text-sm text-muted-foreground">{t("equipmentNotFound")}</p>;
  }

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("equipmentTitle"), href: "/service-orders/equipment" },
        { label: equipment.name },
      ]}
      title={equipment.name}
      tabs={[
        {
          id: "overview",
          label: t("tabOverview"),
          content: (
            <div className="grid gap-4 lg:grid-cols-3">
              <Card>
                <CardHeader>
                  <CardTitle>{t("category")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{equipment.category || "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("location")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{equipment.location || "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("installDate")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">
                    {equipment.installDate ? formatDate(String(equipment.installDate)) : "—"}
                  </p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("detailsTitle")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("warrantyEnd")}</dt>
                      <dd className="text-sm">
                        {equipment.warrantyEnd ? formatDate(String(equipment.warrantyEnd)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("owner")}</dt>
                      <dd className="text-sm">
                        {equipment.ownerContactId ? `#${equipment.ownerContactId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("item")}</dt>
                      <dd className="text-sm">{equipment.itemId ? `#${equipment.itemId}` : "—"}</dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("fixedAsset")}</dt>
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
