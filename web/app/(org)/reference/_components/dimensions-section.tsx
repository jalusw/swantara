"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { FolderIcon, Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Button } from "@/components/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Switch } from "@/components/switch";
import type { TreeNode } from "@/components/tree-view";
import { TreeView } from "@/components/tree-view";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Dimension } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";

export type DimensionRow = {
  id: string;
  name: string;
  code: string;
  kind: "expense" | "revenue" | "cost" | "other";
  parentId: string | null;
  active: boolean;
};

function toDimensionRow(account: Dimension): DimensionRow {
  return {
    id: String(account.id),
    name: account.name,
    code: account.code ?? "",
    kind: (account.kind ?? "other") as DimensionRow["kind"],
    parentId: account.parentId != null ? String(account.parentId) : null,
    active: account.active ?? true,
  };
}

const kinds = ["expense", "revenue", "cost", "other"] as const;

function useDimensionSchema() {
  const t = useTranslations("Reference");
  return z.object({
    name: z.string().min(1, t("validationNameRequired")),
    code: z.string(),
    kind: z.enum(kinds),
    parentId: z.string(),
    active: z.boolean(),
  });
}
type DimensionValues = z.infer<ReturnType<typeof useDimensionSchema>>;

export function DimensionsSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Reference");
  const [editing, setEditing] = useState<DimensionRow | null>(null);
  const [open, setOpen] = useState(false);

  const query = useOrgListQuery<{ accounts: Dimension[] }, Record<string, never>>(
    "dimensions",
    (organizationId) => getSwantaraService().dimensions.list(organizationId),
  );

  const accounts = (query.data?.accounts ?? []).map(toDimensionRow);

  const schema = useDimensionSchema();

  const form = useForm<DimensionValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      code: "",
      kind: "expense",
      parentId: "",
      active: true,
    },
  });

  function openCreate() {
    setEditing(null);
    form.reset({
      name: "",
      code: "",
      kind: "expense",
      parentId: "",
      active: true,
    });
    setOpen(true);
  }

  function openEdit(row: DimensionRow) {
    setEditing(row);
    form.reset({
      name: row.name,
      code: row.code,
      kind: row.kind,
      parentId: row.parentId ?? "",
      active: row.active,
    });
    setOpen(true);
  }

  function handleSubmit(values: DimensionValues) {
    const payload = {
      name: values.name,
      code: values.code || undefined,
      kind: values.kind as Dimension["kind"],
      parentId: values.parentId === "" ? null : Number(values.parentId),
      active: values.active,
    };

    if (editing) {
      return getSwantaraService()
        .dimensions.update(Number(orgId), Number(editing.id), payload)
        .then(() => {
          toast.success(t("dimensionSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update dimension account", error);
          toast.error(t("toastFailed"));
        });
    } else {
      return getSwantaraService()
        .dimensions.create(Number(orgId), payload)
        .then(() => {
          toast.success(t("dimensionSaved"));
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create dimension account", error);
          toast.error(t("toastFailed"));
        });
    }
  }

  function deleteRow(id: string) {
    void getSwantaraService()
      .dimensions.delete(Number(orgId), Number(id))
      .then(() => void query.refetch())
      .catch((error) => {
        logger.error("Failed to delete dimension account", error);
        toast.error(t("toastFailed"));
      });
  }

  const tree: TreeNode[] = buildTree(accounts, (row) => (
    <RowActions
      editLabel={t("actionEdit")}
      deleteLabel={t("actionDelete")}
      confirmTitle={t("deleteAccountTitle")}
      confirmDescription={t("deleteAccountDescription")}
      onEdit={() => openEdit(row)}
      onDelete={() => deleteRow(row.id)}
    />
  ));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{t("addAccount")}</span>
        </Button>
      </div>

      {query.isLoading ? (
        <p className="text-sm text-muted-foreground">{t("loading")}</p>
      ) : (
        <TreeView
          items={tree}
          defaultExpandedIds={tree.map((node) => node.id)}
          aria-label={t("dimensionsTitle")}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? t("editAccount") : t("newAccount")}</DialogTitle>
            <DialogDescription>{t("dimensionDialogDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("fieldName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
              </FormField>
              <FormField name="code" label={t("fieldCode")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldCode")} />}
              </FormField>
              <FormField name="kind" label={t("fieldType")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldType")}>
                      <SelectValue placeholder={t("fieldType")} />
                    </SelectTrigger>
                    <SelectContent>
                      {kinds.map((kind) => (
                        <SelectItem key={kind} value={kind}>
                          {String(kind)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="parentId" label={t("fieldParentAccount")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldParentAccount")}>
                      <SelectValue placeholder={t("noParentOption")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">{t("noParentOption")}</SelectItem>
                      {accounts
                        .filter((row) => row.id !== editing?.id)
                        .map((row) => (
                          <SelectItem key={row.id} value={row.id}>
                            {row.name}
                          </SelectItem>
                        ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="active" label={t("fieldActive")}>
                {({ field }) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                    aria-label={t("fieldActive")}
                  />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {t("actionCancel")}
              </Button>
              <SubmitButton>{t("saveAccount")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function buildTree(
  rows: DimensionRow[],
  actionsFor: (row: DimensionRow) => React.ReactNode,
): TreeNode[] {
  const childrenOf = (parentId: string | null) =>
    rows
      .filter((row) => row.parentId === parentId)
      .map(
        (row): TreeNode => ({
          id: row.id,
          icon: <FolderIcon className="size-4 text-muted-foreground" aria-hidden />,
          label: (
            <span className="flex items-center gap-2">
              <span className="">{row.name}</span>
              {row.code ? (
                <span className="font-mono text-xs text-muted-foreground">{row.code}</span>
              ) : null}
            </span>
          ),
          children: childrenOf(row.id),
          actions: actionsFor(row),
        }),
      );

  return childrenOf(null);
}
