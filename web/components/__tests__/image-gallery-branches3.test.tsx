import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ImageGallery, Lightbox } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const IMAGES = [
  { src: "https://example.com/a.jpg", alt: "First", caption: "First caption" },
  { src: "https://example.com/b.jpg", alt: "Second" },
];

describe("ImageGallery branches3", () => {
  it("renders nothing when the index is out of range", () => {
    const { container } = renderWithProviders(
      <Lightbox images={IMAGES} index={9} onClose={() => undefined} />,
    );

    expect(container.querySelector('[data-slot="lightbox"]')).not.toBeInTheDocument();
  });

  it("uses a custom aria label for the lightbox dialog", () => {
    renderWithProviders(
      <Lightbox images={IMAGES} index={0} onClose={() => undefined} aria-label="Custom preview" />,
    );

    expect(screen.getByRole("dialog", { name: "Custom preview" })).toBeInTheDocument();
  });

  it("ignores arrow keys when no index handler is provided", async () => {
    const user = userEvent.setup();
    renderWithProviders(<Lightbox images={IMAGES} index={0} onClose={() => undefined} />);

    expect(await screen.findByText("1 / 2")).toBeInTheDocument();
    await user.keyboard("{ArrowRight}");
    expect(screen.getByText("1 / 2")).toBeInTheDocument();
    await user.keyboard("{ArrowLeft}");
    expect(screen.getByText("1 / 2")).toBeInTheDocument();
  });

  it("wraps arrow navigation from the last image to the first", async () => {
    const user = userEvent.setup();
    const onIndexChange = vi.fn();
    renderWithProviders(
      <Lightbox
        images={IMAGES}
        index={1}
        onIndexChange={onIndexChange}
        onClose={() => undefined}
      />,
    );

    expect(await screen.findByText("2 / 2")).toBeInTheDocument();
    await user.keyboard("{ArrowRight}");
    expect(onIndexChange).toHaveBeenCalledWith(0);
  });

  it("closes when the backdrop button is clicked", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderWithProviders(<Lightbox images={IMAGES} index={0} onClose={onClose} />);

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Close preview" }));
    expect(onClose).toHaveBeenCalled();
  });

  it("opens the lightbox from a thumbnail without a caption", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} />);

    await user.click(screen.getByRole("button", { name: "Open Second" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("2 / 2")).toBeInTheDocument();
  });
});
