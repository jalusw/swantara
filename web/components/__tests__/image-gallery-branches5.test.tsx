import { fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { ImageGallery } from "@/components/image-gallery";
import { renderWithProviders } from "@/lib/tests";

const images = [
  { src: "u1", alt: "Front view", caption: "Delivery truck front" },
  { src: "u2", alt: "Side view" },
  { src: "u3", alt: "Label close-up" },
];

function openFirst() {
  renderWithProviders(<ImageGallery images={images} />);
  fireEvent.click(screen.getByRole("button", { name: "Open Front view" }));
  expect(screen.getByRole("dialog")).toBeInTheDocument();
}

describe("ImageGallery branches5", () => {
  it("closes the lightbox with Escape", () => {
    openFirst();

    fireEvent.keyDown(window, { key: "Escape" });

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("navigates with arrow keys and wraps around", () => {
    openFirst();

    fireEvent.keyDown(window, { key: "ArrowRight" });
    expect(screen.getByRole("img", { name: "Side view" })).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "ArrowLeft" });
    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
  });

  it("ignores unrelated keys", () => {
    openFirst();

    fireEvent.keyDown(screen.getByRole("dialog"), { key: "a" });

    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });

  it("wraps around at the ends with arrow keys", () => {
    openFirst();

    fireEvent.keyDown(window, { key: "ArrowLeft" });
    expect(screen.getByRole("img", { name: "Label close-up" })).toBeInTheDocument();

    fireEvent.keyDown(window, { key: "ArrowRight" });
    expect(screen.getByRole("img", { name: "Front view" })).toBeInTheDocument();
  });
});
