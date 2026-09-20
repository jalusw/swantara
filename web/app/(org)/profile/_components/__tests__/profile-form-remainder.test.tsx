import { fireEvent, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { HttpResponse, http } from "msw";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithProviders, server } from "@/lib/tests";
import { ProfileForm } from "../profile-form";

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
      [{ target, contentRect: { width: 200, height: 200 } }] as ResizeObserverEntry[],
      this,
    );
  }
  unobserve() {}
  disconnect() {}
}

function stubCropperApis() {
  vi.spyOn(window, "Image").mockImplementation(function FakeImageConstructor() {
    return new FakeImage();
  } as unknown as typeof Image);
  window.ResizeObserver = FakeResizeObserver as unknown as typeof ResizeObserver;
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
}

afterEach(() => {
  vi.restoreAllMocks();
});

function useProfileHandlers(user: Record<string, unknown>) {
  server.use(
    http.get("*/api/v1/me", () =>
      HttpResponse.json({ success: true, message: "OK.", data: { user } }),
    ),
  );
}

const baseUser = {
  id: 1,
  username: "alex",
  first_name: "Alex",
  last_name: "Rivera",
  email: "alex@acme.com",
  phone: null,
  avatar: null,
  bio: "Hello",
  birthday: "1990-05-20T00:00:00Z",
  active: true,
  sex: "male",
  address: "Main street 1",
  city: "Jakarta",
  postal_code: "10110",
  system_role_id: 1,
  email_verified_at: null,
  phone_verified_at: null,
  created_at: "2026-01-01T00:00:00Z",
  updated_at: "2026-01-01T00:00:00Z",
};

beforeEach(() => {});

describe("ProfileForm remainder", () => {
  it("renders fallback initials for the avatar", async () => {
    useProfileHandlers({ ...baseUser, avatar: "https://cdn.example.com/a.jpg" });
    renderWithProviders(<ProfileForm />);

    expect(await screen.findByText("AR")).toBeInTheDocument();
  });

  it("renders fallback initials when names are missing", async () => {
    useProfileHandlers({ ...baseUser, first_name: "", last_name: "", avatar: null });
    renderWithProviders(<ProfileForm />);

    await screen.findByText("Upload photo");
  });

  it("opens the cropper after selecting a photo", async () => {
    stubCropperApis();
    useProfileHandlers(baseUser);
    renderWithProviders(<ProfileForm />);

    await screen.findByDisplayValue("Alex");

    const file = new File(["photo"], "photo.png", { type: "image/png" });
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    expect(await screen.findByRole("button", { name: "Apply crop" })).toBeInTheDocument();
  });

  it("uploads the cropped avatar and shows a success toast", async () => {
    stubCropperApis();
    let avatarUploaded = false;
    useProfileHandlers(baseUser);
    server.use(
      http.put("*/api/v1/me/avatar", () => {
        avatarUploaded = true;
        return HttpResponse.json({ success: true, message: "OK.", data: {} });
      }),
    );
    const originalFetch = globalThis.fetch;
    vi.spyOn(globalThis, "fetch").mockImplementation(async (input, init) => {
      if (typeof input === "string" && input.startsWith("data:")) {
        return { blob: async () => new Blob(["cropped"], { type: "image/jpeg" }) } as Response;
      }
      return originalFetch(input, init);
    });
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    await screen.findByDisplayValue("Alex");

    const file = new File(["photo"], "photo.png", { type: "image/png" });
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [file] } });

    await user.click(await screen.findByRole("button", { name: "Apply crop" }));

    await waitFor(() => expect(avatarUploaded).toBe(true));
  });

  it("shows a validation error when the first name is cleared", async () => {
    useProfileHandlers(baseUser);
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    const firstName = await screen.findByDisplayValue("Alex");
    await user.clear(firstName);
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    expect(await screen.findByText("First name is required")).toBeInTheDocument();
  });

  it("shows an error toast when saving fails", async () => {
    useProfileHandlers(baseUser);
    server.use(
      http.put("*/api/v1/users/:id", () =>
        HttpResponse.json({ success: false, message: "Cannot save." }, { status: 500 }),
      ),
    );
    const user = userEvent.setup();
    renderWithProviders(<ProfileForm />);

    await screen.findByDisplayValue("Alex");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    await waitFor(() => expect(screen.getByDisplayValue("Alex")).toBeInTheDocument());
  });
});
