"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
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
import { humanizeKey } from "@/lib/utils/case";

const scopeOptions = ["all", "category", "item", "variant"] as const;
const computeTypeOptions = ["fixed", "percent", "formula"] as const;

function computeTypeLabel(computeType: string): string {
  if (computeType === "fixed") return "Fixed price";
  if (computeType === "percent") return "Percent discount";
  return humanizeKey(computeType);
}

const ruleSchema = z.object({
  appliesTo: z.enum(scopeOptions),
  itemId: z.string(),
  categoryId: z.string(),
  minQty: z.string().refine((v) => Number(v) >= 0),
  computeType: z.enum(computeTypeOptions),
  fixedPrice: z.string(),
  discountPct: z.string(),
  dateStart: z.string(),
  dateEnd: z.string(),
});
type RuleValues = z.infer<typeof ruleSchema>;

export function PriceBookDetail({ orgId, priceBookId }: { orgId: string; priceBookId: string }) {
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
    void getSwantaraService()
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
        toast.success("Rule added.");
        setDialogOpen(false);
        form.reset();
        void rulesQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  const columns: ColumnDef<PriceRule>[] = [
    {
      accessorKey: "appliesTo",
      header: "Scope",
      cell: ({ row }) => (
        <Badge variant="secondary">{humanizeKey(String(row.original.appliesTo))}</Badge>
      ),
    },
    {
      accessorKey: "itemId",
      header: "Item",
      cell: ({ row }) => {
        if (!row.original.itemId) return <span className="text-muted-foreground">—</span>;
        const item = products.find((p) => p.id === row.original.itemId);
        return <span className="">{item?.name ?? `#${row.original.itemId}`}</span>;
      },
    },
    {
      accessorKey: "categoryId",
      header: "Category",
      cell: ({ row }) => {
        if (!row.original.categoryId) return <span className="text-muted-foreground">—</span>;
        const cat = categories.find((c) => c.id === row.original.categoryId);
        return <span>{cat?.name ?? `#${row.original.categoryId}`}</span>;
      },
    },
    {
      accessorKey: "minQty",
      header: "Min. quantity",
      cell: ({ row }) => <span className="tabular-nums">{row.original.minQty}</span>,
    },
    {
      accessorKey: "computeType",
      header: "Compute type",
      cell: ({ row }) => <Badge>{computeTypeLabel(String(row.original.computeType))}</Badge>,
    },
    {
      accessorKey: "fixedPrice",
      header: "Price",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.fixedPrice != null ? row.original.fixedPrice.toFixed(2) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "discountPct",
      header: "Discount %",
      cell: ({ row }) => (
        <span className="tabular-nums">
          {row.original.discountPct != null ? `${row.original.discountPct}%` : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateStart",
      header: "Valid from",
      cell: ({ row }) => (
        <span className="text-muted-foreground">
          {row.original.dateStart ? formatDate(new Date(row.original.dateStart)) : "—"}
        </span>
      ),
    },
    {
      accessorKey: "dateEnd",
      header: "Valid to",
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
            <CardTitle>{"Pricing rules"}</CardTitle>
            <CardDescription>
              {"Rules that determine prices for products, categories, or all items."}
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={columns}
            data={rules}
            getRowId={(row) => String(row.id)}
            searchKeys={[]}
            searchPlaceholder={"Search rules…"}
            filterLabel={"Scope"}
            allLabel={"All rules"}
            ariaLabel={"All rules"}
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
                <span>{"Add rule"}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{"New pricing rule"}</DialogTitle>
            <DialogDescription>{"Add a rule to this price_book."}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleCreateRule}>
            <div className="grid gap-4 sm:grid-cols-2">
              <FormField name="appliesTo" label={"Scope"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Scope"}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {scopeOptions.map((scope) => (
                        <SelectItem key={scope} value={scope}>
                          {humanizeKey(String(scope))}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="computeType" label={"Compute type"}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={"Compute type"}>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {computeTypeOptions.map((ct) => (
                        <SelectItem key={ct} value={ct}>
                          {computeTypeLabel(String(ct))}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              {watchedScope === "item" || watchedScope === "variant" ? (
                <FormField name="itemId" label={"Item"}>
                  {({ field, id }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id={id} aria-label={"Item"}>
                        <SelectValue placeholder={"Item"} />
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
                <FormField name="categoryId" label={"Category"}>
                  {({ field, id }) => (
                    <Select value={field.value} onValueChange={field.onChange}>
                      <SelectTrigger id={id} aria-label={"Category"}>
                        <SelectValue placeholder={"Category"} />
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
              <FormField name="minQty" label={"Min. quantity"}>
                {({ field, id }) => (
                  <Input {...field} id={id} type="number" min="0" inputMode="decimal" />
                )}
              </FormField>
              {watchedCompute === "fixed" ? (
                <FormField name="fixedPrice" label={"Price"}>
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
                <FormField name="discountPct" label={"Discount %"}>
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
              <FormField name="dateStart" label={"Valid from"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="dateEnd" label={"Valid to"}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save rule"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
