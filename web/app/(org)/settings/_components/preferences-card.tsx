import { Save, Trash2 } from "lucide-react";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Label } from "@/components/label";
import { Separator } from "@/components/separator";
import { Switch } from "@/components/switch";

function SwitchRow({
  name,
  label,
  description,
  disabled,
}: {
  name: string;
  label: string;
  description: string;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="space-y-0.5">
        <Label htmlFor={name}>{label}</Label>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <Switch id={name} defaultChecked disabled={disabled} />
    </div>
  );
}

export async function PreferencesCard({ orgId }: { orgId: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Preferences"}</CardTitle>
        <CardDescription>{"How your organization behaves by default."}</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <SwitchRow
          name={`notify-${orgId}`}
          label={"Weekly digest"}
          description={"Receive a summary of activity every Monday."}
          disabled
        />
        <Separator />
        <SwitchRow
          name={`autoExport-${orgId}`}
          label={"Auto-export reports"}
          description={"Export scheduled reports to your email automatically."}
          disabled
        />
        <Separator />
        <SwitchRow
          name={`notifyNewMember-${orgId}`}
          label={"New member alerts"}
          description={"Notify owners when someone joins or leaves."}
          disabled
        />
        <Separator />
        <SwitchRow
          name={`passwordless-${orgId}`}
          label={"Passwordless sign-in"}
          description={"Allow members to sign in without a password."}
          disabled
        />
      </CardContent>
      <CardContent className="flex items-center justify-between gap-2 border-t pt-4">
        <span className="text-xs text-muted-foreground">
          {"Changes apply to everyone in this organization."}
        </span>
        <div className="flex shrink-0 gap-2">
          <Button variant="outline" disabled>
            <Trash2 />
            <span>{"Danger zone"}</span>
          </Button>
          <Button disabled>
            <Save />
            <span>{"Save changes"}</span>
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
