import { fireEvent, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { SignaturePad } from "@/components/signature-pad";
import { renderWithProviders } from "@/lib/tests";

const context = vi.hoisted(() => ({
  setTransform: vi.fn(),
  beginPath: vi.fn(),
  moveTo: vi.fn(),
  lineTo: vi.fn(),
  stroke: vi.fn(),
  clearRect: vi.fn(),
  lineCap: "",
  lineJoin: "",
  lineWidth: 0,
  strokeStyle: "",
}));

class FakeResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
}

let stubbedResizeObserver: typeof ResizeObserver | undefined;

function renderPad(onChange?: (dataUrl: string | null) => void) {
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(
    context as unknown as CanvasRenderingContext2D,
  );
  vi.spyOn(HTMLCanvasElement.prototype, "toDataURL").mockReturnValue("data:image/png;base64,FAKE");
  stubbedResizeObserver = window.ResizeObserver;
  window.ResizeObserver = FakeResizeObserver as unknown as typeof ResizeObserver;
  return renderWithProviders(<SignaturePad onChange={onChange} />);
}

const overlayRole = (name = /draw with a mouse/) => screen.getByRole("button", { name });

describe("SignaturePad", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    if (stubbedResizeObserver) {
      window.ResizeObserver = stubbedResizeObserver;
      stubbedResizeObserver = undefined;
    }
  });

  it("shows a guide placeholder and no clear action while empty", () => {
    renderPad();

    expect(screen.getByText("Sign here")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /clear signature/i })).not.toBeInTheDocument();
  });

  it("draws strokes and emits a PNG data URL on release", () => {
    const onChange = vi.fn();
    renderPad(onChange);

    const surface = overlayRole();
    fireEvent.pointerDown(surface, {
      button: 0,
      pointerId: 1,
      clientX: 10,
      clientY: 10,
    });
    fireEvent.pointerMove(surface, { pointerId: 1, clientX: 60, clientY: 40 });
    fireEvent.pointerUp(surface, { pointerId: 1 });

    expect(context.moveTo).toHaveBeenCalled();
    expect(context.lineTo).toHaveBeenCalled();
    expect(context.stroke).toHaveBeenCalled();
    expect(onChange).toHaveBeenCalledWith("data:image/png;base64,FAKE");
    expect(screen.getByRole("button", { name: /clear signature/i })).toBeInTheDocument();
  });

  it("clears the canvas and emits null", async () => {
    const onChange = vi.fn();
    renderPad(onChange);

    fireEvent.pointerDown(overlayRole(), {
      button: 0,
      pointerId: 1,
      clientX: 5,
      clientY: 5,
    });
    fireEvent.pointerUp(overlayRole(), { pointerId: 1 });
    fireEvent.click(screen.getByRole("button", { name: /clear signature/i }));

    expect(context.clearRect).toHaveBeenCalled();
    expect(onChange).toHaveBeenLastCalledWith(null);
    expect(screen.getByText("Sign here")).toBeInTheDocument();
  });
});
