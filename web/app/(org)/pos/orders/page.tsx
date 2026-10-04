import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { PosOrdersSection } from "./_components/pos-orders-section";

export default async function OrgPosOrdersPage() {
  const t = await getTranslations("Pos");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("ordersTitle")} description={t("ordersDescription")} />
      <PosOrdersSection />
    </div>
  );
}
