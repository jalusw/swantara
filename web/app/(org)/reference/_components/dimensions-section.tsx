"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { FolderIcon, Plus } from "lucide-react";
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
  return z.object({
    name: z.string().min(1, "Enter a name."),
    code: z.string(),
    kind: z.enum(kinds),
    parentId: z.string(),
    active: z.boolean(),
  });
}
type DimensionValues = z.infer<ReturnType<typeof useDimensionSchema>>;

export function DimensionsSection({ orgId }: { orgId: string }) {
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
      void getSwantaraService()
        .dimensions.update(Number(orgId), Number(editing.id), payload)
        .then(() => {
          toast.success("Dimension account saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update dimension account", error);
          toast.error("Could not disable the organization.");
        });
    } else {
      void getSwantaraService()
        .dimensions.create(Number(orgId), payload)
        .then(() => {
          toast.success("Dimension account saved.");
          setOpen(false);
          form.reset();
          void query.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create dimension account", error);
          toast.error("Could not disable the organization.");
        });
    }
  }

  function deleteRow(id: string) {
    void getSwantaraService()
      .dimensions.delete(Number(orgId), Number(id))
      .then(() => void query.refetch())
      .catch((error) => {
        logger.error("Failed to delete dimension account", error);
        toast.error("Could not disable the organization.");
      });
  }

  const tree: TreeNode[] = buildTree(accounts, (row) => (
    <RowActions
      editLabel={"Edit"}
      deleteLabel={"Delete"}
      confirmTitle={"Delete this account?"}
      confirmDescription={"The dimension account and its children will be removed."}
      onEdit={() => openEdit(row)}
      onDelete={() => deleteRow(row.id)}
    />
  ));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{"Add account"}</span>
        </Button>
      </div>

      {query.isLoading ? (
        <p className="text-sm text-muted-foreground">{"Loading..."}</p>
      ) : (
        <TreeView
          items={tree}
          defaultExpandedIds={tree.map((node) => node.id)}
          aria-label={"Dimension accounts"}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editing ? "Edit dimension account" : "New dimension account"}
            </DialogTitle>
            <DialogDescription>
              {"Cost centers and dimensions used for reporting."}
            </DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="code" label={"Code"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Code"} />}
              </FormField>
              <FormField name="kind" label={"Kind"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Kind"}>
                      <SelectValue placeholder={"Kind"} />
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
              <FormField name="parentId" label={"Parent account"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Parent account"}>
                      <SelectValue placeholder={"None (top-level)"} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">{"None (top-level)"}</SelectItem>
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
              <FormField name="active" label={"Active"}>
                {({ field }) => (
                  <Switch
                    checked={field.value}
                    onCheckedChange={(checked) => field.onChange(Boolean(checked))}
                    aria-label={"Active"}
                  />
                )}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save account"}</SubmitButton>
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
