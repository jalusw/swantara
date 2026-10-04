"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
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
import { journalTypeTone } from "./journal-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function JournalsSection({ orgId }: { orgId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");
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
      header: () => t("fieldName"),
      cell: ({ row }) => <span className="">{row.original.name}</span>,
    },
    {
      accessorKey: "code",
      header: () => t("fieldCode"),
      cell: ({ row }) => (
        <span className="font-mono text-muted-foreground">{row.original.code ?? "—"}</span>
      ),
    },
    {
      accessorKey: "type",
      header: () => t("fieldType"),
      cell: ({ row }) => (
        <Badge variant="outline" className={journalTypeTone(row.original.type)}>
          {(t as unknown as (k: string) => string)(`journalType_${row.original.type}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "defaultAccountId",
      header: () => t("fieldDefaultAccount"),
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
          editLabel={tCommon("edit")}
          deleteLabel={tCommon("delete")}
          confirmTitle={t("deleteJournal")}
          confirmDescription={t("deleteJournalDescription")}
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
        searchPlaceholder={t("searchJournals")}
        emptyTitle={t("journalsEmpty")}
        ariaLabel={t("journalsTitle")}
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
            <span>{t("addJournal")}</span>
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
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");

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
      toast.success(journal ? t("toastJournalUpdated") : t("toastJournalCreated"));
      onSave();
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{journal ? t("editJournal") : t("createJournal")}</DialogTitle>
          <DialogDescription>{t("journalDialogDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldName")}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldCode")}</span>
              <Input value={code} onChange={(e) => setCode(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldType")}</span>
            <Select value={type} onValueChange={(v) => setType(v as Journal["type"])}>
              <SelectTrigger aria-label={t("fieldType")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {journalTypes.map((tp) => (
                  <SelectItem key={tp} value={tp}>
                    {(t as unknown as (k: string) => string)(`journalType_${tp}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldDefaultAccount")}</span>
            <Select value={defaultAccountId} onValueChange={(v) => setDefaultAccountId(v ?? "")}>
              <SelectTrigger aria-label={t("fieldDefaultAccount")}>
                <SelectValue placeholder={t("noneLabel")} />
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
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!name}>
            {tCommon("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
