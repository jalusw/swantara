"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { EntityFormDialog } from "@/components/entity-form-dialog";
import { FormField } from "@/components/form";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { ME_ORGANIZATIONS_QUERY_KEY } from "@/lib/queries/me";
import { setActiveOrg } from "@/lib/server/active-org-actions";
import { getSwantaraService, SwantaraUnprocessableError } from "@/lib/services/swantara";

const COUNTRY_CODES = [
  "ID",
  "SG",
  "US",
  "GB",
  "DE",
  "MY",
  "TH",
  "PH",
  "VN",
  "AU",
  "JP",
  "KR",
] as const;

function useCountryOptions(): Array<{ code: string; name: string }> {
  const locale = useLocale();
  return COUNTRY_CODES.map((code) => {
    let name: string | undefined;
    try {
      name = new Intl.DisplayNames([locale], { type: "region" }).of(code);
    } catch {
      name = undefined;
    }
    return { code, name: name ?? code };
  });
}

function resolveStandardCode(countryCode: string): string {
  return countryCode.toUpperCase() === "ID" ? "PSAK_EMKM" : "IFRS";
}

function useCreateOrganizationSchema() {
  const tOrg = useTranslations("Organization" as unknown as "Common");
  const tx = tOrg as unknown as (key: string) => string;
  return z.object({
    name: z.string().min(1, tx("nameRequired")),
    countryCode: z.string().min(2, tx("countryRequired")),
  });
}

export type CreateOrganizationDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function CreateOrganizationDialog({ open, onOpenChange }: CreateOrganizationDialogProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const tOrg = useTranslations("Organization" as unknown as "Common");
  const tx = tOrg as unknown as (key: string) => string;
  const tCommon = useTranslations("Common");
  const countryOptions = useCountryOptions();
  const schema = useCreateOrganizationSchema();
  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", countryCode: "" },
  });

  const mutation = useMutation({
    mutationFn: (values: Values) =>
      getSwantaraService().organizations.quickCreate({
        name: values.name,
        countryCode: values.countryCode,
        standardCode: resolveStandardCode(values.countryCode),
      }),
    onSuccess: async ({ organization }) => {
      await setActiveOrg(organization.id);
      void queryClient.invalidateQueries({ queryKey: ME_ORGANIZATIONS_QUERY_KEY });
      form.reset();
      onOpenChange(false);
      toast.success(tx("createSuccess"));
      router.push("/dashboard");
      router.refresh();
    },
    onError: (error: unknown) => {
      if (error instanceof SwantaraUnprocessableError) {
        for (const fieldError of error.fieldErrors ?? []) {
          if (fieldError.field === "name") form.setError("name", { message: fieldError.message });
          if (fieldError.field === "country_code")
            form.setError("countryCode", { message: fieldError.message });
        }
        toast.error(error.message);
        return;
      }
      toast.error(tx("createFailed"));
    },
  });

  function handleSubmit(values: Values) {
    mutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title={tx("dialogTitle")}
      description={tx("dialogDescription")}
      form={form}
      onSubmit={handleSubmit}
      isPending={mutation.isPending}
      submitLabel={tCommon("save")}
      cancelLabel={tCommon("cancel")}
      className="sm:max-w-lg"
    >
      <FormField name="name" label={tx("nameLabel")}>
        {({ field, id }) => (
          <Input {...field} id={id} placeholder="Acme Inc." autoComplete="organization" />
        )}
      </FormField>
      <FormField name="countryCode" label={tx("countryLabel")}>
        {({ field, id }) => (
          <Select value={field.value} onValueChange={field.onChange}>
            <SelectTrigger id={id} aria-label={tx("countryLabel")}>
              <SelectValue placeholder={tx("countryPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {countryOptions.map((country) => (
                <SelectItem key={country.code} value={country.code}>
                  {country.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
      </FormField>
    </EntityFormDialog>
  );
}
