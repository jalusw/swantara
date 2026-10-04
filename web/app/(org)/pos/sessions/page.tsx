import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { PosSessionsSection } from "./_components/pos-sessions-section";

export default async function OrgPosSessionsPage() {
  const t = await getTranslations("Pos");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("sessionsTitle")} description={t("sessionsDescription")} />
      <PosSessionsSection orgId={id} />
    </div>
  );
}
