"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useEffect, useTransition } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { getCsrfToken } from "@/lib/constants/cookies";
import { ME_ORGANIZATIONS_QUERY_KEY, ME_QUERY_KEY } from "@/lib/queries/me";
import { useDigitalFingerprintStore } from "@/stores/digital-fingerprint.store";

export function useLoginFormSchema() {
  return z.object({
    email: z.string().min(1, { message: "Email is required." }),
    password: z.string().min(1, { message: "Password is required" }),
  });
}

export type LoginFormsSchema = z.infer<ReturnType<typeof useLoginFormSchema>>;

const loginFormDefaultValues: LoginFormsSchema = {
  email: "",
  password: "",
};

export type UseLoginFormParams = {
  defaultValues?: LoginFormsSchema;
};

export function useLoginForm({ defaultValues = loginFormDefaultValues }: UseLoginFormParams) {
  const [isPending, startTransition] = useTransition();
  useDigitalFingerprintStore((state) => state.fingerprint);
  const router = useRouter();
  const queryClient = useQueryClient();
  const loginFormSchema = useLoginFormSchema();

  const form = useForm<LoginFormsSchema>({
    resolver: zodResolver(loginFormSchema),
    defaultValues,
  });

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (typeof PasswordCredential === "undefined") return;
    if (!navigator.credentials?.get) return;

    navigator.credentials
      .get({ password: true, mediation: "optional" } as CredentialRequestOptions)
      .then((credential) => {
        if (credential instanceof PasswordCredential) {
          form.setValue("email", credential.id, {
            shouldDirty: true,
            shouldTouch: true,
          });
          form.setValue("password", credential.password, {
            shouldDirty: true,
            shouldTouch: true,
          });
        }
      })
      .catch(() => {});
  }, [form]);

  const handleSubmit = form.handleSubmit((data: LoginFormsSchema) => {
    startTransition(async () => {
      try {
        const response = await fetch("/api/v1/auth/login", {
          method: "POST",
          headers: {
            "content-type": "application/json",
            "x-csrf-token": getCsrfToken() ?? "",
          },
          credentials: "include",
          body: JSON.stringify(data),
        });
        if (response.ok) {
          if (typeof PasswordCredential !== "undefined" && navigator.credentials?.store) {
            try {
              const credential = new PasswordCredential({
                id: data.email,
                password: data.password,
                name: data.email,
              });
              await navigator.credentials.store(credential);
            } catch {}
          }
          queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
          queryClient.invalidateQueries({
            queryKey: ME_ORGANIZATIONS_QUERY_KEY,
          });
          router.push("/onboarding");
          return;
        }
        if (response.status === 401) {
          toast.error("Invalid email or password.");
          return;
        }
        if (response.status === 429) {
          const retryAfter = response.headers.get("retry-after");
          const seconds = retryAfter ? parseInt(retryAfter, 10) : undefined;
          toast.warning("Too many requests. Please try again later.", {
            description: seconds ? `Try again in ${seconds} seconds.` : undefined,
          });
          return;
        }
        if (response.status >= 500) {
          toast.error("Something went wrong. Please try again later.");
          return;
        }
        toast.error("Can't sign in at the moment, please try again later.");
      } catch {
        toast.error("Network error. Please check your connection.");
      }
    });
  });

  return {
    isPending,
    handleSubmit,
    form,
  };
}
