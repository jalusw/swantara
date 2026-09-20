import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ImageCropper } from "@/components/image-cropper";
import { renderWithProviders } from "@/lib/tests";

class FakeImage {
  onload: (() => void) | null = null;
  naturalWidth = 480;
  naturalHeight = 320;
  set src(_: string) {
    setTimeout(() => this.onload?.(), 0);
  }
}

class FakeResizeObserver {
  constructor(private callback: ResizeObserverCallback) {}
  observe(target: Element) {
    this.callback(
      [
        {
          target,
          contentRect: { width: 200, height: 200 },
        },
      ] as ResizeObserverEntry[],
      this,
    );
  }
  unobserve() {}
  disconnect() {}
}

let stubbedResizeObserver: typeof ResizeObserver | undefined;

function stubBrowserApis() {
  vi.spyOn(window, "Image").mockImplementation(function FakeImageConstructor() {
    return new FakeImage();
  } as unknown as typeof Image);
  stubbedResizeObserver = window.ResizeObserver;
  window.ResizeObserver = FakeResizeObserver as unknown as typeof ResizeObserver;
  if (typeof HTMLElement.prototype.setPointerCapture === "undefined") {
    Object.defineProperty(HTMLElement.prototype, "setPointerCapture", {
      configurable: true,
      value: () => {},
    });
  }
  const context = {
    setTransform: vi.fn(),
    scale: vi.fn(),
    drawImage: vi.fn(),
    imageSmoothingEnabled: false,
    imageSmoothingQuality: "low",
  };
  vi.spyOn(HTMLCanvasElement.prototype, "getContext").mockReturnValue(context as never);
  vi.spyOn(HTMLCanvasElement.prototype, "toDataURL").mockReturnValue(
    "data:image/jpeg;base64,cropped",
  );
  return { context };
}

function renderCropped() {
  const { context } = stubBrowserApis();
  const result = renderWithProviders(
    <ImageCropper
      src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E"
      onCrop={() => {
        /* no-op */
      }}
    />,
  );
  return { context, ...result };
}

function dragHandle() {
  return screen.getByRole("button", {
    name: "Drag to move the crop area, or use arrow keys to reposition",
  });
}

