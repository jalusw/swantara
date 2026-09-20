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
import type { Journal, JournalEntry } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, getLocalDateString } from "@/lib/utils";
import { humanizeKey } from "@/lib/utils/case";
import { moveStateTone } from "./journal-entry-utils";

export function JournalEntriesSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const movementsQuery = useOrgListQuery<{ movements: JournalEntry[] }, Record<string, never>>(
    "journalEntries",
    (organizationId) => getSwantaraService().journalEntries.list(organizationId),
  );
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const movements = movementsQuery.data?.movements ?? [];
  const journals = journalsQuery.data?.journals ?? [];
  const journalMap = new Map(journals.map((j) => [j.id, j.name]));

  function handleSaved() {
    setDialogOpen(false);
    void movementsQuery.refetch();
  }

  const columns: ColumnDef<JournalEntry>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => (
        <a
          href={`/accounting/journal-entries/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {row.original.name ?? `JE-${row.original.id}`}
        </a>
      ),
    },
    {
      accessorKey: "journalId",
      header: "Journal",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {journalMap.get(row.original.journalId) ?? `#${row.original.journalId}`}
        </span>
      ),
    },
    {
      accessorKey: "date",
      header: "Date",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{formatDate(row.original.date)}</span>
      ),
    },
    {
      accessorKey: "ref",
      header: "Reference",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">{row.original.ref ?? "—"}</span>
      ),
    },
    {
      accessorKey: "state",
      header: "State",
      cell: ({ row }) => (
        <Badge variant="outline" className={moveStateTone(row.original.state)}>
          {humanizeKey(String(row.original.state))}
        </Badge>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={movements}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        statusKey="state"
        statusOptions={[
          { value: "draft", label: "Draft" },
          { value: "posted", label: "Posted" },
          { value: "cancelled", label: "Cancelled" },
        ]}
        searchPlaceholder={"Search journal entries..."}
        filterLabel={"State"}
        allLabel={"All"}
        ariaLabel={"All"}
        emptyTitle={"No journal entries found."}
        status={
          movementsQuery.isLoading
            ? { type: "loading" }
            : movementsQuery.isError
              ? {
                  type: "error",
                  message: movementsQuery.error.message,
                  onRetry: () => void movementsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{"New entry"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <MoveFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          journals={journals}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function MoveFormDialog({
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
  const [journalId, setJournalId] = useState("");
  const [date, setDate] = useState(getLocalDateString());
  const [ref_, setRef] = useState("");
  const [description, setDescription] = useState("");

  const [lines, setLines] = useState<
    { id: string; accountId: string; name: string; debit: string; credit: string }[]
  >([
    { id: "1", accountId: "", name: "", debit: "", credit: "" },
    { id: "2", accountId: "", name: "", debit: "", credit: "" },
  ]);

  useEffect(() => {
    if (!open) return;
    setJournalId("");
    setDate(getLocalDateString());
    setRef("");
    setDescription("");
    setLines([
      { id: "1", accountId: "", name: "", debit: "", credit: "" },
      { id: "2", accountId: "", name: "", debit: "", credit: "" },
    ]);
  }, [open]);

  function addLine() {
    setLines((prev) => [
      ...prev,
      { id: String(Date.now()), accountId: "", name: "", debit: "", credit: "" },
    ]);
  }

  function updateLine(id: string, patch: Partial<(typeof lines)[number]>) {
    setLines((prev) => prev.map((l) => (l.id === id ? { ...l, ...patch } : l)));
  }

  function removeLine(id: string) {
    setLines((prev) => prev.filter((l) => l.id !== id));
  }

  const totalDebit = lines.reduce((s, l) => s + (Number(l.debit) || 0), 0);
  const totalCredit = lines.reduce((s, l) => s + (Number(l.credit) || 0), 0);
  const isBalanced = Math.abs(totalDebit - totalCredit) < 0.01;

  function handleSubmit() {
    if (!journalId || !isBalanced) return;
    void getSwantaraService()
      .journalEntries.create(Number(orgId) || 0, {
        organizationId: Number(orgId) || 0,
        journalId: Number(journalId) || 0,
        date,
        ref: ref_ || "",
        originType: "",
        originId: 0,
        description,
        lines: lines
          .filter((l) => l.accountId)
          .map((l) => ({
            accountId: Number(l.accountId) || 0,
            dimensionId: null,
            name: l.name,
            debit: Number(l.debit) || 0,
            credit: Number(l.credit) || 0,
          })),
      })
      .then(() => {
        toast.success("Journal entry created successfully");
        onSave();
      });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{"Create journal entry"}</DialogTitle>
          <DialogDescription>{"Description"}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-3">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Journal"}</span>
              <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
                <SelectTrigger aria-label={"Journal"}>
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
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Date"}</span>
              <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Reference"}</span>
              <Input value={ref_} onChange={(e) => setRef(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Description"}</span>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} />
          </div>
          <div className="flex flex-col gap-3 rounded-md border p-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm">{"Journal lines"}</h3>
              <Button type="button" variant="outline" size="sm" onClick={addLine}>
                {"Add line"}
              </Button>
            </div>
            {lines.map((line, _index) => (
              <div key={line.id} className="grid gap-2 rounded-md border p-3 sm:grid-cols-12">
                <div className="sm:col-span-4">
                  <span className="text-xs text-muted-foreground">{"Account"}</span>
                  <Input
                    value={line.accountId}
                    onChange={(e) => updateLine(line.id, { accountId: e.target.value })}
                    placeholder={"Account ID"}
                  />
                </div>
                <div className="sm:col-span-3">
                  <span className="text-xs text-muted-foreground">{"Name"}</span>
                  <Input
                    value={line.name}
                    onChange={(e) => updateLine(line.id, { name: e.target.value })}
                  />
                </div>
                <div className="sm:col-span-2">
                  <span className="text-xs text-muted-foreground">{"Debit"}</span>
                  <Input
                    value={line.debit}
                    onChange={(e) => updateLine(line.id, { debit: e.target.value })}
                    type="number"
                    min="0"
                    step="any"
                  />
                </div>
                <div className="sm:col-span-2">
                  <span className="text-xs text-muted-foreground">{"Credit"}</span>
                  <Input
                    value={line.credit}
                    onChange={(e) => updateLine(line.id, { credit: e.target.value })}
                    type="number"
                    min="0"
                    step="any"
                  />
                </div>
                <div className="sm:col-span-1 flex items-end">
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => removeLine(line.id)}
                  >
                    {"Remove"}
                  </Button>
                </div>
              </div>
            ))}
            <div className="flex justify-between text-sm">
              <span>
                {"Total debit"}: {totalDebit.toFixed(2)}
              </span>
              <span>
                {"Total credit"}: {totalCredit.toFixed(2)}
              </span>
              <span className={isBalanced ? "text-success" : "text-destructive"}>
                {isBalanced ? "Balanced" : "Unbalanced"}
              </span>
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!journalId || !isBalanced}>
            {"Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
