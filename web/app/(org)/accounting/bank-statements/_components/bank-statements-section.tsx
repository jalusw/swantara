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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { BankStatement, Journal } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney, getLocalDateString } from "@/lib/utils";
import { bankStatementStateLabel, bankStatementStateTone } from "./statement-utils";

export function BankStatementsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const statementsQuery = useOrgListQuery<
    { bankStatements: BankStatement[] },
    Record<string, never>
  >("bankStatements", (organizationId) => getSwantaraService().bankStatements.list(organizationId));
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const statements = statementsQuery.data?.bankStatements ?? [];
  const journalMap = new Map((journalsQuery.data?.journals ?? []).map((j) => [j.id, j.name]));

  function handleSaved() {
    setDialogOpen(false);
    void statementsQuery.refetch();
  }

  const columns: ColumnDef<BankStatement>[] = [
    {
      accessorKey: "name",
      header: "Reference",
      cell: ({ row }) => (
        <a
          href={`/accounting/bank-statements/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `BS-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "journalId",
      header: "Bank journal",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {journalMap.get(row.original.journalId ?? 0) ?? `#${row.original.journalId}`}
        </span>
      ),
    },
    {
      accessorKey: "date",
      header: "Date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.date ? formatDate(row.original.date) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "balanceStart",
      header: "Opening balance",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatMoney(row.original.balanceStart)}
        </span>
      ),
    },
    {
      accessorKey: "balanceEnd",
      header: "Closing balance",
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums ">{formatMoney(row.original.balanceEnd)}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge variant="outline" className={bankStatementStateTone(row.original.state)}>
          {bankStatementStateLabel(row.original.state)}
        </Badge>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={statements}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "open", label: "Open" },
          { value: "reconciled", label: "Reconciled" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search statements..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"Accounts"}
        emptyTitle={"No bank statements found."}
        status={
          statementsQuery.isLoading
            ? { type: "loading" }
            : statementsQuery.isError
              ? {
                  type: "error",
                  message: statementsQuery.error.message,
                  onRetry: () => void statementsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New statement"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <StatementFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          journals={journalsQuery.data?.journals ?? []}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function StatementFormDialog({
  open,
  onOpenChange,
  orgId,
  journals,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  journals: Journal[];
  onSave: () => void;
}) {
  const [name, setName] = useState("");
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());
  const [balanceStart, setBalanceStart] = useState("0");

  useEffect(() => {
    if (!open) return;
    setName("");
    setJournalId("");
    setDate(getLocalDateString());
    setBalanceStart("0");
  }, [open]);

  function handleSubmit() {
    if (!journalId) return;
    void getSwantaraService()
      .bankStatements.create(Number(orgId) || 0, {
        organizationId: Number(orgId) || 0,
        journalId: Number(journalId) || 0,
        date,
        balanceStart: Number(balanceStart) || 0,
        balanceEnd: Number(balanceStart) || 0,
        lines: [],
      })
      .then(() => {
        toast.success("Bank statement created successfully");
        onSave();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{"Create bank statement"}</DialogTitle>
          <DialogDescription>{"Create a new bank statement for reconciliation."}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Bank journal"}</span>
            <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
              <SelectTrigger aria-label={"Bank journal"}>
                <SelectValue placeholder={"Select a journal"} />
              </SelectTrigger>
              <SelectContent>
                {journals.map((j) => (
                  <SelectItem key={j.id} value={String(j.id)}>
                    {j.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Reference"}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Date"}</span>
              <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Opening balance"}</span>
            <Input
              value={balanceStart}
              onChange={(e) => setBalanceStart(e.target.value)}
              type="number"
              step="any"
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!journalId}>
            {"Create"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
