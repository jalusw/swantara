"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
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

const COUNTRY_OPTIONS = [
  { code: "ID", name: "Indonesia" },
  { code: "SG", name: "Singapore" },
  { code: "US", name: "United States" },
  { code: "GB", name: "United Kingdom" },
  { code: "DE", name: "Germany" },
  { code: "MY", name: "Malaysia" },
  { code: "TH", name: "Thailand" },
  { code: "PH", name: "Philippines" },
  { code: "VN", name: "Vietnam" },
  { code: "AU", name: "Australia" },
  { code: "JP", name: "Japan" },
  { code: "KR", name: "South Korea" },
] as const;

function resolveStandardCode(countryCode: string): string {
  return countryCode.toUpperCase() === "ID" ? "PSAK_EMKM" : "IFRS";
}

function useCreateOrganizationSchema() {
  return z.object({
    name: z.string().min(1, "Company name is required"),
    countryCode: z.string().min(2, "Country is required"),
  });
}

export type CreateOrganizationDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function CreateOrganizationDialog({ open, onOpenChange }: CreateOrganizationDialogProps) {
  const router = useRouter();
  const queryClient = useQueryClient();
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
      toast.success("Organization created");
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
      toast.error("Failed to create organization. Please try again.");
    },
  });

  function handleSubmit(values: Values) {
    mutation.mutate(values);
  }

  return (
    <EntityFormDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Create organization"
      description="Create a new company to manage separately."
      form={form}
      onSubmit={handleSubmit}
      isPending={mutation.isPending}
      className="sm:max-w-lg"
    >
      <FormField name="name" label="Company name">
        {({ field, id }) => (
          <Input {...field} id={id} placeholder="Acme Inc." autoComplete="organization" />
        )}
      </FormField>
      <FormField name="countryCode" label="Country">
        {({ field, id }) => (
          <Select value={field.value} onValueChange={field.onChange}>
            <SelectTrigger id={id} aria-label="Country">
              <SelectValue placeholder="Select country" />
            </SelectTrigger>
            <SelectContent>
              {COUNTRY_OPTIONS.map((country) => (
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
