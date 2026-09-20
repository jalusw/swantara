import { PageHeader } from "@/components/page-header";
import { PosOrdersSection } from "./_components/pos-orders-section";

export default async function OrgPosOrdersPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"POS Orders"} description={"Browse and manage point of sale orders."} />
      <PosOrdersSection />
    </div>
  );
}
