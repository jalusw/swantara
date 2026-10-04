"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
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
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { Item, ItemCategory, PriceRule } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

const scopeOptions = ["all", "category", "item", "variant"] as const;
const computeTypeOptions = ["fixed", "percent", "formula"] as const;

function useRuleSchema() {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  return z.object({
    appliesTo: z.enum(scopeOptions),
    itemId: z.string(),
    categoryId: z.string(),
    minQty: z.string().refine((v) => Number(v) >= 0, t("validationNonNegative")),
    computeType: z.enum(computeTypeOptions),
    fixedPrice: z.string(),
    discountPct: z.string(),
    dateStart: z.string(),
    dateEnd: z.string(),
  });
}
type RuleValues = z.infer<ReturnType<typeof useRuleSchema>>;

export function PriceBookDetail({ orgId, priceBookId }: { orgId: string; priceBookId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Products");
  const tCommon = useTranslations("Common");
  const ruleSchema = useRuleSchema();
  const [dialogOpen, setDialogOpen] = useState(false);

  const rulesQuery = useOrgListQuery<{ rules: PriceRule[] }, Record<string, never>>(
    "price_books",
    (organizationId) =>
      getSwantaraService().priceBooks.rules.list(organizationId, Number(priceBookId)),
  );

  const productsQuery = useOrgListQuery<{ products: Item[] }, Record<string, never>>(
    "products",
    (organizationId) => getSwantaraService().products.list(organizationId),
  );

  const categoriesQuery = useOrgListQuery<{ categories: ItemCategory[] }, Record<string, never>>(
    "productCategories",
    (organizationId) => getSwantaraService().productCategories.list(organizationId),
  );

  const rules = rulesQuery.data?.rules ?? [];
  const products = productsQuery.data?.products ?? [];
  const categories = categoriesQuery.data?.categories ?? [];

  const form = useForm<RuleValues>({
    resolver: zodResolver(ruleSchema),
    defaultValues: {
      appliesTo: "all",
      itemId: "",
      categoryId: "",
      minQty: "1",
      computeType: "fixed",
      fixedPrice: "",
      discountPct: "",
      dateStart: "",
      dateEnd: "",
    },
  });

  const watchedScope = form.watch("appliesTo");
  const watchedCompute = form.watch("computeType");

  function handleCreateRule(values: RuleValues) {
    return getSwantaraService()
      .priceBooks.rules.create(Number(orgId), Number(priceBookId), {
        appliesTo: values.appliesTo,
        itemId: values.itemId ? Number(values.itemId) : null,
        categoryId: values.categoryId ? Number(values.categoryId) : null,
        minQty: Number(values.minQty) || 1,
        computeType: values.computeType,
        fixedPrice: values.fixedPrice ? Number(values.fixedPrice) : null,
        discountPct: values.discountPct ? Number(values.discountPct) : null,
        dateStart: values.dateStart ? new Date(values.dateStart) : null,
        dateEnd: values.dateEnd ? new Date(values.dateEnd) : null,
      })
      .then(() => {
        toast.success(t("toastRuleAdded"));
        setDialogOpen(false);
        form.reset();
        void rulesQuery.refetch();
      })
      .catch(() => void toast.error(t("toastRuleFailed")));
  }

  const columns: ColumnDef<PriceRule>[] = [
    {
      accessorKey: "appliesTo",
      header: () => t("fieldScope"),
      cell: ({ row }) => (
        <Badge variant="secondary">
          {(t as unknown as (k: string) => string)(`ruleScope_${row.original.appliesTo}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "itemId",
      header: () => t("fieldItem"),
      cell: ({ row }) => {
        if (!row.original.itemId) return <span className="text-muted-foreground">—</span>;
        const item = products.find((p) => p.id === row.original.itemId);
        return <span className="">{item?.name ?? `#${row.original.itemId}`}</span>;
      },
    },
    {
      accessorKey: "categoryId",
      header: () => t("fieldCategory"),
      cell: ({ row }) => {
        if (!row.original.categoryId) return <span className="text-muted-foreground">—</span>;
        const cat = categories.find((c) => c.id === row.original.categoryId);
        return <span>{cat?.name ?? `#${row.original.categoryId}`}</span>;
      },
    },
    {
      accessorKey: "minQty",
      header: () => t("fieldMinQty"),
      cell: ({ row }) => <span className="tabular-nums">{row.original.minQty}</span>,
    },
    {
      accessorKey: "computeType",
      header: () => t("fieldComputeType"),
      cell: ({ row }) => (
        <Badge>
          {(t as unknown as (k: string) => string)(`computeType_${row.original.computeType}`)}
        </Badge>
      ),
    },
    {
      accessorKey: "fixedPrice",
      header: () => t("fieldPrice"),
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.fixedPrice != null ? row.original.fixedPrice.toFixed(2) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "discountPct",
      header: () => t("fieldDiscountPct"),
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.discountPct != null ? `${row.original.discountPct}%` : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: () => t("fieldStartDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.dateStart ? formatDate(new Date(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: () => t("fieldEndDate"),
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.dateEnd ? formatDate(new Date(row.original.dateEnd)) : "—"}
        </span>
      ),
    },
  ];

  const isLoading = rulesQuery.isLoading || productsQuery.isLoading || categoriesQuery.isLoading;
  const error = rulesQuery.isError
    ? rulesQuery.error
    : productsQuery.isError
      ? productsQuery.error
      : categoriesQuery.isError
        ? categoriesQuery.error
        : null;

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle>{t("pricingRules")}</CardTitle>
            <CardDescription>{t("pricingRulesDescription")}</CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={columns}
            data={rules}
            getRowId={(row) => String(row.id)}
            searchKeys={[]}
            searchPlaceholder={t("searchRules")}
            filterLabel={t("fieldScope")}
            allLabel={t("allRules")}
            ariaLabel={t("allRules")}
            statusOptions={[]}
            status={
              isLoading
                ? { type: "loading" }
                : error
                  ? {
                      type: "error",
                      message: error.message,
                      onRetry: () => void rulesQuery.refetch(),
                    }
                  : undefined
            }
            actions={
              <Button size="sm" onClick={() => setDialogOpen(true)}>
                <Plus />
                <span>{t("addRule")}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{t("newPricingRule")}</DialogTitle>
            <DialogDescription>{t("newPricingRuleDescription")}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleCreateRule}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="appliesTo" label={t("fieldScope")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldScope")}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {scopeOptions.map((scope) => (
                        <SelectItem key={scope} value={scope}>
                          {(t as unknown as (k: string) => string)(`ruleScope_${scope}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="computeType" label={t("fieldComputeType")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("fieldComputeType")}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {computeTypeOptions.map((ct) => (
                        <SelectItem key={ct} value={ct}>
                          {(t as unknown as (k: string) => string)(`computeType_${ct}`)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              {watchedScope === "item" || watchedScope === "variant" ? (
                <FormField name="itemId" label={t("fieldItem")}>
                  {({ field, id }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id={id} aria-label={t("fieldItem")}>
                        <SelectValue placeholder={t("fieldItem")} />
                      </SelectTrigger>
                      <SelectContent>
                        {products.map((p) => (
                          <SelectItem key={p.id} value={String(p.id)}>
                            {p.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormField>
              ) : null}
              {watchedScope === "category" ? (
                <FormField name="categoryId" label={t("fieldCategory")}>
                  {({ field, id }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id={id} aria-label={t("fieldCategory")}>
                        <SelectValue placeholder={t("fieldCategory")} />
                      </SelectTrigger>
                      <SelectContent>
                        {categories.map((c) => (
                          <SelectItem key={c.id} value={String(c.id)}>
                            {c.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormField>
              ) : null}
              <FormField name="minQty" label={t("fieldMinQty")}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" min="0" inputMode="decimal" />
                )}
              </FormField>
              {watchedCompute === "fixed" ? (
                <FormField name="fixedPrice" label={t("fieldPrice")}>
                  {({ field, id }) => (
                    <Input
                      {...field}
                      id={id}
                      type="number"
                      step="any"
                      min="0"
                      inputMode="decimal"
                    />
                  )}
                </FormField>
              ) : null}
              {watchedCompute === "percent" ? (
                <FormField name="discountPct" label={t("fieldDiscountPct")}>
                  {({ field, id }) => (
                    <Input
                      {...field}
                      id={id}
                      type="number"
                      step="any"
                      min="0"
                      max="100"
                      inputMode="decimal"
                    />
                  )}
                </FormField>
              ) : null}
              <FormField name="dateStart" label={t("fieldStartDate")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={t("fieldEndDate")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                {tCommon("cancel")}
              </Button>
              <SubmitButton>{t("saveRule")}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
