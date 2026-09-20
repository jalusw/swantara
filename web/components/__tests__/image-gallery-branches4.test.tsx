import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { ImageGallery, Lightbox } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const IMAGES = [
  { src: "https://example.com/a.jpg", alt: "First", caption: "First caption" },
  { src: "https://example.com/b.jpg", alt: "Second" },
  { src: "https://example.com/c.jpg", alt: "Third" },
];

describe("ImageGallery branches4", () => {
  it("returns null for out-of-range index", () => {
    const { container } = renderWithProviders(
      <Lightbox images={IMAGES} index={9} onClose={() => undefined} />,
    );

    expect(container.querySelector('[data-slot="lightbox"]')).not.toBeInTheDocument();
  });

  it("navigates with ArrowLeft and ArrowRight keys", async () => {
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

    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await user.keyboard("{ArrowLeft}");
    expect(onIndexChange).toHaveBeenCalledWith(0);
    await user.keyboard("{ArrowRight}");
    expect(onIndexChange).toHaveBeenCalledWith(2);
  });

  it("closes through the backdrop button", async () => {
    const user = userEvent.setup();
    const onClose = vi.fn();
    renderWithProviders(<Lightbox images={IMAGES} index={0} onClose={onClose} />);

    await screen.findByRole("dialog");
    const backdrop = document.querySelector('[data-slot="lightbox"] > button');
    if (backdrop) await user.click(backdrop as HTMLElement);
    expect(onClose).toHaveBeenCalled();
  });

  it("uses a custom aria label for the gallery lightbox", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} aria-label="Custom set" columns={2} />);

    await user.click(screen.getByRole("button", { name: "Open First" }));
    expect(await screen.findByRole("dialog", { name: "Custom set" })).toBeInTheDocument();
  });

  it("wraps from first to last with previous button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} />);

    await user.click(screen.getByRole("button", { name: "Open First" }));
    expect(await screen.findByText("1 / 3")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Previous image" }));
    expect(await screen.findByText("3 / 3")).toBeInTheDocument();
  });

  it("closes the gallery lightbox and reopens another image", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} />);

    await user.click(screen.getByRole("button", { name: "Open Second" }));
    expect(await screen.findByText("2 / 3")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Open Third" }));
    expect(await screen.findByText("3 / 3")).toBeInTheDocument();
  });
});
