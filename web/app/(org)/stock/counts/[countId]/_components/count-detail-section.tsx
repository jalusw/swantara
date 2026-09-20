"use client";

import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { DetailPageSkeleton } from "@/components/detail-page-skeleton";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { RecordLayout } from "@/components/record-layout";
import { useOrgListQuery, useOrgQuery } from "@/lib/hooks/use-org-query";
import type { StockCount, StockCountLine } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatNumber } from "@/lib/utils";

export function CountDetail({ orgId, countId }: { orgId: string; countId: string }) {
  const [postingDialogOpen, setPostingDialogOpen] = useState(false);
  const [posting, setPosting] = useState(false);

  const countQuery = useOrgQuery<{ count: StockCount }>(
    "inventoryCount",
    countId,
    (organizationId) =>
      getSwantaraService().inventory.inventoryCount(organizationId, Number(countId)),
  );

  const linesQuery = useOrgListQuery<{ lines: StockCountLine[] }, Record<string, never>>(
    `inventoryCountLines-${countId}`,
    (organizationId) =>
      getSwantaraService().inventory.inventoryCountLines(organizationId, Number(countId)),
  );

  const count = countQuery.data?.count;
  const lines = linesQuery.data?.lines ?? [];

  if (countQuery.isLoading || linesQuery.isLoading) {
    return <DetailPageSkeleton />;
  }

  if (!count) {
    return <p className="text-sm text-muted-foreground">{"Count not found."}</p>;
  }

  const isDraft = count.state === "draft";
  const hasDiff = lines.some((line) => line.diffQty !== 0);
  const totalDiff = lines.reduce((sum, line) => sum + line.diffQty, 0);

  function handlePost() {
    if (!count) return;
    setPosting(true);
    void getSwantaraService()
      .inventory.postStockCount(Number(orgId), count.id, {
        journalId: 1,
        gainLossAccountId: 1,
        date: new Date().toISOString(),
      })
      .then(() => {
        setPostingDialogOpen(false);
        void countQuery.refetch();
        void linesQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."))
      .finally(() => setPosting(false));
  }

  return (
    <>
      <RecordLayout
        breadcrumbItems={[
          { label: "All counts", href: "/stock/counts" },
          { label: count.name ?? `IC-${count.id}` },
        ]}
        title={count.name ?? `IC-${count.id}`}
        status={
          <Badge variant={count.state === "posted" ? "default" : "secondary"}>{count.state}</Badge>
        }
        tabs={[
          {
            id: "lines",
            label: "Count lines",
            content: (
              <div className="flex flex-col gap-4">
                {isDraft ? (
                  <div className="flex items-center gap-2">
                    <Button
                      size="sm"
                      onClick={() => setPostingDialogOpen(true)}
                      disabled={!hasDiff}
                    >
                      {"Post count"}
                    </Button>
                    {hasDiff ? (
                      <span className="text-sm text-muted-foreground">
                        {`${lines.filter((l) => l.diffQty !== 0).length} lines differ (total ${formatNumber(totalDiff)}).`}
                      </span>
                    ) : null}
                  </div>
                ) : null}
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">{"Difference preview"}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    {lines.length === 0 ? (
                      <p className="text-sm text-muted-foreground">{"No count lines found."}</p>
                    ) : (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b text-left text-muted-foreground">
                              <th className="pb-2 ">{"Item"}</th>
                              <th className="pb-2 ">{"Theoretical"}</th>
                              <th className="pb-2 ">{"Counted"}</th>
                              <th className="pb-2 ">{"Diff"}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {lines.map((line) => (
                              <tr key={line.id} className="border-b last:border-0">
                                <td className="py-2">{line.itemId}</td>
                                <td className="py-2 tabular-nums">
                                  {formatNumber(line.theoreticalQty)}
                                </td>
                                <td className="py-2 tabular-nums">
                                  {formatNumber(line.countedQty)}
                                </td>
                                <td className="py-2 tabular-nums">
                                  <span
                                    className={
                                      line.diffQty > 0
                                        ? "text-success"
                                        : line.diffQty < 0
                                          ? "text-destructive"
                                          : ""
                                    }
                                  >
                                    {line.diffQty > 0 ? "+" : ""}
                                    {formatNumber(line.diffQty)}
                                  </span>
                                </td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </CardContent>
                </Card>
              </div>
            ),
          },
        ]}
      />
      <Dialog open={postingDialogOpen} onOpenChange={setPostingDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{"Post count"}</DialogTitle>
            <DialogDescription>
              {
                "Are you sure you want to post this count? This will create journal entries for any differences."
              }
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPostingDialogOpen(false)}>
              {"Cancel"}
            </Button>
            <Button onClick={handlePost} disabled={posting}>
              {posting ? "Posting…" : "Post count"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
