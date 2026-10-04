"use client";

import { useTranslations } from "next-intl";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/table";
import { cn } from "@/lib/utils/style";

import type { AttributeOption } from "./attribute-editor";

export function generateSku(templateName: string, attributes: Record<string, string>) {
  const base = templateName
    .replace(/[^a-zA-Z0-9 ]/g, "")
    .split(" ")
    .filter(Boolean)
    .map((word) => word.slice(0, 3))
    .join("")
    .slice(0, 6)
    .toUpperCase();
  const suffix = Object.values(attributes)
    .join("-")
    .replace(/[^a-zA-Z0-9-]/g, "")
    .toUpperCase();
  return suffix ? `${base}-${suffix}` : base;
}

export function buildMatrix(attributes: AttributeOption[]) {
  const usable = attributes.filter(
    (attribute) => attribute.name.trim() && attribute.values.length > 0,
  );
  return usable.reduce<Record<string, string>[]>(
    (combinations, attribute) =>
      combinations.flatMap((combination) =>
        attribute.values.map((value) => ({
          ...combination,
          [attribute.name.trim()]: value,
        })),
      ),
    [{}],
  );
}

export function VariantMatrixPreview({
  templateName,
  attributes,
}: {
  templateName: string;
  attributes: AttributeOption[];
}) {
  const matrix = buildMatrix(attributes);
  const t = (
    useTranslations as unknown as (
      ns: string,
    ) => (key: string, values?: Record<string, string | number>) => string
  )("Products");
  const skus = matrix.map((combination) => generateSku(templateName || "", combination));
  const duplicates = [...skus.filter((sku, index) => skus.indexOf(sku) !== index)];

  if (matrix.length === 0 || Object.keys(matrix[0] ?? {}).length === 0) {
    return <p className="text-sm text-muted-foreground">{t("noAttributesHint")}</p>;
  }

  return (
    <div className="flex flex-col gap-2">
      <p className="text-sm">{t("variantsCreated", { count: matrix.length })}</p>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("variantsTab")}</TableHead>
              <TableHead>SKU</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {matrix.map((combination, index) => (
              <TableRow key={JSON.stringify(combination)}>
                <TableCell>
                  <span className="text-sm text-muted-foreground">
                    {Object.entries(combination)
                      .map(([name, value]) => `${name}: ${value}`)
                      .join(" · ")}
                  </span>
                </TableCell>
                <TableCell>
                  <span
                    className={cn(
                      "font-mono text-xs",
                      duplicates.includes(skus[index]!) && "text-destructive",
                    )}
                  >
                    {skus[index]!}
                  </span>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      {duplicates.length > 0 ? (
        <p className="text-sm text-destructive" role="alert">
          {t("duplicateSkuDetail", { sku: duplicates[0] ?? "" })}
        </p>
      ) : null}
    </div>
  );
}
