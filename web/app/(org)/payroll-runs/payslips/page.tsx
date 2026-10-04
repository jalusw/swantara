import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { PayslipsSection } from "./_components/payslips-section";

export default async function OrgPayslipsPage() {
  const t = await getTranslations("Payroll");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("payslipsTitle")} description={t("payslipsSubtitle")} />
      <PayslipsSection />
    </div>
  );
}
