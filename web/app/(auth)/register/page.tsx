import Link from "next/link";
import AuthShell from "../_components/auth-shell";
import RegisterForm from "./_components/register-form";

export default async function RegisterPage() {
  return (
    <AuthShell title="Create your account" subtitle="Register your new account to get started.">
      <RegisterForm />
      <p className="text-sm text-muted-foreground">
        {"Have an account already?"}{" "}
        <Link
          className="inline-flex min-h-11 items-center rounded-md px-1 font-medium text-primary underline-offset-4 hover:underline focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          href="/login"
        >
          {"Log in"}
        </Link>
      </p>
    </AuthShell>
  );
}
