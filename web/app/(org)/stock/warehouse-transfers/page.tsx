import { PageHeader } from "@/components/page-header";

import { TransfersSection } from "./_components/transfers-section";

export default async function OrgTransfersPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Transfers"} description={"Inter-warehouse transfers through transit."} />
      <TransfersSection />
    </div>
  );
}
