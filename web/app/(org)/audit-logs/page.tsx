import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { AuditLogsSection } from "./_components/audit-logs-section";

export default async function OrgAuditLogsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Audit logs"}
        description={"Track all changes made to records across your organization."}
      />
      <AuditLogsSection orgId={id} />
    </div>
  );
}
