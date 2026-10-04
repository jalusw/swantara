import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { MrpSection } from "./_components/planning-section";

export default async function OrgMrpPage() {
  const t = await getTranslations("Planning");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <MrpSection orgId={id} />
    </div>
  );
}
