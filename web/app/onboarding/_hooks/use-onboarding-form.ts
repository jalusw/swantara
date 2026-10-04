import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { useCallback } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { setActiveOrg } from "@/lib/server/active-org-actions";
import {
  getSwantaraService,
  SwantaraError,
  SwantaraUnprocessableError,
} from "@/lib/services/swantara";

import { resolveStandardCode } from "../_components/onboarding-utils";
import {
  type OnboardingFormStore,
  type OnboardingStep,
  useOnboardingFormStore,
} from "./onboarding-form-store";

export type OnboardingFormMessages = {
  nameRequired: string;
  countryRequired: string;
};

const onboardingFormDefaultMessages: OnboardingFormMessages = {
  nameRequired: "Company name is required",
  countryRequired: "Country is required",
};

export function useOnboardingFormSchema(
  messages: OnboardingFormMessages = onboardingFormDefaultMessages,
) {
  return z.object({
    name: z.string().min(1, { message: messages.nameRequired }),
    countryCode: z.string().min(2, { message: messages.countryRequired }),
  });
}

export type OnboardingFormSchema = z.infer<ReturnType<typeof useOnboardingFormSchema>>;

const onboardingFormDefaultValues = {
  name: "",
  countryCode: "",
} as unknown as OnboardingFormSchema;

const QUICK_CREATE_FIELD_MAP: Record<string, keyof OnboardingFormSchema> = {
  name: "name",
  country_code: "countryCode",
};

const stepFields: Record<OnboardingStep, (keyof OnboardingFormSchema)[]> = {
  companyInfo: ["name", "countryCode"],
};

const STEP_ORDER: OnboardingStep[] = ["companyInfo"];

export type UseOnboardingFormParams = {
  defaultValues?: OnboardingFormSchema;
};

export function useOnboardingForm({
  defaultValues = onboardingFormDefaultValues,
}: UseOnboardingFormParams) {
  const t = useTranslations("Onboarding");
  const tCommon = useTranslations("Common");
  const tx = t as unknown as (key: string) => string;
  const txCommon = tCommon as unknown as (key: string) => string;
  const onboardingFormSchema = useOnboardingFormSchema({
    nameRequired: tx("companyNameRequired"),
    countryRequired: tx("countryRequired"),
  });
  const router = useRouter();

  const form = useForm<OnboardingFormSchema>({
    resolver: zodResolver(onboardingFormSchema),
    defaultValues,
  });

  const handleItemChange = useCallback(
    async (next: string) => {
      const { step, setStep } = useOnboardingFormStore.getState();
      const nextStep = next as OnboardingStep;
      if (nextStep === step) return;
      if (STEP_ORDER.indexOf(nextStep) < STEP_ORDER.indexOf(step)) {
        form.clearErrors();
        setStep(nextStep);
        return;
      }
      const isValid = await form.trigger(stepFields[step]);
      if (!isValid) return;
      form.clearErrors();
      setStep(nextStep);
    },
    [form],
  );

  const quickCreateMutation = useMutation({
    mutationFn: (data: OnboardingFormSchema) => {
      const service = getSwantaraService();
      return service.organizations.quickCreate({
        name: data.name,
        countryCode: data.countryCode,
        standardCode: resolveStandardCode(data.countryCode),
      });
    },
    onSuccess: async ({ organization }) => {
      await setActiveOrg(organization.id);
      router.push("/dashboard");
      router.refresh();
    },
    onError: (error: unknown) => {
      if (error instanceof SwantaraUnprocessableError) {
        for (const fieldError of error.fieldErrors ?? []) {
          const target = QUICK_CREATE_FIELD_MAP[fieldError.field];
          if (target) {
            form.setError(target, { message: fieldError.message });
          }
        }
        toast.error(error.message);
        return;
      }
      if (error instanceof SwantaraError && error.status >= 500) {
        toast.error(txCommon("serverError"));
        return;
      }
      if (error instanceof SwantaraError && error.status === 0) {
        toast.error(txCommon("networkError"));
        return;
      }
      if (error instanceof SwantaraError) {
        toast.error(error.message || tx("createFailed"));
      }
      form.setError("name", {
        message: tx("createFailed"),
      });
    },
    onSettled: () => {
      useOnboardingFormStore.getState().setPending(false);
    },
  });

  const handleSubmit = form.handleSubmit((data: OnboardingFormSchema) => {
    useOnboardingFormStore.getState().setPending(true);
    quickCreateMutation.mutate(data);
  });

  return {
    form,
    handleItemChange,
    handleSubmit,
  };
}

export type { OnboardingFormStore, OnboardingStep };
export { useOnboardingFormStore };
