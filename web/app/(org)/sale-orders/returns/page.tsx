import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { RmasSection } from "./_components/rmas-section";

export default async function OrgRmasPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Sales");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("rmasTitle")} description={t("rmasSubtitle")} />
      <RmasSection orgId={id} />
    </div>
  );
}
