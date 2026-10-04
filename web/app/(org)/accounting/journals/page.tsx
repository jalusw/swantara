import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { JournalsSection } from "./_components/journals-section";

export default async function OrgJournalsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("journalsTitle")} description={t("journalsDescription")} />
      <JournalsSection orgId={id} />
    </div>
  );
}
