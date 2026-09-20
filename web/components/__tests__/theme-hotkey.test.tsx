import { render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useThemeStore } from "@/stores/theme.store";
import { ThemeHotkey } from "../theme-hotkey";

describe("ThemeHotkey", () => {
  afterEach(() => {
    useThemeStore.setState({ theme: "dark", resolvedTheme: "dark" });
  });

  it("toggles theme on Mod+Shift+T", () => {
    useThemeStore.setState({ theme: "dark", resolvedTheme: "dark" });

    render(<ThemeHotkey />);

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "T",
        ctrlKey: true,
        shiftKey: true,
        bubbles: true,
      }),
    );

    expect(useThemeStore.getState().theme).toBe("light");
  });

  it("toggles theme on Cmd+Shift+T (macOS)", () => {
    useThemeStore.setState({ theme: "dark", resolvedTheme: "dark" });

    render(<ThemeHotkey />);

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "T",
        metaKey: true,
        shiftKey: true,
        bubbles: true,
      }),
    );

    expect(useThemeStore.getState().theme).toBe("light");
  });

  it("does not toggle when only Mod+T is pressed (no Shift)", () => {
    useThemeStore.setState({ theme: "dark", resolvedTheme: "dark" });

    render(<ThemeHotkey />);

    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "T",
        ctrlKey: true,
        bubbles: true,
      }),
    );

    expect(useThemeStore.getState().theme).toBe("dark");
  });

  it("cleans up the event listener on unmount", () => {
    const removeSpy = vi.spyOn(window, "removeEventListener");

    const { unmount } = render(<ThemeHotkey />);
    unmount();

    expect(removeSpy).toHaveBeenCalledWith("keydown", expect.any(Function));
    removeSpy.mockRestore();
  });
});
