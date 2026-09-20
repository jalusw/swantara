import { BackLink } from "@/components/back-link";
import { requireActiveOrgId } from "@/lib/server/active-org";
import { ProjectDetail } from "./_components/project-detail-section";

export default async function OrgProjectDetailPage({
  params,
}: {
  params: Promise<{ projectId: string }>;
}) {
  const { projectId } = await params;
  const id = String(await requireActiveOrgId());
  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <BackLink href={"/projects"}>{"Back to projects"}</BackLink>
      <ProjectDetail orgId={id} projectId={projectId} />
    </div>
  );
}
