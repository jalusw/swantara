"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { TaxYear } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import { isYearLocked, yearDurationMonths } from "./tax-year-utils";

export function TaxYearsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const yearsQuery = useOrgListQuery<{ taxYears: TaxYear[] }, Record<string, never>>(
    "taxYears",
    (organizationId) => getSwantaraService().taxYears.list(organizationId),
  );

  const years = yearsQuery.data?.taxYears ?? [];

  function handleSaved() {
    setDialogOpen(false);
    void yearsQuery.refetch();
  }

  const columns: ColumnDef<TaxYear>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "dateStart",
      header: "Start date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateStart ? formatDate(row.original.dateStart) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: "End date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateEnd ? formatDate(row.original.dateEnd) : "—"}
        </span>
      ),
    },
    {
      id: "duration",
      header: "Duration",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {yearDurationMonths(row.original)} {"months"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge
          variant="outline"
          className={
            row.original.state === "done"
              ? "success"
              : row.original.state === "locked"
                ? "warning"
                : "neutral"
          }
        >
          {row.original.state ?? "open"}
        </Badge>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => {
        const locked = isYearLocked(row.original);
        return (
          <div className="flex gap-1">
            {locked ? (
              <Badge variant="secondary" className="text-xs">
                {"Locked"}
              </Badge>
            ) : null}
          </div>
        );
      },
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={years}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search tax years..."}
        emptyTitle={"No tax years defined."}
        ariaLabel={"Accounts"}
        status={
          yearsQuery.isLoading
            ? { type: "loading" }
            : yearsQuery.isError
              ? {
                  type: "error",
                  message: yearsQuery.error.message,
                  onRetry: () => void yearsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"Add tax year"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <TaxYearFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function TaxYearFormDialog({
  open,
  onOpenChange,
  orgId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  onSave: () => void;
}) {
  const [name, setName] = useState("");
  const [dateStart, setDateStart] = useState("");
  const [dateEnd, setDateEnd] = useState("");

  useEffect(() => {
    if (!open) return;
    setName("");
    setDateStart("");
    setDateEnd("");
  }, [open]);

  function handleSubmit() {
    if (!name || !dateStart || !dateEnd) return;
    void getSwantaraService()
      .taxYears.create(Number(orgId), {
        organizationId: Number(orgId),
        name,
        dateStart,
        dateEnd,
      })
      .then(() => {
        toast.success("Tax year created successfully");
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Create tax year"}</DialogTitle>
          <DialogDescription>
            {"Define the name and date range for this tax year."}
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Name"}</span>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Start date"}</span>
              <Input type="date" value={dateStart} onChange={(e) => setDateStart(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"End date"}</span>
              <Input type="date" value={dateEnd} onChange={(e) => setDateEnd(e.target.value)} />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!name || !dateStart || !dateEnd}>
            {"Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
