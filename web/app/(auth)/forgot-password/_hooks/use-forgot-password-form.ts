import { zodResolver } from "@hookform/resolvers/zod";
import { isAxiosError } from "axios";
import { StatusCodes } from "http-status-codes";
import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { axiosInstance } from "@/lib/client/axios";
import { isServerError } from "@/lib/client/error";

export const forgotPasswordMethods = ["email", "phone"] as const;
export type ForgotPasswordMethod = (typeof forgotPasswordMethods)[number];

export function detectMethod(identifier: string): ForgotPasswordMethod {
  return identifier.includes("@") ? "email" : "phone";
}

export type ForgotPasswordFormMessages = {
  identifierRequired: string;
  invalidEmail: string;
  invalidPhone: string;
};

const forgotPasswordFormDefaultMessages: ForgotPasswordFormMessages = {
  identifierRequired: "Please enter your email or phone number.",
  invalidEmail: "Please enter a valid email address.",
  invalidPhone: "Please enter a valid phone number.",
};

export function useForgotPasswordFormSchema(
  messages: ForgotPasswordFormMessages = forgotPasswordFormDefaultMessages,
) {
  return z
    .object({
      identifier: z.string().min(1, { message: messages.identifierRequired }),
    })
    .superRefine((value, ctx) => {
      const trimmed = value.identifier.trim();
      if (trimmed.includes("@")) {
        if (!z.email().safeParse(trimmed).success) {
          ctx.addIssue({
            code: "custom",
            path: ["identifier"],
            message: messages.invalidEmail,
          });
        }
        return;
      }
      const digits = trimmed.replace(/[\s-]/g, "");
      if (!/^\+?\d{7,15}$/.test(digits)) {
        ctx.addIssue({
          code: "custom",
          path: ["identifier"],
          message: messages.invalidPhone,
        });
      }
    });
}

export type ForgotPasswordFormSchema = z.infer<ReturnType<typeof useForgotPasswordFormSchema>>;

const forgotPasswordFormDefaultValues: ForgotPasswordFormSchema = {
  identifier: "",
};

export function useForgotPasswordForm() {
  const [isPending, startTransition] = useTransition();
  const [isSubmitted, setIsSubmitted] = useState(false);
  const t = useTranslations("Auth");
  const tCommon = useTranslations("Common");
  const tx = t as unknown as (key: string) => string;
  const txCommon = tCommon as unknown as (
    key: string,
    values?: Record<string, string | number>,
  ) => string;
  const forgotPasswordFormSchema = useForgotPasswordFormSchema({
    identifierRequired: tx("identifierRequired"),
    invalidEmail: tx("invalidEmail"),
    invalidPhone: tx("invalidPhone"),
  });

  const form = useForm<ForgotPasswordFormSchema>({
    resolver: zodResolver(forgotPasswordFormSchema),
    defaultValues: forgotPasswordFormDefaultValues,
  });

  const handleSubmit = form.handleSubmit((data: ForgotPasswordFormSchema) => {
    startTransition(async () => {
      try {
        const method = detectMethod(data.identifier.trim());
        await axiosInstance.post("/auth/password-reset/request", {
          method,
          identifier: data.identifier.trim(),
        });
        setIsSubmitted(true);
      } catch (error: unknown) {
        if (isAxiosError(error) && error.response?.status === StatusCodes.TOO_MANY_REQUESTS) {
          const retryAfter = error.response.headers?.["retry-after"];
          const seconds = retryAfter ? parseInt(retryAfter as string, 10) : undefined;
          toast.warning(txCommon("rateLimited"), {
            description: seconds ? txCommon("retryInSeconds", { seconds }) : undefined,
          });
          return;
        }
        if (isServerError(error)) {
          toast.error(txCommon("serverError"));
          return;
        }
        toast.error(tx("sendResetFailed"));
      }
    });
  });

  return {
    form,
    handleSubmit,
    isPending,
    isSubmitted,
  };
}
