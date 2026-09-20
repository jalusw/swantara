"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { RowActions } from "@/app/(org)/_components/row-actions";
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
import type { Account, Journal } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { journalTypeLabels, journalTypeTone } from "./journal-utils";

export function JournalsSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingJournal, setEditingJournal] = useState<Journal | null>(null);

  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );
  const accountsQuery = useOrgListQuery<{ accounts: Account[] }, Record<string, never>>(
    "accounts",
    (organizationId) => getSwantaraService().accounts.list(organizationId),
  );

  const journals = journalsQuery.data?.journals ?? [];
  const accounts = accountsQuery.data?.accounts ?? [];
  const accountMap = new Map(accounts.map((a) => [a.id, `${a.code} — ${a.name}`]));

  function handleEdit(journal: Journal) {
    setEditingJournal(journal);
    setDialogOpen(true);
  }

  function handleCreate() {
    setEditingJournal(null);
    setDialogOpen(true);
  }

  function handleDelete(journal: Journal) {
    void getSwantaraService()
      .journals.delete(Number(orgId), journal.id)
      .then(() => void journalsQuery.refetch());
  }

  function handleSaved() {
    setDialogOpen(false);
    setEditingJournal(null);
    void journalsQuery.refetch();
  }

  const columns: ColumnDef<Journal>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "code",
      header: "Code",
      cell: ({ row }) => (
        <span className="font-mono text-muted-foreground">{row.original.code ?? "—"}</span>
      ),
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: ({ row }) => (
        <Badge variant="outline" className={journalTypeTone(row.original.type)}>
          {journalTypeLabels[row.original.type]}
        </Badge>
      ),
    },
    {
      accessorKey: "defaultAccountId",
      header: "Default account",
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.defaultAccountId
            ? (accountMap.get(row.original.defaultAccountId) ?? "—")
            : "—"}
        </span>
      ),
    },
    {
      id: "actions",
      header: "",
      cell: ({ row }) => (
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete journal"}
          confirmDescription={
            "Are you sure you want to delete this journal? This action cannot be undone."
          }
          onEdit={() => handleEdit(row.original)}
          onDelete={() => handleDelete(row.original)}
        />
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={journals}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={"Search journals..."}
        emptyTitle={"No journals yet."}
        ariaLabel={"Accounts"}
        status={
          journalsQuery.isLoading
            ? { type: "loading" }
            : journalsQuery.isError
              ? {
                  type: "error",
                  message: journalsQuery.error.message,
                  onRetry: () => void journalsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={handleCreate}>
            <Plus />
            <span>{"Add journal"}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <JournalFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          journal={editingJournal}
          accounts={accounts}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function JournalFormDialog({
  open,
  onOpenChange,
  orgId,
  journal,
  accounts,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  journal: Journal | null;
  accounts: Account[];
  onSave: () => void;
}) {
  const [name, setName] = useState(journal?.name ?? "");
  const [code, setCode] = useState(journal?.code ?? "");
  const [type, setType] = useState<Journal["type"]>(journal?.type ?? "general");
  const [defaultAccountId, setDefaultAccountId] = useState(
    journal?.defaultAccountId != null ? String(journal.defaultAccountId) : "",
  );

  const journalTypes = ["sale", "purchase", "bank", "cash", "general"] as const;

  function handleSubmit() {
    if (!name) return;
    const request = {
      name,
      code: code || null,
      type,
      defaultAccountId: defaultAccountId ? Number(defaultAccountId) : null,
      bankAccountId: null as number | null,
    };
    const promise = journal
      ? getSwantaraService().journals.update(Number(orgId), journal.id, request)
      : getSwantaraService().journals.create(Number(orgId), request);
    void promise.then(() => {
      toast.success(journal ? "Journal updated successfully" : "Journal created successfully");
      onSave();
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{journal ? "Edit journal" : "Create journal"}</DialogTitle>
          <DialogDescription>
            {"Configure the journal name, type, and default account."}
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Name"}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{"Code"}</span>
              <Input value={code} onChange={(e) => setCode(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Type"}</span>
            <Select value={type} onValueChange={(v) => setType(v as Journal["type"])}>
              <SelectTrigger aria-label={"Type"}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {journalTypes.map((tp) => (
                  <SelectItem key={tp} value={tp}>
                    {journalTypeLabels[tp]}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{"Default account"}</span>
            <Select value={defaultAccountId} onValueChange={(v) => setDefaultAccountId(v ?? "")}>
              <SelectTrigger aria-label={"Default account"}>
                <SelectValue placeholder={"None"} />
              </SelectTrigger>
              <SelectContent>
                {accounts.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>
                    {a.code} — {a.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {"Cancel"}
          </Button>
          <Button onClick={handleSubmit} disabled={!name}>
            {"Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
