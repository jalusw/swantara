import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import { StatusCodes } from "http-status-codes";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useLayoutEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { axiosInstance } from "@/lib/client/axios";
import { isNetworkError, isServerError } from "@/lib/client/error";
import { getCsrfToken } from "@/lib/constants/cookies";
import { ME_ORGANIZATIONS_QUERY_KEY, ME_QUERY_KEY } from "@/lib/queries/me";
import type { RegistrationStep } from "./register-form.types";
import {
  type RegisterFormSchema,
  registerFormDefaultValues,
  stepFields,
  useRegisterFormSchema,
} from "./register-form-schema";
import { type RegisterFormStore, useRegisterFormStore } from "./register-form-store";

export function useRegisterForm({
  defaultValues = registerFormDefaultValues,
}: UseRegisterFormParams) {
  const registerFormSchema = useRegisterFormSchema();
  const router = useRouter();
  const queryClient = useQueryClient();

  const form = useForm<RegisterFormSchema>({
    resolver: zodResolver(registerFormSchema),
    defaultValues,
  });

  const goToNextStep = useCallback(async () => {
    const { step } = useRegisterFormStore.getState();
    const fields = stepFields[step];

    const isValid = await form.trigger(fields);
    if (!isValid) {
      return;
    }

    form.clearErrors();

    if (step === "email") {
      const email = form.getValues("email").trim();
      useRegisterFormStore.getState().setCheckingEmail(true);
      try {
        const response = await axiosInstance.post<{ data: { available: boolean } }>(
          "/auth/email/check",
          { email },
        );
        if (!response.data.data.available) {
          form.setError("email", { message: "This email is already taken." });
          return;
        }
      } catch {
        toast.warning("Couldn't verify email right now. Please try again.");
      } finally {
        useRegisterFormStore.getState().setCheckingEmail(false);
      }
    }

    switch (step) {
      case "name":
        useRegisterFormStore.getState().setStep("email");
        break;
      case "email":
        useRegisterFormStore.getState().setStep("password");
        break;
      case "password":
        break;
    }
  }, [form]);

  const goToPrevStep = useCallback(() => {
    const { step } = useRegisterFormStore.getState();
    form.clearErrors();
    switch (step) {
      case "email":
        useRegisterFormStore.getState().setStep("name");
        break;
      case "password":
        useRegisterFormStore.getState().setStep("email");
        break;
    }
  }, [form]);

  const handleSubmit = form.handleSubmit(async (data: RegisterFormSchema) => {
    useRegisterFormStore.getState().setPending(true);
    try {
      const { passwordConfirmation: _passwordConfirmation, ...rest } = data;
      await axiosInstance.post("/auth/register", {
        ...rest,
        phone: "",
      });
      try {
        const response = await fetch("/api/v1/auth/login", {
          method: "POST",
          headers: {
            "content-type": "application/json",
            "x-csrf-token": getCsrfToken() ?? "",
          },
          credentials: "include",
          body: JSON.stringify({ email: data.email.trim(), password: data.password }),
        });
        if (!response.ok) {
          router.push("/login");
          return;
        }
      } catch {
        router.push("/login");
        return;
      }
      queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
      queryClient.invalidateQueries({ queryKey: ME_ORGANIZATIONS_QUERY_KEY });
      router.push("/onboarding");
      return;
    } catch (error: unknown) {
      if (isAxiosError(error)) {
        const status = error.response?.status;
        const data = error.response?.data as
          | {
              message?: string;
              error?: string;
              errorCode?: string;
              error_code?: string;
              field_errors?: Array<{ field: string; message: string }>;
            }
          | undefined;
        const message = (data?.message ?? "").toLowerCase();
        const isEmailTaken =
          status === StatusCodes.CONFLICT ||
          (status === StatusCodes.UNPROCESSABLE_ENTITY &&
            (message.includes("already registered") || message.includes("already taken")));

        if (isEmailTaken) {
          useRegisterFormStore.getState().setStep("email");
          form.setError("email", { message: "This email is already taken." });
          return;
        }

        if (!error.response) {
          toast.error("Network error. Please check your connection.");
          return;
        }

        if (status === StatusCodes.TOO_MANY_REQUESTS) {
          const retryAfter = error.response.headers?.["retry-after"];
          const seconds = retryAfter ? parseInt(retryAfter as string, 10) : undefined;
          toast.warning("Too many requests. Please try again later.", {
            description: seconds ? `Try again in ${seconds} seconds.` : undefined,
          });
          return;
        }

        if ((status ?? 0) >= StatusCodes.INTERNAL_SERVER_ERROR) {
          toast.error("Something went wrong. Please try again later.");
          return;
        }

        if (data?.field_errors?.length) {
          for (const fieldError of data.field_errors) {
            const field = fieldError.field as keyof RegisterFormSchema;
            if (field in registerFormDefaultValues) {
              form.setError(field, { message: fieldError.message });
            }
          }
          if (data.message) toast.error(data.message);
          return;
        }

        if (status === StatusCodes.BAD_REQUEST || status === StatusCodes.UNPROCESSABLE_ENTITY) {
          if (message) {
            const isEmailMessage = message.includes("email");
            if (isEmailMessage) {
              useRegisterFormStore.getState().setStep("email");
              form.setError("email", { message: "This email is already taken." });
            } else {
              toast.error(data?.message ?? "Registration failed. Please try again.");
            }
            return;
          }
        }
      }

      if (isNetworkError(error)) {
        toast.error("Network error. Please check your connection.");
        return;
      }
      if (isServerError(error)) {
        toast.error("Something went wrong. Please try again later.");
        return;
      }
      form.setError("email", { message: "Registration failed. Please try again." });
      return;
    } finally {
      useRegisterFormStore.getState().setPending(false);
    }
  });

  const emailValue = form.watch("email");

  useEffect(() => {
    const email = emailValue?.trim();
    useRegisterFormStore.getState().setEmail(email ?? "");
  }, [emailValue]);

  useEffect(() => {
    useRegisterFormStore.setState({
      step: "name",
      isFirstStep: true,
      isLastStep: false,
      isPending: false,
      isCheckingEmail: false,
      email: "",
    });
  }, []);

  useLayoutEffect(() => {
    useRegisterFormStore.setState({ goToNextStep, goToPrevStep });
  }, [goToNextStep, goToPrevStep]);

  return {
    form,
    handleSubmit,
  };
}

export type { RegistrationStep, RegisterFormSchema };
export type { RegisterFormStore };
export { useRegisterFormStore };

export { useRegisterFormSchema };

export type UseRegisterFormParams = {
  defaultValues?: RegisterFormSchema;
};
