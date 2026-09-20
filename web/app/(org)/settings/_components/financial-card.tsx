import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import { Field } from "@/components/field";
import { Label } from "@/components/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";

export async function FinancialCard({ orgId }: { orgId: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{"Financials"}</CardTitle>
        <CardDescription>
          {"Currency, tax year, and localization used across reports."}
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2">
        <Field>
          <Label htmlFor={`currency-${orgId}`}>{"Default currency"}</Label>
          <Select disabled>
            <SelectTrigger id={`currency-${orgId}`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="USD">USD — US Dollar</SelectItem>
              <SelectItem value="IDR">IDR — Indonesian Rupiah</SelectItem>
              <SelectItem value="SGD">SGD — Singapore Dollar</SelectItem>
              <SelectItem value="EUR">EUR — Euro</SelectItem>
            </SelectContent>
          </Select>
        </Field>
        <Field>
          <Label htmlFor={`taxYear-${orgId}`}>{"Tax year"}</Label>
          <Select disabled>
            <SelectTrigger id={`taxYear-${orgId}`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="jan">{`Starts in ${"Jan"}`}</SelectItem>
              <SelectItem value="apr">{`Starts in ${"Apr"}`}</SelectItem>
              <SelectItem value="jul">{`Starts in ${"Jul"}`}</SelectItem>
              <SelectItem value="oct">{`Starts in ${"Oct"}`}</SelectItem>
            </SelectContent>
          </Select>
        </Field>
        <Field>
          <Label htmlFor={`timezone-${orgId}`}>{"Timezone"}</Label>
          <Select disabled>
            <SelectTrigger id={`timezone-${orgId}`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="asia-jakarta">Asia/Jakarta (UTC+7)</SelectItem>
              <SelectItem value="asia-singapore">Asia/Singapore (UTC+8)</SelectItem>
              <SelectItem value="europe-berlin">Europe/Berlin (UTC+1)</SelectItem>
              <SelectItem value="america-new_york">America/New York (UTC-5)</SelectItem>
            </SelectContent>
          </Select>
        </Field>
        <Field>
          <Label htmlFor={`lang-${orgId}`}>{"Language"}</Label>
          <Select disabled>
            <SelectTrigger id={`lang-${orgId}`}>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="en">English</SelectItem>
              <SelectItem value="id">Bahasa Indonesia</SelectItem>
            </SelectContent>
          </Select>
        </Field>
      </CardContent>
    </Card>
  );
}
