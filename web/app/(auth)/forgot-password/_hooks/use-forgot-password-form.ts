import { zodResolver } from "@hookform/resolvers/zod";
import { isAxiosError } from "axios";
import { StatusCodes } from "http-status-codes";
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

export function useForgotPasswordFormSchema() {
  return z
    .object({
      identifier: z.string().min(1, { message: "Please enter your email or phone number." }),
    })
    .superRefine((value, ctx) => {
      const trimmed = value.identifier.trim();
      if (trimmed.includes("@")) {
        if (!z.email().safeParse(trimmed).success) {
          ctx.addIssue({
            code: "custom",
            path: ["identifier"],
            message: "Please enter a valid email address.",
          });
        }
        return;
      }
      const digits = trimmed.replace(/[\s-]/g, "");
      if (!/^\+?\d{7,15}$/.test(digits)) {
        ctx.addIssue({
          code: "custom",
          path: ["identifier"],
          message: "Please enter a valid phone number.",
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
  const forgotPasswordFormSchema = useForgotPasswordFormSchema();

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
          toast.warning("Too many requests. Please try again later.", {
            description: seconds ? `Try again in ${seconds} seconds.` : undefined,
          });
          return;
        }
        if (isServerError(error)) {
          toast.error("Something went wrong. Please try again later.");
          return;
        }
        toast.error("We couldn't send the reset link. Please try again later.");
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
