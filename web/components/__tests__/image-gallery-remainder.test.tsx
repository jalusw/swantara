import { fireEvent, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { ImageGallery } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const images = [
  { src: "u1", alt: "Front view", caption: "Delivery truck front" },
  { src: "u2", alt: "Side view" },
  { src: "u3", alt: "Label close-up" },
];

describe("ImageGallery remainder", () => {
  it("disables navigation when only one image is present", () => {
    renderWithProviders(<ImageGallery images={[{ src: "only", alt: "Only" }]} />);

    fireEvent.click(screen.getByRole("button", { name: "Open Only" }));

    expect(screen.getByRole("button", { name: "Previous image" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Next image" })).toBeDisabled();
    expect(screen.getByText("1 / 1")).toBeInTheDocument();
  });

  it("falls back to alt text when a caption is missing", () => {
    renderWithProviders(<ImageGallery images={images} />);

    fireEvent.click(screen.getByRole("button", { name: "Open Side view" }));

    expect(screen.getByRole("img", { name: "Side view" })).toBeInTheDocument();
    expect(screen.getByText("Side view")).toBeInTheDocument();
  });

  it("navigates with next and previous buttons", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={images} />);

    await user.click(screen.getByRole("button", { name: "Open Front view" }));
    await user.click(screen.getByRole("button", { name: "Next image" }));

    expect(screen.getByRole("img", { name: "Side view" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Previous image" }));

    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
  });

  it("closes the lightbox with the close button", async () => {
    const user = userEvent.setup();
    renderWithProviders(<ImageGallery images={images} />);

    await user.click(screen.getByRole("button", { name: "Open Front view" }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Close preview" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("respects custom columns, className and aria-label", () => {
    const { container } = renderWithProviders(
      <ImageGallery
        images={images}
        columns={2}
        className="custom-gallery"
        aria-label="Custom gallery"
      />,
    );

    const gallery = container.querySelector('[data-slot="image-gallery"]');
    expect(gallery).toHaveStyle({ gridTemplateColumns: "repeat(2, minmax(0, 1fr))" });
    expect(gallery).toHaveClass("custom-gallery");
    expect(gallery).toHaveAttribute("aria-label", "Custom gallery");
  });

  it("renders nothing interactive when the image list is empty", () => {
    renderWithProviders(<ImageGallery images={[]} />);

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });
});
