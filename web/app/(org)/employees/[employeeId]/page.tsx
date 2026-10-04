import { getTranslations } from "next-intl/server";
import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { EmployeeDetail } from "./_components/employee-detail-section";

export default async function EmployeeDetailPage({
  params,
}: {
  params: Promise<{ employeeId: string }>;
}) {
  const { employeeId } = await params;
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Employees");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/employees"}>{t("backToEmployees")}</BackLink>
      <EmployeeDetail orgId={id} employeeId={employeeId} />
    </div>
  );
}
