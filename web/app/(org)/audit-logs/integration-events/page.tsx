import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { IntegrationEventsSection } from "./_components/integration-events-section";

export default async function OrgIntegrationEventsPage() {
  const t = await getTranslations("AuditLogs");
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("eventsTitle")} description={t("eventsDescription")} />
      <IntegrationEventsSection orgId={id} />
    </div>
  );
}
