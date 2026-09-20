import { act } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import { useLineResetEffect } from "../use-line-reset-effect";

describe("useLineResetEffect", () => {
  it("resets to a single empty line when opened", () => {
    const setLines = vi.fn();
    const createEmptyLine = vi.fn(() => ({ id: "new" }));
    renderHookWithProviders(() => useLineResetEffect(true, setLines, createEmptyLine));
    expect(createEmptyLine).toHaveBeenCalledTimes(1);
    expect(setLines).toHaveBeenCalledWith([{ id: "new" }]);
  });

  it("does nothing while closed and resets again on reopen", () => {
    const setLines = vi.fn();
    const createEmptyLine = vi.fn(() => ({ id: "row" }));
    const { rerender } = renderHookWithProviders(
      ({ open }) => useLineResetEffect(open, setLines, createEmptyLine),
      { initialProps: { open: false } },
    );
    expect(setLines).not.toHaveBeenCalled();
    act(() => {
      rerender({ open: true });
    });
    expect(setLines).toHaveBeenCalledWith([{ id: "row" }]);
  });
});
