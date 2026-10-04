import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DeferralsSection } from "./_components/deferrals-section";

export default async function DeferralsPage() {
  const t = await getTranslations("Deferrals");
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <DeferralsSection orgId={id} />
    </div>
  );
}
