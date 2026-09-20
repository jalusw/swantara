"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import type { ColumnDef } from "@tanstack/react-table";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { InteractiveEntityTable } from "@/app/(org)/_components/interactive-entity-table";
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
import { Switch } from "@/components/switch";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { PriceBook } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { activeColumn, nameColumn } from "@/lib/utils/table-columns";

const currencyOptions = ["USD", "IDR", "EUR", "SGD", "GBP", "JPY"];

function usePriceBookFormSchema() {
  return z.object({
    name: z.string().min(1, "Enter a price_book name."),
    currencyCode: z.string(),
    active: z.boolean(),
  });
}
type PriceBookFormValues = z.infer<ReturnType<typeof usePriceBookFormSchema>>;

export function PriceBooksSection({ orgId }: { orgId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);

  const query = useOrgListQuery<{ priceBooks: PriceBook[] }, Record<string, never>>(
    "price_books",
    (organizationId) => getSwantaraService().priceBooks.list(organizationId),
  );

  const price_books = query.data?.priceBooks ?? [];

  const schema = usePriceBookFormSchema();

  const form = useForm<PriceBookFormValues>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", currencyCode: "USD", active: true },
  });

  function handleCreate(values: PriceBookFormValues) {
    void getSwantaraService()
      .priceBooks.create(Number(orgId), {
        name: values.name,
        currencyCode: values.currencyCode || null,
        organizationId: Number(orgId),
        active: values.active,
      })
      .then(() => {
        toast.success("Rule added.");
        setDialogOpen(false);
        form.reset();
        void query.refetch();
      });
  }

  const columns: ColumnDef<PriceBook>[] = [
    nameColumn<PriceBook>({
      basePath: "price_books",
      header: "Name",
    }),
    {
      accessorKey: "currencyCode",
      header: "Currency",
      cell: ({ row }) => (
        <span className="text-muted-foreground">{row.original.currencyCode ?? "—"}</span>
      ),
    },
    activeColumn<PriceBook>({
      header: "Status",
      activeLabel: "Active",
      inactiveLabel: "Inactive",
    }),
  ];

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <Card>
        <CardHeader className="flex-row items-start justify-between gap-4">
          <div className="flex flex-col gap-1">
            <CardTitle>{"PriceBooks"}</CardTitle>
            <CardDescription>
              {"Manage pricing rules for products across customers and channels."}
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <InteractiveEntityTable
            columns={columns}
            data={price_books}
            getRowId={(row) => String(row.id)}
            searchKeys={["name"]}
            statusKey="active"
            statusOptions={[
              { value: "true", label: "Active" },
              { value: "false", label: "Inactive" },
            ]}
            searchPlaceholder={"Search price_books…"}
            filterLabel={"Filter by status"}
            allLabel={"All statuses"}
            ariaLabel={"All price_books"}
            status={
              query.isLoading
                ? { type: "loading" }
                : query.isError
                  ? {
                      type: "error",
                      message: query.error.message,
                      onRetry: () => void query.refetch(),
                    }
                  : undefined
            }
            actions={
              <Button size="sm" onClick={() => setDialogOpen(true)}>
                <Plus />
                <span>{"Add price_book"}</span>
              </Button>
            }
          />
        </CardContent>
      </Card>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{"New price_book"}</DialogTitle>
            <DialogDescription>{"Create a new price_book with a base currency."}</DialogDescription>
          </DialogHeader>
          <Form form={form} onSubmit={handleCreate}>
            <div className="flex flex-col gap-4">
              <FormField name="name" label={"Name"}>
                {({ field, id }) => <Input {...field} id={id} placeholder={"Name"} />}
              </FormField>
              <FormField name="currencyCode" label={"Currency"}>
                {({ field, id }) => (
                  <select
                    id={id}
                    value={field.value}
                    onChange={field.onChange}
                    className="rounded-md border border-input bg-background px-3 py-2 text-sm"
                  >
                    {currencyOptions.map((code) => (
                      <option key={code} value={code}>
                        {code}
                      </option>
                    ))}
                  </select>
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
              <Button variant="outline" onClick={() => setDialogOpen(false)}>
                {"Cancel"}
              </Button>
              <SubmitButton>{"Save"}</SubmitButton>
            </DialogFooter>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
