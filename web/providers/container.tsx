import { CommandPaletteProvider } from "@/components/command-palette";
import { ThemeHotkey } from "@/components/theme-hotkey";
import { Toaster } from "@/components/toast";
import { HotkeysProvider } from "./hotkeys";
import { SessionProvider } from "./session";
import { TanstackQueryProvider } from "./tanstack-query";
import { ThemeProvider } from "./theme";

export type ProviderContainer = {
  children: React.ReactNode;
};

export async function ProviderContainer({ children }: ProviderContainer) {
  return (
    <ThemeProvider>
      <TanstackQueryProvider>
        <SessionProvider>
          <HotkeysProvider>
            <ThemeHotkey />
            <CommandPaletteProvider>{children}</CommandPaletteProvider>
          </HotkeysProvider>
        </SessionProvider>
        <Toaster />
      </TanstackQueryProvider>
    </ThemeProvider>
  );
}
