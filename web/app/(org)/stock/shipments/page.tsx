import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";

import { ShipmentsSection } from "./_components/shipments-section";

export default async function OrgShipmentsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Stock",
  );
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("shipmentsTitle")} description={t("shipmentsDescription")} />
      <ShipmentsSection />
    </div>
  );
}
