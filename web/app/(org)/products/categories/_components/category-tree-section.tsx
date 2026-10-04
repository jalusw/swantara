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
import type { TreeNode } from "@/components/tree-view";
import { TreeView } from "@/components/tree-view";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { ItemCategory } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";
import {
  costMethods,
  type StubItemCategory,
  valuationModes,
} from "../../_components/products-data";

type TFn = (key: string, values?: Record<string, string | number>) => string;

type AccountField =
  | "incomeAccountId"
  | "expenseAccountId"
  | "stockCostAccountId"
  | "stockInputAccountId"
  | "stockOutputAccountId"
  | "cogsAccountId";

const accountFields: AccountField[] = [
  "incomeAccountId",
  "expenseAccountId",
  "stockCostAccountId",
  "stockInputAccountId",
  "stockOutputAccountId",
  "cogsAccountId",
];

function useCategorySchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  return z.object({
    name: z.string().min(1, t("validationCategoryNameRequired")),
    parentId: z.string(),
    costMethod: z.enum(["none", ...costMethods]),
    valuation: z.enum(["none", ...valuationModes]),
    incomeAccountId: z.string(),
    expenseAccountId: z.string(),
    stockCostAccountId: z.string(),
    stockInputAccountId: z.string(),
    stockOutputAccountId: z.string(),
    cogsAccountId: z.string(),
  });
}

type Values = z.infer<ReturnType<typeof useCategorySchema>>;

function toCategoryRow(category: ItemCategory): StubItemCategory {
  return {
    id: String(category.id),
    name: category.name,
    parentId: category.parentId != null ? String(category.parentId) : null,
    incomeAccountId: category.incomeAccountId != null ? String(category.incomeAccountId) : null,
    expenseAccountId: category.expenseAccountId != null ? String(category.expenseAccountId) : null,
    stockCostAccountId:
      category.stockCostAccountId != null ? String(category.stockCostAccountId) : null,
    stockInputAccountId:
      category.stockInputAccountId != null ? String(category.stockInputAccountId) : null,
    stockOutputAccountId:
      category.stockOutputAccountId != null ? String(category.stockOutputAccountId) : null,
    cogsAccountId: category.cogsAccountId != null ? String(category.cogsAccountId) : null,
    costMethod: category.costMethod ?? null,
    valuation: category.valuation ?? null,
  };
}

