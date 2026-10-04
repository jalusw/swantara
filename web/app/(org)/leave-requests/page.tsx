import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { LeaveRequestsSection } from "./_components/leave-requests-section";

export default async function OrgLeaveRequestsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Leave");

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("subtitle")} />
      <LeaveRequestsSection orgId={id} />
    </div>
  );
}
