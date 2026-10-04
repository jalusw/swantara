import { getTranslations } from "next-intl/server";
import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { DepartmentsSection } from "./_components/departments-section";

export default async function OrgDepartmentsPage() {
  const id = String(await requireActiveOrgId());
  const t = await getTranslations("Employees");
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={t("departmentsTitle")} description={t("departmentsSubtitle")} />
      <DepartmentsSection orgId={id} />
    </div>
  );
}
