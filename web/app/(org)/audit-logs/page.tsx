import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AuditLogsSection } from "./_components/audit-logs-section";

export default async function OrgAuditLogsPage() {
  const t = await getTranslations("AuditLogs");
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <AuditLogsSection orgId={id} />
    </div>
  );
}
