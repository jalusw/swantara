import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { OrganizationAdminSection } from "./_components/organization-admin-section";
import { SystemConfigSection } from "./_components/system-config-section";

export default async function OrgAdminPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Organization admin"}
        description={
          "Configure your organization profile, fiscal settings, and system configuration."
        }
      />
      <OrganizationAdminSection orgId={id} />
      <SystemConfigSection orgId={id} />
    </div>
  );
}
