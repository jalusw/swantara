"use client";

import { toast } from "sonner";
import { moduleKey } from "@/app/(org)/_components/module-collapsible";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Label } from "@/components/label";
import { Switch } from "@/components/switch";
import { useActiveModuleIds, useOrgModules, useUpdateOrgModule } from "@/lib/hooks/use-org-modules";
import { usePermissions } from "@/lib/hooks/use-permissions";

export function ModulesCard() {
  const { has } = usePermissions();
  const { data, isPending } = useOrgModules();
  const activeModules = useActiveModuleIds();
  const updateModule = useUpdateOrgModule();
  const canManage = has("organization.update");

  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Modules"}</CardTitle>
        <CardDescription>
          {
            "Choose which parts of your business show in the sidebar. New organizations start with everything on."
          }
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {isPending ? (
          <p className="text-sm text-muted-foreground">{"Loading modules..."}</p>
        ) : (
          (data?.modules ?? []).map((module) => {
            const label = moduleKey[module.moduleId as keyof typeof moduleKey] ?? module.moduleId;
            return (
              <div key={module.moduleId} className="flex items-center justify-between gap-4">
                <Label htmlFor={`module-${module.moduleId}`}>{label}</Label>
                <Switch
                  id={`module-${module.moduleId}`}
                  checked={activeModules?.has(module.moduleId) ?? module.active}
                  disabled={!canManage || updateModule.isPending}
                  onCheckedChange={(active) => {
                    updateModule.mutate(
                      { moduleId: module.moduleId, active },
                      {
                        onError: () => toast.error("Couldn't update the module. Please try again."),
                      },
                    );
                  }}
                />
              </div>
            );
          })
        )}
      </CardContent>
    </Card>
  );
}
