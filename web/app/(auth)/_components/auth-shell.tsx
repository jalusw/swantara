import { Card, CardContent, CardHeader } from "@/components/card";
import { LocaleSwitcher } from "@/components/locale-switcher";

export type AuthShellProps = {
  title: string;
  subtitle?: string;
  children: React.ReactNode;
};

export default async function AuthShell({ title, subtitle = "", children }: AuthShellProps) {
  return (
    <div data-slot="auth-shell" className="relative flex min-h-screen bg-muted/20">
      <div className="absolute top-4 right-4">
        <LocaleSwitcher variant="outline" />
      </div>
      <div className="mx-auto flex w-full max-w-md flex-col justify-center p-4 sm:p-6">
        <Card>
          <CardHeader className="border-b">
            <h1 className="font-heading text-xl leading-snug">{title.trim()}</h1>
            <p className="text-sm text-muted-foreground">{subtitle.trim()}</p>
          </CardHeader>
          <CardContent className="flex flex-col gap-4">{children}</CardContent>
        </Card>
      </div>
    </div>
  );
}
