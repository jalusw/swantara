import { act } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { renderHookWithProviders } from "@/lib/tests";
import { useDebouncedValue } from "../use-debounced-value";

afterEach(() => {
  vi.useRealTimers();
});

describe("useDebouncedValue", () => {
  it("returns the initial value immediately", () => {
    const { result } = renderHookWithProviders(() => useDebouncedValue("first", 200));
    expect(result.current).toBe("first");
  });

  it("updates after the delay elapses", () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHookWithProviders(
      ({ value }: { value: string }) => useDebouncedValue(value, 200),
      { initialProps: { value: "first" } },
    );
    rerender({ value: "second" });
    expect(result.current).toBe("first");
    act(() => {
      vi.advanceTimersByTime(200);
    });
    expect(result.current).toBe("second");
  });

  it("restarts the delay on rapid changes", () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHookWithProviders(
      ({ value }: { value: string }) => useDebouncedValue(value, 200),
      { initialProps: { value: "a" } },
    );
    rerender({ value: "b" });
    act(() => {
      vi.advanceTimersByTime(100);
    });
    rerender({ value: "c" });
    act(() => {
      vi.advanceTimersByTime(100);
    });
    expect(result.current).toBe("a");
    act(() => {
      vi.advanceTimersByTime(100);
    });
    expect(result.current).toBe("c");
  });
});