export function CategoryTreeSection({ orgId }: { orgId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const tCommon = useTranslations("Common");
  const schema = useCategorySchema();
  const [editing, setEditing] = useState<StubItemCategory | null>(null);
  const [open, setOpen] = useState(false);

  const categoriesQuery = useOrgListQuery<{ categories: ItemCategory[] }, Record<string, never>>(
    "productCategories",
    (organizationId) => getSwantaraService().productCategories.list(organizationId),
  );

  const categories = (categoriesQuery.data?.categories ?? []).map(toCategoryRow);

  const accountsQuery = useOrgListQuery<
    { accounts: { id: number; name: string }[] },
    Record<string, never>
  >("accounts", (organizationId) => getSwantaraService().accounts.list(organizationId));
  const glAccountOptions = (accountsQuery.data?.accounts ?? []).map((account) => ({
    id: String(account.id),
    name: account.name,
  }));

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues(),
  });

  function openCreate() {
    setEditing(null);
    form.reset(emptyValues());
    setOpen(true);
  }

  function openEdit(category: StubItemCategory) {
    setEditing(category);
    form.reset({
      name: category.name,
      parentId: category.parentId ?? "",
      costMethod: category.costMethod ?? "none",
      valuation: category.valuation ?? "none",
      incomeAccountId: category.incomeAccountId ?? "",
      expenseAccountId: category.expenseAccountId ?? "",
      stockCostAccountId: category.stockCostAccountId ?? "",
      stockInputAccountId: category.stockInputAccountId ?? "",
      stockOutputAccountId: category.stockOutputAccountId ?? "",
      cogsAccountId: category.cogsAccountId ?? "",
    });
    setOpen(true);
  }

  function handleSubmit(values: Values) {
    const accounts = Object.fromEntries(
      accountFields.map((field) => [field, values[field] === "" ? null : values[field]]),
    ) as Pick<StubItemCategory, AccountField>;

    if (editing) {
      return getSwantaraService()
        .productCategories.update(Number(orgId), Number(editing.id), {
          name: values.name,
          parentId: values.parentId === "" ? null : Number(values.parentId),
          costMethod: values.costMethod === "none" ? null : values.costMethod,
          valuation: values.valuation === "none" ? null : values.valuation,
          incomeAccountId: accounts.incomeAccountId ? Number(accounts.incomeAccountId) : null,
          expenseAccountId: accounts.expenseAccountId ? Number(accounts.expenseAccountId) : null,
          stockCostAccountId: accounts.stockCostAccountId
            ? Number(accounts.stockCostAccountId)
            : null,
          stockInputAccountId: accounts.stockInputAccountId
            ? Number(accounts.stockInputAccountId)
            : null,
          stockOutputAccountId: accounts.stockOutputAccountId
            ? Number(accounts.stockOutputAccountId)
            : null,
          cogsAccountId: accounts.cogsAccountId ? Number(accounts.cogsAccountId) : null,
        })
        .then(() => {
          toast.success(t("toastCategorySaved"));
          setOpen(false);
          void categoriesQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to update category", error);
          toast.error(t("toastCategoryFailed"));
        });
    } else {
      return getSwantaraService()
        .productCategories.create(Number(orgId), {
          name: values.name,
          parentId: values.parentId === "" ? null : Number(values.parentId),
          costMethod: values.costMethod === "none" ? null : values.costMethod,
          valuation: values.valuation === "none" ? null : values.valuation,
          incomeAccountId: accounts.incomeAccountId ? Number(accounts.incomeAccountId) : null,
          expenseAccountId: accounts.expenseAccountId ? Number(accounts.expenseAccountId) : null,
          stockCostAccountId: accounts.stockCostAccountId
            ? Number(accounts.stockCostAccountId)
            : null,
          stockInputAccountId: accounts.stockInputAccountId
            ? Number(accounts.stockInputAccountId)
            : null,
          stockOutputAccountId: accounts.stockOutputAccountId
            ? Number(accounts.stockOutputAccountId)
            : null,
          cogsAccountId: accounts.cogsAccountId ? Number(accounts.cogsAccountId) : null,
        })
        .then(() => {
          toast.success(t("toastCategorySaved"));
          setOpen(false);
          void categoriesQuery.refetch();
        })
        .catch((error) => {
          logger.error("Failed to create category", error);
          toast.error(t("toastCategoryFailed"));
        });
    }
  }

  function deleteCategory(category: StubItemCategory) {
    if (categories.some((row) => row.parentId === category.id)) {
      toast.error(t("deleteCategoryBlocked"));
      return;
    }
    void getSwantaraService()
      .productCategories.delete(Number(orgId), Number(category.id))
      .then(() => {
        void categoriesQuery.refetch();
      })
      .catch((error) => {
        logger.error("Failed to delete category", error);
        toast.error(t("toastCategoryFailed"));
      });
  }

  const tree: TreeNode[] = buildTree(categories, (category) => (
    <RowActions
      editLabel={tCommon("edit")}
      deleteLabel={tCommon("delete")}
      confirmTitle={t("deleteCategoryTitle")}
      confirmDescription={t("deleteCategoryDescription")}
      confirmLabel={tCommon("delete")}
      onEdit={() => openEdit(category)}
      onDelete={() => deleteCategory(category)}
    />
  ));

  return (
    <div className="flex flex-col gap-4">
      <div className="flex justify-end">
        <Button size="sm" onClick={openCreate}>
          <Plus />
          <span>{t("addCategory")}</span>
        </Button>
      </div>

      {categoriesQuery.isLoading ? (
        <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>
      ) : tree.length === 0 ? (
        <p className="text-sm text-muted-foreground">{t("categoriesEmpty")}</p>
      ) : (
        <TreeView
          items={tree}
          defaultExpandedIds={categories.map((category) => category.id)}
          aria-label={t("categoriesTitle")}
        />
      )}

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>{editing ? t("editCategory") : t("newCategory")}</DialogTitle>
            <DialogDescription>{t("categoriesDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleSubmit}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="name" label={t("fieldName")}>
                {({ field, id }) => <Input {...field} id={id} placeholder={t("fieldName")} />}
              </FormField>
              <FormField name="parentId" label={t("fieldParentCategory")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldParentCategory")}>
                      <SelectValue placeholder={t("noParent")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">{t("noParent")}</SelectItem>
                      {categories
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
              <FormField name="costMethod" label={t("fieldCostMethod")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldCostMethod")}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {["none", ...costMethods].map((method) => (
                        <SelectItem key={method} value={method}>
                          {(t as unknown as (k: string) => string)(`costMethod_${method}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="valuation" label={t("fieldValuation")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldValuation")}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {["none", ...valuationModes].map((mode) => (
                        <SelectItem key={mode} value={mode}>
                          {(t as unknown as (k: string) => string)(`valuationMode_${mode}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              {accountFields.map((field) => (
                <FormField
                  key={field}
                  name={field}
                  label={(t as unknown as (k: string) => string)(`categoryAccountField_${field}`)}
                >
                  {({ field: accountField, id }) => (
                    <Select value={accountField.value} onValueChange={accountField.onChange}>
                      <SelectTrigger
                        id={id}
                        aria-label={(t as unknown as (k: string) => string)(
                          `categoryAccountField_${field}`,
                        )}
                      >
                        <SelectValue placeholder={t("noAccount")} />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="">{t("noAccount")}</SelectItem>
                        {glAccountOptions.map((account) => (
                          <SelectItem key={account.id} value={account.id}>
                            {account.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormField>
              ))}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                {tCommon("cancel")}
              </Button>
              <SubmitButton>{t("saveCategory")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

function emptyValues(): Values {
  return {
    name: "",
    parentId: "",
    costMethod: "none",
    valuation: "none",
    incomeAccountId: "",
    expenseAccountId: "",
    stockCostAccountId: "",
    stockInputAccountId: "",
    stockOutputAccountId: "",
    cogsAccountId: "",
  };
}

function buildTree(
  rows: StubItemCategory[],
  actionsFor: (category: StubItemCategory) => React.ReactNode,
): TreeNode[] {
  const childrenOf = (parentId: string | null): TreeNode[] =>
    rows
      .filter((row) => row.parentId === parentId)
      .map((row) => ({
        id: row.id,
        icon: <FolderIcon className="size-4 text-muted-foreground" aria-hidden />,
        label: <span className="">{row.name}</span>,
        children: childrenOf(row.id),
        actions: actionsFor(row),
      }));

  return childrenOf(null);
}
