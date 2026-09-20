import AuthShell from "../_components/auth-shell";
import ResetPasswordForm from "./_components/reset-password-form";

export type ResetPasswordPageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function ResetPasswordPage({ searchParams }: ResetPasswordPageProps) {
  const { token } = await searchParams;

  return (
    <AuthShell title={"Set a new password"} subtitle={"Choose a new password for your account."}>
      <ResetPasswordForm token={token} />
    </AuthShell>
  );
}
