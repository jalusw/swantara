import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosConfigsSection } from "./_components/pos-configs-section";

export default async function OrgPosConfigsPage() {
  const t = await getTranslations("Pos");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("configsTitle")} description={t("configsDescription")} />
      <PosConfigsSection orgId={id} />
    </div>
  );
}
