import { PageHeader } from "@/components/page-header";

import { ShipmentsSection } from "./_components/shipments-section";

export default async function OrgShipmentsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader
        title={"Shipments"}
        description={"Incoming, outgoing, and internal shipments with validation steps."}
      />
      <ShipmentsSection />
    </div>
  );
}
