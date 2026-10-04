import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { OrganizationAdminSection } from "./_components/organization-admin-section";
import { SystemConfigSection } from "./_components/system-config-section";

export default async function OrgAdminPage() {
  const t = await getTranslations("Admin");
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("title")} description={t("description")} />
      <OrganizationAdminSection orgId={id} />
      <SystemConfigSection orgId={id} />
    </div>
  );
}
