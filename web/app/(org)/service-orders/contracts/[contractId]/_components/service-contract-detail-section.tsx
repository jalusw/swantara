"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { RecordLayout } from "@/components/record-layout";
import { useOrgQuery } from "@/lib/hooks/use-org-query";
import type { ServiceContract } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { formatDate } from "@/lib/utils";
import {
  canActivate,
  canCancel,
  serviceContractStateTone,
} from "../../_components/service-contract-utils";

export function ServiceContractDetail({
  orgId,
  contractId,
}: {
  orgId: string;
  contractId: string;
}) {
  const t = useTranslations("Service");
  const tCommon = useTranslations("Common");
  const contractState = (state: string) =>
    (t as unknown as (k: string) => string)(`contractState_${state}`);
  const queryClient = useQueryClient();

  const query = useOrgQuery<{ serviceContract: ServiceContract }>(
    "serviceContract",
    contractId,
    (organizationId) =>
      getSwantaraService().serviceContracts.get(organizationId, Number(contractId)),
  );

  const contract = query.data?.serviceContract;

  const actionMutation = useMutation({
    mutationFn: (action: "activate" | "cancel") =>
      getSwantaraService().serviceContracts[action](Number(orgId), Number(contractId)),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["serviceContract"] });
      void queryClient.invalidateQueries({ queryKey: ["serviceContracts"] });
      toast.success(t("saved"));
    },
    onError: () => {
      toast.error(t("actionFailed"));
    },
  });

  function handleAction(action: "activate" | "cancel") {
    if (actionMutation.isPending) return;
    actionMutation.mutate(action);
  }

  if (query.isLoading) {
    return <p className="text-sm text-muted-foreground">{tCommon("loading")}</p>;
  }

  if (!contract) {
    return <p className="text-sm text-muted-foreground">{t("contractNotFound")}</p>;
  }

  const tone = serviceContractStateTone(contract.state);

  const stateActions = (
    <div className="flex items-center gap-2">
      {canActivate(contract.state) ? (
        <Button
          size="sm"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("activate")}
        >
          {t("action_activate")}
        </Button>
      ) : null}
      {canCancel(contract.state) ? (
        <Button
          size="sm"
          variant="destructive"
          disabled={actionMutation.isPending}
          onClick={() => handleAction("cancel")}
        >
          {tCommon("cancel")}
        </Button>
      ) : null}
    </div>
  );

  return (
    <RecordLayout
      breadcrumbItems={[
        { label: t("contractsTitle"), href: "/service-orders/contracts" },
        { label: contract.name },
      ]}
      title={contract.name}
      status={
        <Badge
          variant="outline"
          className={
            tone === "success"
              ? "border-success text-success"
              : tone === "warning"
                ? "border-warning text-warning"
                : tone === "danger"
                  ? "border-destructive text-destructive"
                  : tone === "info"
                    ? "border-info text-info"
                    : ""
          }
        >
          {contractState(contract.state)}
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
                  <CardTitle>{t("coverage")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{contract.coverage || "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("slaResponseHours")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm tabular-nums">{contract.slaResponseHours ?? "—"}</p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>{t("customer")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm">{contract.contactId ? `#${contract.contactId}` : "—"}</p>
                </CardContent>
              </Card>
              <Card className="lg:col-span-3">
                <CardHeader>
                  <CardTitle>{t("detailsTitle")}</CardTitle>
                </CardHeader>
                <CardContent>
                  <dl className="grid gap-4 sm:grid-cols-2">
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("equipment")}</dt>
                      <dd className="text-sm">
                        {contract.equipmentId ? `#${contract.equipmentId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("subscription")}</dt>
                      <dd className="text-sm">
                        {contract.subscriptionId ? `#${contract.subscriptionId}` : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("startDate")}</dt>
                      <dd className="text-sm">
                        {contract.dateStart ? formatDate(String(contract.dateStart)) : "—"}
                      </dd>
                    </div>
                    <div>
                      <dt className="text-sm text-muted-foreground">{t("endDate")}</dt>
                      <dd className="text-sm">
                        {contract.dateEnd ? formatDate(String(contract.dateEnd)) : "—"}
                      </dd>
                    </div>
                  </dl>
                </CardContent>
              </Card>
            </div>
          ),
        },
      ]}
    />
  );
}
