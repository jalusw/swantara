import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";

import { TransfersSection } from "./_components/transfers-section";

export default async function OrgTransfersPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Stock",
  );
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("transfersTitle")} description={t("transfersDescription")} />
      <TransfersSection />
    </div>
  );
}
