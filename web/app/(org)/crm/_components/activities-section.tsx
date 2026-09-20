"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { ColumnDef } from "@tanstack/react-table";
import { Check, Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { ActiveBadge } from "@/components/active-badge";
import { ActivityTimeline } from "@/components/activity-timeline";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { CrmActivity } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { ActivityFormDialog } from "./activity-form-dialog";

export function ActivitiesSection({ orgId }: { orgId: string }) {
  const queryClient = useQueryClient();
  const [activityDialogOpen, setActivityDialogOpen] = useState(false);
  const [editingActivity, setEditingActivity] = useState<CrmActivity | null>(null);

  const activitiesQuery = useOrgListQuery<{ activities: CrmActivity[] }, Record<string, never>>(
    "crmActivities",
    (organizationId) => getSwantaraService().crmActivities.list(organizationId),
  );

  const activities = activitiesQuery.data?.activities ?? [];
  const isLoading = activitiesQuery.isLoading;
  const error = activitiesQuery.error;

  function handleActivitySave() {
    setActivityDialogOpen(false);
    setEditingActivity(null);
    void activitiesQuery.refetch();
  }

  const doneMutation = useMutation({
    mutationFn: (activityId: number) =>
      getSwantaraService().crmActivities.done(Number(orgId), activityId),
    onSuccess: () => {
      toast.success("Activity marked done.");
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (activityId: number) =>
      getSwantaraService().crmActivities.delete(Number(orgId), activityId),
    onSuccess: () => {
      toast.success("Activity deleted.");
      void queryClient.invalidateQueries({ queryKey: ["crmActivities"] });
    },
    onError: () => {
      toast.error("Something went wrong. Please try again.");
    },
  });

  const activityColumns: ColumnDef<CrmActivity>[] = [
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <Badge variant="secondary">{humanizeKey(String(row.original.type))}</Badge>
      ),
    },
    {
      accessorKey: "summary",
      header: "Summary",
      cell: ({ row }) => <span className="">{row.original.summary}</span>,
    },
    {
      accessorKey: "dueDate",
      header: "Due date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dueDate ? formatDate(row.original.dueDate) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "done",
      header: "Done",
      cell: ({ row }) => (
        <ActiveBadge active={row.original.done}>{row.original.done ? "Done" : "Open"}</ActiveBadge>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <div className="flex items-center justify-end gap-1">
          {!row.original.done ? (
            <Button
              size="icon-sm"
              variant="ghost"
              aria-label={"Mark done"}
              onClick={() => doneMutation.mutate(row.original.id)}
              disabled={doneMutation.isPending}
            >
              <Check className="size-4" />
            </Button>
          ) : null}
          <RowActions
            editLabel={"Edit activity"}
            deleteLabel={"Delete activity"}
            confirmTitle={"Delete this activity?"}
            confirmDescription={"The activity will be removed permanently."}
            onEdit={() => {
              setEditingActivity(row.original);
              setActivityDialogOpen(true);
            }}
            onDelete={() => {
              deleteMutation.mutate(row.original.id);
            }}
          />
        </div>
      ),
    },
  ];

  return (
    <>
      <div className="grid gap-4 lg:grid-cols-5">
        <Card className="lg:col-span-3">
          <CardHeader className="flex flex-row items-center justify-between gap-4">
            <div>
              <CardTitle className="text-base">{"Activities"}</CardTitle>
              <p className="text-sm text-muted-foreground">
                {"Calls, meetings, emails and tasks tied to a lead or contact."}
              </p>
            </div>
          </CardHeader>
          <CardContent>
            <InteractiveEntityTable
              columns={activityColumns}
              data={activities}
              getRowId={(row) => String(row.id)}
              searchKeys={["summary"]}
              statusOptions={[]}
              searchPlaceholder={"Search activities…"}
              filterLabel={"Search activities…"}
              allLabel={"All activities"}
              ariaLabel={"All activities"}
              emptyTitle={"No activities"}
              status={
                isLoading
                  ? { type: "loading" }
                  : error
                    ? {
                        type: "error",
                        message: error.message,
                        onRetry: () => void activitiesQuery.refetch(),
                      }
                    : undefined
              }
              actions={
                <Button
                  size="sm"
                  onClick={() => {
                    setEditingActivity(null);
                    setActivityDialogOpen(true);
                  }}
                >
                  <Plus />
                  <span>{"Add activity"}</span>
                </Button>
              }
            />
          </CardContent>
        </Card>
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">{"Activity timeline"}</CardTitle>
            <p className="text-sm text-muted-foreground">
              {"Calls, meetings, emails and tasks tied to a lead or contact."}
            </p>
          </CardHeader>
          <CardContent>
            {activities.length === 0 ? (
              <p className="text-sm text-muted-foreground">{"No activities"}</p>
            ) : (
              <ActivityTimeline
                items={activities
                  .slice()
                  .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
                  .slice(0, 20)
                  .map((act) => ({
                    id: String(act.id),
                    title: act.summary,
                    description: act.note ?? (act.type ? humanizeKey(String(act.type)) : undefined),
                    timestamp: act.dueDate
                      ? formatDate(act.dueDate)
                      : act.doneAt
                        ? formatDate(act.doneAt)
                        : undefined,
                    tone: act.done ? "success" : "neutral",
                  }))}
              />
            )}
          </CardContent>
        </Card>
      </div>

      {activityDialogOpen ? (
        <ActivityFormDialog
          open={activityDialogOpen}
          onOpenChange={(open) => {
            if (!open) setEditingActivity(null);
            setActivityDialogOpen(open);
          }}
          orgId={orgId}
          initial={editingActivity}
          onSave={handleActivitySave}
        />
      ) : null}
    </>
  );
}
