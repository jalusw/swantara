"use client";

import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { Journal, JournalEntry } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, getLocalDateString } from "@/lib/utils";
import { canReverse, moveStateTone } from "../../_components/journal-entry-utils";

type TFn = (key: string, values?: Record<string, string | number>) => string;

export function JournalEntryDetailSection({ orgId, entryId }: { orgId: string; entryId: string }) {
  const t = (useTranslations as unknown as (ns: string) => TFn)("Accounting");
  const tCommon = useTranslations("Common");
  const movementQuery = useOrgQuery<{ movement: JournalEntry }>(
    "journalEntries",
    entryId,
    (organizationId) => getSwantaraService().journalEntries.get(organizationId, Number(entryId)),
  );

  const journalsQuery = useOrgListQuery<{ journals: Journal[] }, Record<string, never>>(
    "journals",
    (organizationId) => getSwantaraService().journals.list(organizationId),
  );

  const movement = movementQuery.data?.movement;
  const journalName = (journalsQuery.data?.journals ?? []).find(
    (j) => j.id === movement?.journalId,
  )?.name;

  function handleReverse() {
    if (!movement) return;
    void getSwantaraService()
      .journalEntries.reverse(Number(orgId), movement.id, {
        date: getLocalDateString(),
        journalId: movement.journalId,
        ref: "",
        description: "",
      })
      .then(() => {
        toast.success(t("toastEntryReversed"));
        void movementQuery.refetch();
      });
  }

  if (movementQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (movementQuery.isError) {
    return <div className="text-sm text-destructive">{t("entryLoadFailed")}</div>;
  }

  if (!movement) {
    return <p className="text-sm text-muted-foreground">{t("entryNotFound")}</p>;
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold">{movement.name ?? `JE-${movement.id}`}</h1>
          <p className="text-sm text-muted-foreground">
            {journalName ?? `#${movement.journalId}`} &middot; {formatDate(movement.date)}
          </p>
        </div>
        <div className="flex gap-2">
          <Badge variant="outline" className={moveStateTone(movement.state)}>
            {(t as unknown as (k: string) => string)(`entryState_${movement.state}`)}
          </Badge>
          {canReverse(movement.state) ? (
            <Button size="sm" variant="outline" onClick={handleReverse}>
              {t("reverseEntry")}
            </Button>
          ) : null}
        </div>
      </div>

      {movement.ref ? (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">{t("colReference")}</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-sm">{movement.ref}</p>
          </CardContent>
        </Card>
      ) : null}
    </div>
  );
}
