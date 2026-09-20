import { PageHeader } from "@/components/page-header";
import { PayslipsSection } from "./_components/payslips-section";

export default async function OrgPayslipsPage() {
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Payslips"} description={"View employee payslips and breakdowns."} />
      <PayslipsSection />
    </div>
  );
}
