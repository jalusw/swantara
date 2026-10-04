import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeaveTypesSection } from "./_components/leave-types-section";

export default async function OrgLeaveTypesPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Leave");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("typesTitle")} description={t("typesSubtitle")} />
      <LeaveTypesSection orgId={id} />
    </div>
  );
}