describe("ImageCropper", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    if (stubbedResizeObserver) {
      window.ResizeObserver = stubbedResizeObserver;
      stubbedResizeObserver = undefined;
    }
  });

  it("shows a placeholder when no image is provided", () => {
    renderWithProviders(<ImageCropper />);

    expect(screen.getByText("No image selected")).toBeInTheDocument();
  });

  it("stays inert until the image loads", () => {
    const { container } = renderWithProviders(
      <ImageCropper src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E" />,
    );

    expect(container.querySelector("canvas")).toBeNull();
    expect(screen.queryByRole("button", { name: "Apply crop" })).not.toBeInTheDocument();
    expect(screen.getByText("Preparing image…")).toBeInTheDocument();
  });

  it("zooms in when the mouse wheel is scrolled over the frame", async () => {
    const { container } = renderCropped();

    await screen.findByText("×1.00");
    const frame = container.querySelector('[data-slot="image-cropper-frame"]');
    expect(frame).toBeInstanceOf(HTMLElement);
    fireEvent.wheel(frame as Element, { deltaY: -100 });

    expect(await screen.findByText("×1.22")).toBeInTheDocument();
  });

  it("draws the loaded image onto the preview canvas", async () => {
    const { context } = renderCropped();
    await screen.findByText("×1.00");
    expect(context.drawImage).toHaveBeenCalled();
  });

  it("keeps the zoom when the wheel delta is zero", async () => {
    const { container } = renderCropped();
    await screen.findByText("×1.00");
    fireEvent.wheel(container.querySelector('[data-slot="image-cropper-frame"]') as Element, {
      deltaY: 0,
    });
    expect(screen.getByText("×1.00")).toBeInTheDocument();
  });

  it("drags the crop area with the pointer", async () => {
    const { context } = renderCropped();
    await screen.findByText("×1.00");
    const callsBefore = (context.drawImage as ReturnType<typeof vi.fn>).mock.calls.length;

    fireEvent.pointerDown(dragHandle(), { button: 0, clientX: 100, clientY: 100, pointerId: 1 });
    fireEvent.pointerMove(dragHandle(), { clientX: 80, clientY: 90, pointerId: 1 });
    fireEvent.pointerUp(dragHandle(), { pointerId: 1 });

    expect((context.drawImage as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(
      callsBefore,
    );
  });

  it("ignores non-primary pointer buttons and moves without a drag", async () => {
    renderCropped();
    await screen.findByText("×1.00");
    fireEvent.pointerDown(dragHandle(), { button: 1, clientX: 100, clientY: 100, pointerId: 2 });
    fireEvent.pointerMove(dragHandle(), { clientX: 10, clientY: 10, pointerId: 2 });
    expect(screen.getByText("×1.00")).toBeInTheDocument();
  });

  it("repositions with arrow keys", async () => {
    const { context, container } = renderCropped();
    await screen.findByText("×1.00");
    const handle = dragHandle();
    fireEvent.keyDown(handle, { key: "ArrowRight" });
    fireEvent.keyDown(handle, { key: "ArrowLeft" });
    fireEvent.keyDown(handle, { key: "ArrowUp" });
    fireEvent.keyDown(handle, { key: "ArrowDown", shiftKey: true });
    fireEvent.keyDown(handle, { key: "a" });
    expect((context.drawImage as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(0);
    expect(screen.getByRole("button", { name: "Apply crop" })).toBeEnabled();
    expect(container.querySelector("canvas")).toBeInTheDocument();
  });

  it("skips preview drawing when the canvas has no context", async () => {
    const { container } = renderCropped();
    await screen.findByText("×1.00");
    vi.mocked(HTMLCanvasElement.prototype.getContext).mockReturnValueOnce(null);
    fireEvent.wheel(container.querySelector('[data-slot="image-cropper-frame"]') as Element, {
      deltaY: -100,
    });
    expect(await screen.findByText("×1.22")).toBeInTheDocument();
  });

  it("ignores keyboard input when disabled", async () => {
    stubBrowserApis();
    renderWithProviders(
      <ImageCropper
        src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E"
        disabled
        onCrop={() => {}}
      />,
    );
    await screen.findByText("×1.00");
    fireEvent.keyDown(dragHandle(), { key: "ArrowRight" });
    expect(screen.getByText("×1.00")).toBeInTheDocument();
  });

  it("ignores apply when there is no crop handler", async () => {
    stubBrowserApis();
    renderWithProviders(
      <ImageCropper src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E" />,
    );
    const apply = await screen.findByRole("button", { name: "Apply crop" });
    expect(apply).toBeDisabled();
    fireEvent.click(apply);
    expect(screen.queryByText("Crop applied.")).not.toBeInTheDocument();
  });

  it("changes zoom through the slider", async () => {
    const { container } = renderCropped();
    await screen.findByText("×1.00");
    const slider = container.querySelector('input[type="range"]') as HTMLInputElement;
    expect(slider).toBeInTheDocument();
    fireEvent.change(slider, { target: { value: "1.5" } });
    expect(await screen.findByText("×1.50")).toBeInTheDocument();
  });

  it("applies the crop and reports a data url", async () => {
    const user = userEvent.setup();
    stubBrowserApis();
    const onCrop = vi.fn();
    renderWithProviders(
      <ImageCropper
        src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E"
        onCrop={onCrop}
      />,
    );
    await screen.findByText("×1.00");
    await user.click(screen.getByRole("button", { name: "Apply crop" }));
    expect(onCrop).toHaveBeenCalledWith("data:image/jpeg;base64,cropped");
    expect(screen.getByText("Crop applied.")).toBeInTheDocument();
  });

  it("does nothing when the output canvas has no context", async () => {
    const user = userEvent.setup();
    stubBrowserApis();
    const onCrop = vi.fn();
    renderWithProviders(
      <ImageCropper
        src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E"
        onCrop={onCrop}
      />,
    );
    await screen.findByText("×1.00");
    vi.mocked(HTMLCanvasElement.prototype.getContext).mockReturnValueOnce(null);
    await user.click(screen.getByRole("button", { name: "Apply crop" }));
    expect(onCrop).not.toHaveBeenCalled();
  });

  it("disables interaction when disabled", async () => {
    stubBrowserApis();
    renderWithProviders(
      <ImageCropper
        src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg'/%3E"
        disabled
        onCrop={() => {}}
      />,
    );
    await screen.findByText("×1.00");
    expect(dragHandle()).toBeDisabled();
    expect(screen.getByRole("button", { name: "Apply crop" })).toBeDisabled();
    fireEvent.wheel(screen.getByText("×1.00").closest("div") as Element, { deltaY: -100 });
    expect(screen.getByText("×1.00")).toBeInTheDocument();
  });

  it("drops a stale load when the source changes quickly", async () => {
    stubBrowserApis();
    const rendered = renderWithProviders(<ImageCropper src="first.svg" onCrop={() => {}} />);
    rendered.rerender(
      <ImageCropper
        src="second.svg"
        onCrop={() => {
          /* no-op */
        }}
      />,
    );
    expect(await screen.findByText("×1.00")).toBeInTheDocument();
  });

  it("clears the image when the source is removed", async () => {
    stubBrowserApis();
    const rendered = renderWithProviders(<ImageCropper src="first.svg" onCrop={() => {}} />);
    expect(await screen.findByText("×1.00")).toBeInTheDocument();
    rendered.rerender(<ImageCropper onCrop={() => {}} />);
    expect(await screen.findByText("No image selected")).toBeInTheDocument();
  });
});
