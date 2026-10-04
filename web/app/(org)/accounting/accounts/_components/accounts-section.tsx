"use client";

import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
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
import type { Account } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { cn } from "@/lib/utils/style";
import { type AccountTreeNode, accountTypeTone, buildAccountTree } from "./account-utils";

export function AccountsSection({ orgId }: { orgId: string }) {
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Accounting");
  const tCommon = useTranslations("Common");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingAccount, setEditingAccount] = useState<Account | null>(null);

  const accountsQuery = useOrgListQuery<{ accounts: Account[] }, Record<string, never>>(
    "accounts",
    (organizationId) => getSwantaraService().accounts.list(organizationId),
  );

  const accounts = accountsQuery.data?.accounts ?? [];
  const tree = buildAccountTree(accounts);

  function handleEdit(account: Account) {
    setEditingAccount(account);
    setDialogOpen(true);
  }

  function handleCreate() {
    setEditingAccount(null);
    setDialogOpen(true);
  }

  function handleSaved() {
    setDialogOpen(false);
    setEditingAccount(null);
    void accountsQuery.refetch();
  }

  function renderNode(node: AccountTreeNode, depth: number) {
    return (
      <div key={node.id}>
        <div
          className="flex items-center gap-2 rounded border px-3 py-2 hover:bg-muted/50"
          style={{ marginLeft: depth * 24 }}
        >
          <span className="font-mono text-sm w-20 shrink-0">{node.code}</span>
          <span className="flex-1 text-sm">{node.name}</span>
          <Badge variant="outline" className={cn("text-xs", accountTypeTone(node.type))}>
            {(t as unknown as (k: string) => string)(`accountType_${node.type}`)}
          </Badge>
          {node.reconcilable ? (
            <Badge variant="secondary" className="text-xs">
              {t("badgeReconcilable")}
            </Badge>
          ) : null}
          {!node.active ? (
            <Badge variant="outline" className="text-xs text-muted-foreground">
              {t("badgeInactive")}
            </Badge>
          ) : null}
          <Button size="sm" variant="ghost" onClick={() => handleEdit(node)}>
            {tCommon("edit")}
          </Button>
        </div>
        {node.children.map((child) => renderNode(child, depth + 1))}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">
          {t("accountsCount", { count: accounts.length })}
        </p>
        <Button size="sm" onClick={handleCreate}>
          {t("addAccount")}
        </Button>
      </div>
      {accountsQuery.isLoading ? (
        <div className="text-sm text-muted-foreground">{t("loadingAccounts")}</div>
      ) : tree.length === 0 ? (
        <div className="text-sm text-muted-foreground">{t("accountsEmpty")}</div>
      ) : (
        <div className="flex flex-col gap-1 rounded-md border p-2">
          {tree.map((node) => renderNode(node, 0))}
        </div>
      )}
      {dialogOpen ? (
        <AccountFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          account={editingAccount}
          accounts={accounts}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function AccountFormDialog({
  open,
  onOpenChange,
  orgId,
  account,
  accounts,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  account: Account | null;
  accounts: Account[];
  onSave: () => void;
}) {
  const [code, setCode] = useState(account?.code ?? "");
  const [name, setName] = useState(account?.name ?? "");
  const [type, setType] = useState<Account["type"]>(account?.type ?? "asset");
  const [reconcilable, setReconcilable] = useState(account?.reconcilable ?? false);
  const [parentId, setParentId] = useState(
    account?.parentId != null ? String(account.parentId) : "",
  );
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Accounting");
  const tCommon = useTranslations("Common");

  const accountTypes = [
    "asset",
    "liability",
    "equity",
    "income",
    "expense",
    "receivable",
    "payable",
    "bank",
    "cash",
    "cogs",
    "tax",
    "current_asset",
    "fixed_asset",
    "depreciation",
  ] as const;

  function handleSubmit() {
    if (!code || !name) return;
    const request = {
      code,
      name,
      type,
      reconcilable,
      parentId: parentId ? Number(parentId) : null,
      currencyCode: null as string | null,
      active: true,
    };
    const promise = account
      ? getSwantaraService().accounts.update(Number(orgId), account.id, request)
      : getSwantaraService().accounts.create(Number(orgId), request);
    void promise.then(() => {
      toast.success(account ? t("toastAccountUpdated") : t("toastAccountCreated"));
      onSave();
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{account ? t("editAccount") : t("createAccount")}</DialogTitle>
          <DialogDescription>{t("accountDialogDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldCode")}</span>
              <Input
                value={code}
                onChange={(e) => setCode(e.target.value)}
                placeholder={"e.g. 1000"}
              />
            </div>
            <div className="flex flex-col gap-1">
              <span className="text-sm">{t("fieldName")}</span>
              <Input value={name} onChange={(e) => setName(e.target.value)} />
            </div>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldType")}</span>
            <Select value={type} onValueChange={(v) => setType(v as Account["type"])}>
              <SelectTrigger aria-label={t("fieldType")}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {accountTypes.map((tp) => (
                  <SelectItem key={tp} value={tp}>
                    {(t as unknown as (k: string) => string)(`accountType_${tp}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("fieldParentAccount")}</span>
            <Select value={parentId} onValueChange={(v) => setParentId(v ?? "")}>
              <SelectTrigger aria-label={t("fieldParentAccount")}>
                <SelectValue placeholder={t("noneRootAccount")} />
              </SelectTrigger>
              <SelectContent>
                {accounts
                  .filter((a) => a.id !== account?.id)
                  .map((a) => (
                    <SelectItem key={a.id} value={String(a.id)}>
                      {a.code} — {a.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={reconcilable}
              onChange={(e) => setReconcilable(e.target.checked)}
              className="size-4"
            />
            {t("badgeReconcilable")}
          </label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!code || !name}>
            {tCommon("save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
