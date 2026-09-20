import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { FinancialCard } from "./_components/financial-card";
import { ModulesCard } from "./_components/modules-card";
import { PreferencesCard } from "./_components/preferences-card";
import { ProfileCard } from "./_components/profile-card";

export default async function OrgGeneralSettingsPage() {
  const id = String(await requireActiveOrgId());

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"General settings"}
        description={"Manage your organization's profile, financials, and preferences."}
      />

      <ProfileCard orgId={id} />

      <FinancialCard orgId={id} />

      <ModulesCard />

      <PreferencesCard orgId={id} />
    </div>
  );
}
