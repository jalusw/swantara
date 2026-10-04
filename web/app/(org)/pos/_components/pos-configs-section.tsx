"use client";

import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { toast } from "sonner";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
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
import type { Journal, PosConfig, PriceBook, Warehouse } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";

export function PosConfigsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Pos");
  const [dialogOpen, setDialogOpen] = useState(false);

  const configsQuery = useOrgListQuery<{ configs: PosConfig[] }, Record<string, never>>(
    "posConfigs",
    (organizationId) => getSwantaraService().posConfigs.list(organizationId),
  );
  const warehousesQuery = useOrgListQuery<{ warehouses: Warehouse[] }, Record<string, never>>(
    "warehouses",
    (organizationId) => getSwantaraService().inventory.warehouses(organizationId),
  );
  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );
  const price_booksQuery = useOrgListQuery<{ priceBooks: PriceBook[] }, Record<string, never>>(
    "price_books",
    (organizationId) => getSwantaraService().priceBooks.list(organizationId),
  );

  const configs = configsQuery.data?.configs ?? [];
  const warehouseMap = new Map((warehousesQuery.data?.warehouses ?? []).map((w) => [w.id, w.name]));
  const journalMap = new Map((journalsQuery.data?.journals ?? []).map((j) => [j.id, j.name]));
  const price_bookMap = new Map(
    (price_booksQuery.data?.priceBooks ?? []).map((p) => [p.id, p.name]),
  );

  function handleSaved() {
    setDialogOpen(false);
    void configsQuery.refetch();
  }

  const columns: ColumnDef<PosConfig>[] = [
    {
      accessorKey: "name",
      header: t("colConfigName"),
      cell: ({ row }) => <span className="">{row.original.name || `POS-${row.original.id}`}</span>,
    },
    {
      accessorKey: "warehouseId",
      header: t("colWarehouse"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.warehouseId
            ? (warehouseMap.get(row.original.warehouseId) ?? `#${row.original.warehouseId}`)
            : "—"}
        </span>
      ),
    },
    {
      accessorKey: "journalId",
      header: t("colJournal"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.journalId
            ? (journalMap.get(row.original.journalId) ?? `#${row.original.journalId}`)
            : "—"}
        </span>
      ),
    },
    {
      accessorKey: "priceBookId",
      header: t("colPriceBook"),
      cell: ({ row }) => (
        <span className="text-muted-foreground text-sm">
          {row.original.priceBookId
            ? (price_bookMap.get(row.original.priceBookId) ?? `#${row.original.priceBookId}`)
            : "—"}
        </span>
      ),
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <InteractiveEntityTable
        columns={columns}
        data={configs}
        getRowId={(row) => String(row.id)}
        searchKeys={["name"]}
        searchPlaceholder={t("configsSearchPlaceholder")}
        ariaLabel={t("configsTitle")}
        emptyTitle={t("configsEmpty")}
        status={
          configsQuery.isLoading
            ? { type: "loading" }
            : configsQuery.isError
              ? {
                  type: "error",
                  message: configsQuery.error.message,
                  onRetry: () => void configsQuery.refetch(),
                }
              : undefined
        }
        actions={
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus />
            <span>{t("newConfig")}</span>
          </Button>
        }
      />
      {dialogOpen ? (
        <PosConfigFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          warehouses={warehousesQuery.data?.warehouses ?? []}
          journals={journalsQuery.data?.journals ?? []}
          price_books={price_booksQuery.data?.priceBooks ?? []}
          onSave={handleSaved}
        />
      ) : null}
    </div>
  );
}

function PosConfigFormDialog({
  open,
  onOpenChange,
  orgId,
  warehouses,
  journals,
  price_books,
  onSave,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  orgId: string;
  warehouses: Warehouse[];
  journals: Journal[];
  price_books: PriceBook[];
  onSave: () => void;
}) {
  const t = useTranslations("Pos");
  const tCommon = useTranslations("Common");
  const [name, setName] = useState("");
  const [warehouseId, setWarehouseId] = useState("");
  const [journalId, setJournalId] = useState("");
  const [priceBookId, setPriceBookId] = useState("");

  function handleSubmit() {
    if (!name) return;
    void getSwantaraService()
      .posConfigs.create(Number(orgId), {
        organizationId: Number(orgId),
        name,
        warehouseId: warehouseId ? Number(warehouseId) : null,
        journalId: journalId ? Number(journalId) : null,
        priceBookId: priceBookId ? Number(priceBookId) : null,
      })
      .then(() => {
        toast.success(t("configCreated"));
        onSave();
      })
      .catch(() => toast.error(t("saveFailed")));
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t("newConfig")}</DialogTitle>
          <DialogDescription>{t("configFormDescription")}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colConfigName")}</span>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder={t("configNamePlaceholder")}
            />
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colWarehouse")}</span>
            <Select value={warehouseId} onValueChange={(v) => setWarehouseId(v ?? "")}>
              <SelectTrigger aria-label={t("colWarehouse")}>
                <SelectValue placeholder={t("selectWarehouse")} />
              </SelectTrigger>
              <SelectContent>
                {warehouses.map((w) => (
                  <SelectItem key={w.id} value={String(w.id)}>
                    {w.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex flex-col gap-1">
            <span className="text-sm">{t("colJournal")}</span>
            <Select value={journalId} onValueChange={(v) => setJournalId(v ?? "")}>
              <SelectTrigger aria-label={t("colJournal")}>
                <SelectValue placeholder={t("selectJournal")} />
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
            <span className="text-sm">{t("colPriceBook")}</span>
            <Select value={priceBookId} onValueChange={(v) => setPriceBookId(v ?? "")}>
              <SelectTrigger aria-label={t("colPriceBook")}>
                <SelectValue placeholder={t("selectPriceBook")} />
              </SelectTrigger>
              <SelectContent>
                {price_books.map((p) => (
                  <SelectItem key={p.id} value={String(p.id)}>
                    {p.name}
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
