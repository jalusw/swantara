import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { PaymentsSection } from "./_components/payments-section";

export default async function OrgPaymentsPage() {
  const t = await (getTranslations as unknown as (ns: string) => Promise<(key: string) => string>)(
    "Accounting",
  );
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("paymentsTitle")} description={t("paymentsDescription")} />
      <PaymentsSection />
    </div>
  );
}
