import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Field } from "@/components/field";
import { Input } from "@/components/input";
import { Label } from "@/components/label";
import { PhoneInput } from "@/components/phone-input";
import { Textarea } from "@/components/textarea";

export async function ProfileCard({ orgId }: { orgId: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Profile"}</CardTitle>
        <CardDescription>{"Basic information about your organization."}</CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <Field className="sm:col-span-2">
          <Label htmlFor={`name-${orgId}`}>{"Organization name"}</Label>
          <Input id={`name-${orgId}`} disabled />
        </Field>
        <Field>
          <Label htmlFor={`legal-${orgId}`}>{"Legal name"}</Label>
          <Input id={`legal-${orgId}`} disabled />
        </Field>
        <Field>
          <Label htmlFor={`taxId-${orgId}`}>{"Tax ID"}</Label>
          <Input id={`taxId-${orgId}`} disabled />
        </Field>
        <Field>
          <Label htmlFor={`email-${orgId}`}>{"Billing email"}</Label>
          <Input id={`email-${orgId}`} type="email" disabled />
        </Field>
        <Field>
          <Label htmlFor={`phone-${orgId}`}>{"Phone"}</Label>
          <PhoneInput id={`phone-${orgId}`} disabled />
        </Field>
        <Field className="sm:col-span-2">
          <Label htmlFor={`address-${orgId}`}>{"Address"}</Label>
          <Textarea id={`address-${orgId}`} rows={2} disabled />
        </Field>
      </CardContent>
    </Card>
  );
}
