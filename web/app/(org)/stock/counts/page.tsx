import { PageHeader } from "@/components/page-header";

import { CountsSection } from "./_components/counts-section";

export default async function OrgCountsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Inventory counts"}
        description={"Count stock and review differences before posting."}
      />
      <CountsSection />
    </div>
  );
}
