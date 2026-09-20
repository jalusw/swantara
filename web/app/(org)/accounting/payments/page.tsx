import { PageHeader } from "@/components/page-header";
import { PaymentsSection } from "./_components/payments-section";

export default async function OrgPaymentsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Payments"} description={"View and manage recorded payments."} />
      <PaymentsSection />
    </div>
  );
}
