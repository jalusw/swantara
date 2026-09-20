import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useUnsavedChanges } from "@/components/form";

describe("useUnsavedChanges", () => {
  it("confirms navigation when dirty and the user accepts", () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(true);
    const { result } = renderHook(() => useUnsavedChanges(true));
    const onNavigate = vi.fn();
    result.current.confirmNavigation(onNavigate);
    expect(confirmSpy).toHaveBeenCalledOnce();
    expect(onNavigate).toHaveBeenCalledOnce();
    confirmSpy.mockRestore();
  });

  it("aborts navigation when dirty and the user declines", () => {
    const confirmSpy = vi.spyOn(window, "confirm").mockReturnValue(false);
    const { result } = renderHook(() => useUnsavedChanges(true));
    const onNavigate = vi.fn();
    result.current.confirmNavigation(onNavigate);
    expect(onNavigate).not.toHaveBeenCalled();
    confirmSpy.mockRestore();
  });

  it("navigates without prompting when clean", () => {
    const confirmSpy = vi.spyOn(window, "confirm");
    const { result } = renderHook(() => useUnsavedChanges(false));
    const onNavigate = vi.fn();
    result.current.confirmNavigation(onNavigate);
    expect(confirmSpy).not.toHaveBeenCalled();
    expect(onNavigate).toHaveBeenCalledOnce();
    confirmSpy.mockRestore();
  });
});
