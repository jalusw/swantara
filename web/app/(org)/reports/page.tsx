import { PageHeader } from "@/components/page-header";
import { ReportsHubSection } from "./_components/reports-hub-section";

export default async function OrgReportsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Reports"}
        description={"Generate and download reports for your organization."}
      />
      <ReportsHubSection />
    </div>
  );
}
