"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Camera, Save } from "lucide-react";
import { useTranslations } from "next-intl";
import { useRef, useState } from "react";
import { type UseFormReturn, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/avatar";
import { Button } from "@/components/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/dialog";
import { ErrorState } from "@/components/error-state";
import { Form, FormField, SubmitButton } from "@/components/form";
import { ImageCropper } from "@/components/image-cropper";
import { Input } from "@/components/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/select";
import { Separator } from "@/components/separator";
import { Textarea } from "@/components/textarea";
import { useMeQuery } from "@/lib/hooks/use-me-query";
import { ME_QUERY_KEY } from "@/lib/queries/me";
import { getSwantaraService } from "@/lib/services/swantara";

export function ProfileForm() {
  const t = useTranslations("Profile");
  const meQuery = useMeQuery();
  const queryClient = useQueryClient();
  const user = meQuery.data?.user;
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [cropSrc, setCropSrc] = useState<string | null>(null);

  const schema = z.object({
    firstName: z.string().min(1, t("validation_firstNameRequired")),
    lastName: z.string(),
    bio: z.string(),
    birthday: z.string(),
    sex: z.string(),
    address: z.string(),
    city: z.string(),
    postalCode: z.string(),
  });

  type Values = z.infer<typeof schema>;

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      firstName: user?.firstName ?? "",
      lastName: user?.lastName ?? "",
      bio: user?.bio ?? "",
      birthday: user?.birthday ? new Date(user.birthday).toISOString().split("T")[0]! : "",
      sex: user?.sex ?? "",
      address: user?.address ?? "",
      city: user?.city ?? "",
      postalCode: user?.postalCode ?? "",
    },
    values: {
      firstName: user?.firstName ?? "",
      lastName: user?.lastName ?? "",
      bio: user?.bio ?? "",
      birthday: user?.birthday ? new Date(user.birthday).toISOString().split("T")[0]! : "",
      sex: user?.sex ?? "",
      address: user?.address ?? "",
      city: user?.city ?? "",
      postalCode: user?.postalCode ?? "",
    },
  });

  const initials = [user?.firstName, user?.lastName]
    .filter(Boolean)
    .map((n) => n?.[0])
    .join("")
    .toUpperCase();

  function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (e) => {
      setCropSrc(e.target?.result as string);
    };
    reader.readAsDataURL(file);
    event.target.value = "";
  }

  async function handleCropComplete(croppedDataUrl: string) {
    setCropSrc(null);
    avatarMutation.mutate(croppedDataUrl);
  }

  const avatarMutation = useMutation({
    mutationFn: async (croppedDataUrl: string) => {
      const res = await fetch(croppedDataUrl);
      const blob = await res.blob();
      const file = new File([blob], "avatar.jpg", { type: "image/jpeg" });
      await getSwantaraService().me.updateAvatar(file);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
      toast.success(t("profileUpdated"));
    },
    onError: () => {
      toast.error(t("profileUpdateFailed"));
    },
  });

  const profileMutation = useMutation({
    mutationFn: (values: Values) => {
      if (!user) throw new Error("missing user");
      return getSwantaraService().users.update(user.id, {
        firstName: values.firstName,
        lastName: values.lastName || undefined,
        bio: values.bio || undefined,
        birthday: values.birthday || undefined,
        sex: values.sex || undefined,
        address: values.address || undefined,
        city: values.city || undefined,
        postalCode: values.postalCode || undefined,
      });
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ME_QUERY_KEY });
      toast.success(t("profileUpdated"));
    },
    onError: () => {
      toast.error(t("profileUpdateFailed"));
    },
  });

  function handleSubmit(values: Values) {
    if (!user) return;
    profileMutation.mutate(values);
  }

  if (meQuery.isError) {
    return <ErrorState onReset={() => void meQuery.refetch()} />;
  }

  if (!user) {
    return null;
  }

  return (
    <>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleFileChange}
      />

      <Dialog
        open={Boolean(cropSrc)}
        onOpenChange={(open) => {
          if (!open) setCropSrc(null);
        }}
      >
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{t("cropPhoto")}</DialogTitle>
            <DialogDescription>{t("cropPhotoDescription")}</DialogDescription>
          </DialogHeader>
          {cropSrc && <ImageCropper src={cropSrc} alt={t("avatar")} onCrop={handleCropComplete} />}
        </DialogContent>
      </Dialog>

      <div className="flex flex-col gap-4 sm:gap-6">
        <Card>
          <CardHeader>
            <CardTitle>{t("avatar")}</CardTitle>
            <CardDescription>{t("avatarDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-4">
              <Avatar size="lg">
                {user.avatar ? <AvatarImage src={user.avatar} alt={user.firstName} /> : null}
                <AvatarFallback>{initials}</AvatarFallback>
              </Avatar>
              <Button
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
                disabled={avatarMutation.isPending}
              >
                <Camera />
                {avatarMutation.isPending ? t("uploading") : t("uploadPhoto")}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Form form={form as UseFormReturn<Values>} onSubmit={handleSubmit}>
          <Card>
            <CardHeader>
              <CardTitle>{t("personalInfo")}</CardTitle>
              <CardDescription>{t("personalInfoDescription")}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <FormField name="firstName" label={t("firstName")}>
                {({ field, id }) => <Input {...field} id={id} />}
              </FormField>
              <FormField name="lastName" label={t("lastName")}>
                {({ field, id }) => <Input {...field} id={id} />}
              </FormField>
              <FormField name="birthday" label={t("birthday")}>
                {({ field, id }) => <Input {...field} id={id} type="date" />}
              </FormField>
              <FormField name="sex" label={t("sex")}>
                {({ field, id }) => (
                  <Select value={field.value} onValueChange={field.onChange}>
                    <SelectTrigger id={id} aria-label={t("sex")}>
                      <SelectValue placeholder={t("sex")} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">{t("preferNotToSay")}</SelectItem>
                      <SelectItem value="male">{t("male")}</SelectItem>
                      <SelectItem value="female">{t("female")}</SelectItem>
                    </SelectContent>
                  </Select>
                )}
              </FormField>
              <FormField name="bio" label={t("bio")} className="sm:col-span-2">
                {({ field, id }) => (
                  <Textarea {...field} id={id} placeholder={t("bioPlaceholder")} rows={3} />
                )}
              </FormField>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>{t("address")}</CardTitle>
              <CardDescription>{t("addressDescription")}</CardDescription>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <FormField name="address" label={t("address")} className="sm:col-span-2">
                {({ field, id }) => <Textarea {...field} id={id} rows={2} />}
              </FormField>
              <FormField name="city" label={t("city")}>
                {({ field, id }) => <Input {...field} id={id} />}
              </FormField>
              <FormField name="postalCode" label={t("postalCode")}>
                {({ field, id }) => <Input {...field} id={id} />}
              </FormField>
            </CardContent>
            <Separator />
            <CardContent className="flex items-center justify-end gap-2">
              <SubmitButton loading={profileMutation.isPending}>
                <Save />
                {t("saveChanges")}
              </SubmitButton>
            </CardContent>
          </Card>
        </Form>
      </div>
    </>
  );
}
