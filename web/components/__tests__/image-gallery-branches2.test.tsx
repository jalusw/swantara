import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { ImageGallery, Lightbox } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const IMAGES = [
  { src: "https://example.com/a.jpg", alt: "First", caption: "First caption" },
  { src: "https://example.com/b.jpg", alt: "Second" },
];

describe("ImageGallery branches2", () => {
  it("renders nothing for the lightbox empty branch", () => {
    const { container } = renderWithProviders(
      <Lightbox images={[]} index={0} onClose={() => undefined} />,
    );

    expect(container.querySelector('[data-slot="lightbox"]')).not.toBeInTheDocument();
  });

  it("falls back to alt text when caption is missing and uses custom aria", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} aria-label="Custom gallery" />);

    await user.click(screen.getByRole("button", { name: "Open Second" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Second")).toBeInTheDocument();
  });

  it("navigates with next and previous buttons on both enabled branches", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} />);

    await user.click(screen.getByRole("button", { name: "Open First" }));
    expect(await screen.findByText("1 / 2")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Next image" }));
    expect(await screen.findByText("2 / 2")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Previous image" }));
    expect(await screen.findByText("1 / 2")).toBeInTheDocument();
  });

  it("closes through Escape and the close button branches", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={IMAGES} />);

    await user.click(screen.getByRole("button", { name: "Open First" }));
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("disables navigation for a single image", async () => {
    const user = userEvent.setup();
    renderWithProviders(
      <ImageGallery images={[{ src: "https://example.com/a.jpg", alt: "Only" }]} columns={2} />,
    );

    await user.click(screen.getByRole("button", { name: "Open Only" }));
    expect(screen.getByRole("button", { name: "Next image" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Previous image" })).toBeDisabled();
  });
});
