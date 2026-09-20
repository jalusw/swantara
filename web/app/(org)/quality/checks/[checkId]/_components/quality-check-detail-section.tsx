"use client";

import { useState } from "react";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { StateBadge } from "@/components/state-badge";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { QualityCheck } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import {
  canRecordResult,
  checkResultLabel,
  checkResultTone,
} from "../../../_components/quality-utils";
import { RecordResultDialog } from "../../_components/record-result-dialog";

export function QualityCheckDetail({ orgId, checkId }: { orgId: string; checkId: string }) {
  const [resultDialogOpen, setResultDialogOpen] = useState(false);

  const query = useOrgQuery<{ check: QualityCheck }>("qualityCheck", checkId, (organizationId) =>
    getSwantaraService().qualityChecks.get(organizationId, Number(checkId)),
  );

  const check = query.data?.check;
  const isLoading = query.isLoading;
  const error = query.isError ? query.error : null;

  if (isLoading) {
    return <div className="text-muted-foreground text-sm">{"Loading…"}</div>;
  }

  if (error || !check) {
    return (
      <div className="text-destructive text-sm">{error?.message ?? "Quality check not found."}</div>
    );
  }

  const tone = checkResultTone(check.result);

  return (
    <div className="flex flex-col gap-4 sm:gap-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-lg">{`Quality Check #${check.id}`}</h1>
          <StateBadge tone={tone} label={checkResultLabel(check.result)} />
        </div>
        {canRecordResult(check) ? (
          <Button size="sm" onClick={() => setResultDialogOpen(true)}>
            {"Record result"}
          </Button>
        ) : null}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{"Overview"}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">{"Item"}</p>
            <p className="text-sm">{check.itemId ? `Item #${check.itemId}` : "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Batch"}</p>
            <p className="text-sm">{check.batchId ? `Batch #${check.batchId}` : "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Shipment"}</p>
            <p className="text-sm">{check.shipmentId ? `#${check.shipmentId}` : "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"MO"}</p>
            <p className="text-sm">
              {check.productionOrderId ? `#${check.productionOrderId}` : "—"}
            </p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Measured value"}</p>
            <p className="text-sm">{check.measuredValue != null ? check.measuredValue : "—"}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">{"Checked at"}</p>
            <p className="text-sm">{check.checkedAt ? formatDate(check.checkedAt) : "—"}</p>
          </div>
        </CardContent>
      </Card>

      {resultDialogOpen && check ? (
        <RecordResultDialog
          open={resultDialogOpen}
          onOpenChange={setResultDialogOpen}
          orgId={orgId}
          check={check}
          onSave={() => {
            setResultDialogOpen(false);
            void query.refetch();
          }}
        />
      ) : null}
    </div>
  );
}
