import { fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ImageGallery } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const images = [
  { src: "u1", alt: "Front view", caption: "Delivery truck front" },
  { src: "u2", alt: "Side view" },
  { src: "u3", alt: "Label close-up" },
];

describe("ImageGallery", () => {
  it("renders accessible thumbnails", () => {
    renderWithProviders(<ImageGallery images={images} />);

    expect(screen.getByRole("button", { name: "Open Front view" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Open Label close-up" })).toBeInTheDocument();
  });

  it("opens a lightbox dialog that navigates with the arrow keys", () => {
    renderWithProviders(<ImageGallery images={images} />);

    fireEvent.click(screen.getByRole("button", { name: "Open Front view" }));

    const dialog = screen.getByRole("dialog", { name: "Image gallery" });
    expect(dialog).toHaveAttribute("aria-modal", "true");
    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
    expect(screen.getByText("Delivery truck front")).toBeInTheDocument();
    expect(screen.getByText("1 / 3")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "ArrowRight" });
    expect(screen.getByRole("img", { name: "Side view" })).toBeInTheDocument();
    expect(screen.getByText("2 / 3")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "ArrowLeft" });
    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
  });

  it("closes the lightbox with Escape", () => {
    renderWithProviders(<ImageGallery images={images} />);

    fireEvent.click(screen.getByRole("button", { name: "Open Side view" }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "Escape" });
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
