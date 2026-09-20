import { PageHeader } from "@/components/page-header";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ProjectsSection } from "./_components/projects-section";

export default async function OrgProjectsPage() {
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <PageHeader title={"Projects"} description={"Projects, tasks, milestones and job costing."} />
      <ProjectsSection orgId={id} />
    </div>
  );
}
