"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useId, useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { Input } from "@/components/input";
import { RecordLayout } from "@/components/record-layout";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { DEFAULT_CURRENCY } from "@/lib/constants/currency";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { AssetDepreciationLine, FixedAsset } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate, formatMoney, getLocalDateString } from "@/lib/utils";
import {
  canDispose,
  canGenerateSchedule,
  canPostDepreciation,
  fixedAssetStateTone,
  nbv,
  totalDepreciation,
} from "../../_components/fixed-asset-utils";

export function FixedAssetDetail({ orgId, assetId }: { orgId: string; assetId: string }) {
  const t = useTranslations("FixedAssets");
  const queryClient = useQueryClient();
  const typeId = useId();
  const proceedsId = useId();
  const dateId = useId();
  const [disposeOpen, setDisposeOpen] = useState(false);
  const [disposeState, setDisposeState] = useState<"disposed" | "sold">("sold");
  const [proceedsAmount, setProceedsAmount] = useState("");
  const [postDate, setPostDate] = useState(() => getLocalDateString());

  const assetQuery = useOrgQuery<{ fixedAsset: FixedAsset; lines?: AssetDepreciationLine[] }>(
    "fixedAsset",
    assetId,
    (organizationId) => getSwantaraService().fixedAssets.get(organizationId, Number(assetId)),
  );

  const asset = assetQuery.data?.fixedAsset;
  const lines = assetQuery.data?.lines ?? [];

  const generateScheduleMutation = useMutation({
    mutationFn: () => getSwantaraService().fixedAssets.schedule(Number(orgId), Number(assetId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["fixedAsset", Number(orgId), Number(assetId)],
      });
      void queryClient.invalidateQueries({ queryKey: ["fixedAssets"] });
      toast.success(t("toastScheduleGenerated"));
    },
    onError: () => {
      toast.error(t("toastActionFailed"));
    },
  });

  const postDepreciationMutation = useMutation({
    mutationFn: () =>
      getSwantaraService().fixedAssets.postDepreciation(Number(orgId), Number(assetId), {
        journalId: 1,
        date: postDate,
      }),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ["fixedAsset", Number(orgId), Number(assetId)],
      });
      void queryClient.invalidateQueries({ queryKey: ["fixedAssets"] });
      toast.success(t("toastDepreciationPosted"));
    },
    onError: () => {
      toast.error(t("toastActionFailed"));
    },
  });

  const disposeMutation = useMutation({
    mutationFn: () =>
      getSwantaraService().fixedAssets.dispose(Number(orgId), Number(assetId), {
        journalId: 1,
        date: postDate,
        state: disposeState,
        proceedsAmount: proceedsAmount ? Number(proceedsAmount) : 0,
        proceedsAccountId: null,
      }),
    onSuccess: () => {
      setDisposeOpen(false);
      void queryClient.invalidateQueries({
        queryKey: ["fixedAsset", Number(orgId), Number(assetId)],
      });
      void queryClient.invalidateQueries({ queryKey: ["fixedAssets"] });
      toast.success(t("toastAssetDisposed"));
    },
    onError: () => {
      toast.error(t("toastActionFailed"));
    },
  });

  if (assetQuery.isLoading) {
    return <p className="text-sm text-muted-foreground">{t("loading")}</p>;
  }

  if (!asset) {
    return <p className="text-sm text-muted-foreground">{t("assetNotFound")}</p>;
  }

  const tone = fixedAssetStateTone(asset.state);
  const currentNbv = nbv(asset.purchaseValue, lines);
  const posted = totalDepreciation(lines);

  function handleGenerateSchedule() {
    generateScheduleMutation.mutate();
  }

  function handlePostDepreciation() {
    postDepreciationMutation.mutate();
  }

  function handleDispose() {
    disposeMutation.mutate();
  }

  const nextUnposted = lines.find((l) => !l.posted);

  const stateActions = (
    <div className="flex items-center gap-2">
      {canGenerateSchedule(asset.state) && lines.length === 0 ? (
        <Button size="sm" onClick={handleGenerateSchedule}>
          {t("generateSchedule")}
        </Button>
      ) : null}
      {canPostDepreciation(asset.state, lines) ? (
        <Button size="sm" onClick={handlePostDepreciation}>
          {t("postDepreciation")}
        </Button>
      ) : null}
      {canDispose(asset.state) ? (
        <Button size="sm" variant="outline" onClick={() => setDisposeOpen(true)}>
          {t("disposeSell")}
        </Button>
      ) : null}
    </div>
  );

  return (
    <>
      <RecordLayout
        breadcrumbItems={[{ label: t("title"), href: "/fixed-assets" }, { label: asset.name }]}
        title={asset.name}
        status={
          <Badge
            variant="outline"
            className={
              tone === "success"
                ? "border-success text-success"
                : tone === "warning"
                  ? "border-warning text-warning"
                  : tone === "info"
                    ? "border-info text-info"
                    : tone === "danger"
                      ? "border-destructive text-destructive"
                      : ""
            }
          >
            {(t as unknown as (k: string) => string)(`assetState_${asset.state}`)}
          </Badge>
        }
        actions={stateActions}
        tabs={[
          {
            id: "overview",
            label: t("tabOverview"),
            content: (
              <div className="grid gap-4 lg:grid-cols-3">
                <Card>
                  <CardHeader>
                    <CardTitle>{t("fieldPurchaseValue")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-bold tabular-nums">
                      {formatMoney(asset.purchaseValue, { currency: DEFAULT_CURRENCY })}
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle>{t("netBookValue")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-bold tabular-nums">
                      {formatMoney(currentNbv, { currency: DEFAULT_CURRENCY })}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {t("accumulatedDepreciation")}:{" "}
                      {formatMoney(posted, { currency: DEFAULT_CURRENCY })}
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle>{t("fieldSalvageValue")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-bold tabular-nums">
                      {formatMoney(asset.salvageValue, { currency: DEFAULT_CURRENCY })}
                    </p>
                  </CardContent>
                </Card>
                <Card className="lg:col-span-3">
                  <CardHeader>
                    <CardTitle>{t("assetInformation")}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <dl className="grid gap-4 sm:grid-cols-2">
                      <div>
                        <dt className="text-muted-foreground text-sm">{t("fieldCategory")}</dt>
                        <dd className="text-sm">#{asset.categoryId}</dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground text-sm">
                          {t("fieldAcquisitionDate")}
                        </dt>
                        <dd className="text-sm">
                          {asset.acquisitionDate ? formatDate(String(asset.acquisitionDate)) : "—"}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground text-sm">{t("fieldInServiceDate")}</dt>
                        <dd className="text-sm">
                          {asset.inServiceDate ? formatDate(String(asset.inServiceDate)) : "—"}
                        </dd>
                      </div>
                      <div>
                        <dt className="text-muted-foreground text-sm">{t("colStatus")}</dt>
                        <dd className="text-sm">
                          {(t as unknown as (k: string) => string)(`assetState_${asset.state}`)}
                        </dd>
                      </div>
                    </dl>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            id: "schedule",
            label: t("tabSchedule"),
            content: (
              <div className="space-y-4">
                {nextUnposted ? (
                  <div className="flex items-center gap-4 rounded-lg border p-4">
                    <span className="text-sm text-muted-foreground">{t("nextPeriodToPost")}</span>
                    <span className="text-sm tabular-nums">
                      {formatDate(String(nextUnposted.depreciationDate))} –{" "}
                      {formatMoney(nextUnposted.amount, { currency: DEFAULT_CURRENCY })}
                    </span>
                  </div>
                ) : null}
                {lines.length === 0 ? (
                  <p className="text-sm text-muted-foreground">{t("scheduleEmpty")}</p>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead>
                        <tr className="border-b text-muted-foreground">
                          <th className="pb-2 text-left">{"#"}</th>
                          <th className="pb-2 text-left">{t("colDate")}</th>
                          <th className="pb-2 text-right">{t("colAmount")}</th>
                          <th className="pb-2 text-right">{t("colAccumulated")}</th>
                          <th className="pb-2 text-right">{t("colRemaining")}</th>
                          <th className="pb-2 text-center">{t("colPosted")}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {lines.map((line) => (
                          <tr key={line.id} className="border-b">
                            <td className="py-2 tabular-nums">{line.sequence}</td>
                            <td className="py-2">{formatDate(String(line.depreciationDate))}</td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(line.amount, { currency: DEFAULT_CURRENCY })}
                            </td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(line.accumulated, { currency: DEFAULT_CURRENCY })}
                            </td>
                            <td className="py-2 text-right tabular-nums">
                              {formatMoney(line.remainingValue, { currency: DEFAULT_CURRENCY })}
                            </td>
                            <td className="py-2 text-center">
                              {line.posted ? (
                                <Badge variant="default">{t("posted")}</Badge>
                              ) : (
                                <Badge variant="outline">{t("pending")}</Badge>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            ),
          },
        ]}
      />
      <Dialog open={disposeOpen} onOpenChange={setDisposeOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("disposeDialogTitle")}</DialogTitle>
            <DialogDescription>{t("disposeDialogDescription")}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-2">
              <label htmlFor={typeId} className="text-sm">
                {t("fieldType")}
              </label>
              <Select
                value={disposeState}
                onValueChange={(v) => setDisposeState(v as "disposed" | "sold")}
              >
                <SelectTrigger id={typeId} aria-label={t("fieldType")}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="sold">
                    {(t as unknown as (k: string) => string)("assetState_sold")}
                  </SelectItem>
                  <SelectItem value="disposed">
                    {(t as unknown as (k: string) => string)("assetState_disposed")}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <label htmlFor={proceedsId} className="text-sm">
                {t("fieldProceedsAmount")}
              </label>
              <Input
                id={proceedsId}
                type="number"
                min="0"
                step="0.01"
                value={proceedsAmount}
                onChange={(e) => setProceedsAmount(e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-2">
              <label htmlFor={dateId} className="text-sm">
                {t("fieldDate")}
              </label>
              <Input
                id={dateId}
                type="date"
                value={postDate}
                onChange={(e) => setPostDate(e.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDisposeOpen(false)}>
              {t("cancel")}
            </Button>
            <Button onClick={handleDispose}>{t("confirm")}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
