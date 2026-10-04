"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function TaxYearsSection({ orgId }: { orgId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
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
      header: () => t("fieldName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "dateStart",
      header: () => t("fieldStartDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateStart ? formatDate(row.original.dateStart) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: () => t("fieldEndDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.dateEnd ? formatDate(row.original.dateEnd) : "—"}
        </span>
      ),
    },
    {
      id: "duration",
      header: () => t("colDuration"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {t("durationMonths", { count: yearDurationMonths(row.original) })}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: () => t("colStatus"),
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
          {(t as unknown as (k: string) => string)(`taxYearState_${row.original.state ?? "open"}`)}
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
                {t("locked")}
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
        searchPlaceholder={t("searchTaxYears")}
        emptyTitle={t("taxYearsEmpty")}
        ariaLabel={t("taxYearsTitle")}
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
            <span>{t("addTaxYear")}</span>
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
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");

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
        toast.success(t("toastTaxYearCreated"));
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("createTaxYear")}</DialogTitle>
          <DialogDescription>{t("taxYearDialogDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldName")}</span>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldStartDate")}</span>
              <Input type="date" value={dateStart} onChange={(e) => setDateStart(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldEndDate")}</span>
              <Input type="date" value={dateEnd} onChange={(e) => setDateEnd(e.target.value)} />
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!name || !dateStart || !dateEnd}>
            {tCommon("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
