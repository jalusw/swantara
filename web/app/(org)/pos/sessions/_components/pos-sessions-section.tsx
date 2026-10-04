"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
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
import type { PosConfig, PosSession } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney } from "@/lib/utils";
import { useSession } from "@/providers/session";
import { posSessionStateTone } from "../../_components/pos-utils";

export function PosSessionsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Pos");
  const sessionState = (state: string) =>
    (t as unknown as (k: string) => string)(`sessionState_${state}`);
  const { user } = useSession();
  const [dialogOpen, setDialogOpen] = useState(false);

  const sessionsQuery = useOrgListQuery<{ sessions: PosSession[] }, Record<string, never>>(
    "posSessions",
    (organizationId) => getSwantaraService().posSessions.list(organizationId),
  );
  const configsQuery = useOrgListQuery<{ configs: PosConfig[] }, Record<string, never>>(
    "posConfigs",
    (organizationId) => getSwantaraService().posConfigs.list(organizationId),
  );

  const sessions = sessionsQuery.data?.sessions ?? [];
  const configs = configsQuery.data?.configs ?? [];
  const configMap = new Map(configs.map((c) => [c.id, c.name || `POS-${c.id}`]));

  function handleSaved() {
    setDialogOpen(false);
    void sessionsQuery.refetch();
  }

  const columns: ColumnDef<PosSession>[] = [
    {
      accessorKey: "id",
      header: t("colSession"),
      cell: ({ row }) => (
        <a
          href={`/pos/sessions/${row.original.id}`}
          className=" rounded outline-none focus-visible:ring-2 focus-visible:ring-ring"
        >
          {t("sessionFallback", { id: row.original.id })}
        </a>
      ),
    },
    {
      accessorKey: "configId",
      header: t("colPosConfig"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {configMap.get(row.original.configId) ?? `#${row.original.configId}`}
        </span>
      ),
    },
    {
      accessorKey: "openedAt",
      header: t("colOpenedAt"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.openedAt ? formatDate(row.original.openedAt) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "openingBalance",
      header: t("colOpeningBalance"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {formatMoney(row.original.openingBalance)}
        </span>
      ),
    },
    {
      accessorKey: "closingBalance",
      header: t("colClosingBalance"),
      meta: { align: "right" },
      cell: ({ row }) => (
        <span className="tabular-nums text-muted-foreground">
          {row.original.closingBalance != null ? formatMoney(row.original.closingBalance) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "state",
      header: t("colStatus"),
      cell: ({ row }) => (
        <Badge variant="outline" className={posSessionStateTone(row.original.state)}>
          {sessionState(row.original.state)}
        </Badge>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={sessions}
        getRowId={(row) => String(row.id)}
        searchKeys={["id"]}
        statusKey="state"
        statusOptions={[
          { value: "opened", label: sessionState("opened") },
          { value: "closing", label: sessionState("closing") },
          { value: "closed", label: sessionState("closed") },
        ]}
        searchPlaceholder={t("sessionsSearchPlaceholder")}
        filterLabel={t("colStatus")}
        allLabel={t("allSessions")}
        ariaLabel={t("sessionsTitle")}
        emptyTitle={t("sessionsEmpty")}
        status={
          sessionsQuery.isLoading
            ? { type: "loading" }
            : sessionsQuery.isError
              ? {
                  type: "error",
                  message: sessionsQuery.error.message,
                  onRetry: () => void sessionsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("openSession")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <PosSessionOpenDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          configs={configs}
          cashierId={user?.id ?? 0}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function PosSessionOpenDialog({
  open,
  onOpenChange,
  orgId,
  configs,
  cashierId,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  configs: PosConfig[];
  cashierId: number;
  onSave: () => void;
}) {
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
  const [configId, setConfigId] = useState("");
  const [openingBalance, setOpeningBalance] = useState("0");

  function handleSubmit() {
    if (!configId) return;
    void getSwantaraService()
      .posSessions.create(Number(orgId), {
        configId: Number(configId),
        cashierId,
        openingBalance: Number(openingBalance) || 0,
      })
      .then(() => {
        toast.success(t("sessionOpened"));
        onSave();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("openSession")}</DialogTitle>
          <DialogDescription>{t("openSessionDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colPosConfig")}</span>
            <Select value={configId} onValueChange={(v) => setConfigId(v ?? "")}>
              <SelectTrigger aria-label={t("colPosConfig")}>
                <SelectValue placeholder={t("selectPosConfig")} />
              </SelectTrigger>
              <SelectContent>
                {configs.map((c) => (
                  <SelectItem key={c.id} value={String(c.id)}>
                    {c.name || `POS-${c.id}`}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colOpeningBalance")}</span>
            <Input
              type="number"
              step="any"
              min="0"
              value={openingBalance}
              onChange={(e) => setOpeningBalance(e.target.value)}
            />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {tCommon("cancel")}
          </Button>
          <Button onClick={handleSubmit} disabled={!configId}>
            {t("openAction")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
