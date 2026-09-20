"use client";

import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { PlanningRun } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { cn } from "@/lib/utils/style";
import { PlannedSupplysTable } from "./planning-planned-orders-table";
import { PlanningRunDialog } from "./planning-run-dialog";

export function MrpSection({ orgId }: { orgId: string }) {
  const [runDialogOpen, setRunDialogOpen] = useState(false);
  const [selectedRunId, setSelectedRunId] = useState<number | null>(null);

  const runsQuery = useOrgListQuery<{ runs: PlanningRun[] }, Record<string, never>>(
    "planning.runs",
    (organizationId) => getSwantaraService().planning.runs.list(organizationId),
  );

  const selectedRunQuery = useOrgQuery<{ run: PlanningRun }>(
    "planning.runs",
    selectedRunId ?? 0,
    (organizationId) =>
      getSwantaraService().planning.runs.get(organizationId, selectedRunId as number),
  );

  const runs = runsQuery.data?.runs ?? [];
  const selectedRun = selectedRunQuery.data?.run;

  function handleRun() {
    void getSwantaraService()
      .planning.runs.create(Number(orgId), { horizonDays: 30 })
      .then((result) => {
        toast.success("Planning run created");
        setSelectedRunId(result.run.id);
        void runsQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleConfirm(runId: number) {
    void getSwantaraService()
      .planning.confirmPlannedOrder(Number(orgId), runId, { supplierId: 0 })
      .then(() => {
        toast.success("Planned orders confirmed");
        void runsQuery.refetch();
        void selectedRunQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg">{"Planning Runs"}</h2>
        <Button size="sm" onClick={handleRun}>
          <Plus />
          <span>{"Run Planning"}</span>
        </Button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {runs.map((run) => (
          <Card
            key={run.id}
            className={cn(
              "cursor-pointer transition-colors",
              selectedRunId === run.id ? "border-primary" : "hover:border-muted-foreground/50",
            )}
            onClick={() => setSelectedRunId(run.id)}
          >
            <CardHeader className="pb-2">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">{`Run-${run.id}`}</CardTitle>
                <Badge variant="outline">{run.state}</Badge>
              </div>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              <div>{formatDate(run.createdAt)}</div>
            </CardContent>
          </Card>
        ))}
        {runs.length === 0 ? (
          <div className="text-muted-foreground text-sm sm:col-span-2 lg:col-span-3">
            {"No Planning runs yet"}
          </div>
        ) : null}
      </div>

      {selectedRun ? (
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg">{`Run-${selectedRun.id}`}</h3>
            {selectedRun.state === "done" ? (
              <Button size="sm" onClick={() => handleConfirm(selectedRun.id)}>
                {"Confirm Planned Orders"}
              </Button>
            ) : null}
          </div>

          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-base">{"Demand Sources"}</CardTitle>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              {"Sales orders, manufacturing orders, and forecasts that create demand"}
            </CardContent>
          </Card>

          <PlannedSupplysTable run={selectedRun} />
        </div>
      ) : null}

      {runDialogOpen ? (
        <PlanningRunDialog
          open={runDialogOpen}
          onOpenChange={setRunDialogOpen}
          orgId={orgId}
          onSave={() => {
            setRunDialogOpen(false);
            void runsQuery.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
