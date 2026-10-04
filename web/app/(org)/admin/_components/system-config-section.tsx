"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { EyeOffIcon } from "lucide-react";
import { useTranslations } from "next-intl";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/dialog";
import { Form, FormField, SubmitButton } from "@/components/form";
import { Input } from "@/components/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/table";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { SystemConfig } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { logger } from "@/lib/utils/logger";

const SENSITIVE_PATTERNS = /password|secret|key|token|api_key|credential/i;

function isSensitive(key: string): boolean {
  return SENSITIVE_PATTERNS.test(key);
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function useConfigEditorSchema() {
  const t = useTranslations("Admin");
  return z.object({ value: z.string().min(1, t("validation_valueRequired")) });
}
type ConfigEditorValues = z.infer<ReturnType<typeof useConfigEditorSchema>>;

function ConfigEditorDialog({
  config,
  orgId,
  onSaved,
}: {
  config: SystemConfig;
  orgId: string;
  onSaved: () => void;
}) {
  const t = useTranslations("Admin");
  const tCommon = useTranslations("Common");
  const [open, setOpen] = useState(false);
  const sensitive = isSensitive(config.key);

  const schema = useConfigEditorSchema();

  const form = useForm<ConfigEditorValues>({
    resolver: zodResolver(schema),
    defaultValues: { value: formatValue(config.value) },
  });

  useEffect(() => {
    if (open) {
      form.reset({ value: formatValue(config.value) });
    }
  }, [open, config.value, form]);

  function handleSubmit(values: ConfigEditorValues) {
    let parsed: unknown = values.value;
    try {
      parsed = JSON.parse(values.value);
    } catch {
      parsed = values.value;
    }

    return getSwantaraService()
      .systemConfigs.update(Number(orgId), config.id, { value: parsed })
      .then(() => {
        toast.success(t("configUpdated"));
        setOpen(false);
        onSaved();
      })
      .catch((error) => {
        logger.error("Failed to update config", error);
        toast.error(t("saveFailed"));
      });
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button variant="outline" size="sm" />}>{t("edit")}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("editConfig")}</DialogTitle>
          <DialogDescription className="font-mono text-xs">{config.key}</DialogDescription>
        </DialogHeader>
        <Form form={form} onSubmit={handleSubmit}>
          <FormField name="value" label={t("valueLabel")}>
            {({ field }) => (
              <Input
                {...field}
                type={sensitive ? "password" : "text"}
                placeholder={t("enterValue")}
                autoComplete="off"
              />
            )}
          </FormField>
          {sensitive ? (
            <p className="flex items-center gap-1.5 text-sm text-muted-foreground">
              <EyeOffIcon className="size-3.5" aria-hidden />
              {t("sensitiveHint")}
            </p>
          ) : null}
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              {tCommon("cancel")}
            </Button>
            <SubmitButton>{tCommon("save")}</SubmitButton>
          </DialogFooter>
        </Form>
      </DialogContent>
    </Dialog>
  );
}

export function SystemConfigSection({ orgId }: { orgId: string }) {
  const t = useTranslations("Admin");
  const tCommon = useTranslations("Common");
  const [query, setQuery] = useState("");

  const configQuery = useOrgListQuery<{ systemConfigs: SystemConfig[] }, Record<string, never>>(
    "systemConfigs",
    (organizationId) => getSwantaraService().systemConfigs.list(organizationId),
  );

  const items = configQuery.data?.systemConfigs ?? [];
  const normalized = query.trim().toLowerCase();
  const filtered = items.filter(
    (item) => normalized.length === 0 || item.key.toLowerCase().includes(normalized),
  );

  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <CardTitle>{t("systemConfigTitle")}</CardTitle>
          <CardDescription>{t("systemConfigDescription")}</CardDescription>
        </div>
        <Input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder={t("searchConfigs")}
          className="max-w-64"
          aria-label={t("searchConfigs")}
        />
      </CardHeader>
      <CardContent>
        <Table aria-label={t("systemConfigTitle")}>
          <TableHeader>
            <TableRow>
              <TableHead>{t("colKey")}</TableHead>
              <TableHead>{t("colValue")}</TableHead>
              <TableHead className="w-24 text-right">{t("colEdit")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {configQuery.isLoading ? (
              <TableRow>
                <TableCell colSpan={3} className="py-8 text-center text-muted-foreground">
                  {tCommon("loading")}
                </TableCell>
              </TableRow>
            ) : filtered.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} className="py-8 text-center text-muted-foreground">
                  {t("noConfig")}
                </TableCell>
              </TableRow>
            ) : (
              filtered.map((item) => (
                <TableRow key={item.id}>
                  <TableCell className="font-mono text-xs">{item.key}</TableCell>
                  <TableCell className="font-mono text-xs">
                    {isSensitive(item.key) ? "••••••••" : formatValue(item.value)}
                  </TableCell>
                  <TableCell className="text-right">
                    <ConfigEditorDialog
                      config={item}
                      orgId={orgId}
                      onSaved={() => void configQuery.refetch()}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}
