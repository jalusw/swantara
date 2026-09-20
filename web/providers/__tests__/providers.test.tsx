import { useQuery } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useDigitalFingerprintStore } from "@/stores/digital-fingerprint.store";
import { DigitalFingerPrintProvider } from "../digital-fingerprint";
import { TanstackQueryProvider } from "../tanstack-query";
import { ThemeProvider } from "../theme";

vi.mock("clientjs", () => ({
  ClientJS: class {
    getFingerprint() {
      return "fingerprint-123";
    }
    getBrowser() {
      return "Chrome";
    }
    getOS() {
      return "Linux";
    }
  },
}));

describe("ThemeProvider", () => {
  it("hydrates the theme into the document root", async () => {
    localStorage.setItem("theme", "light");

    render(
      <ThemeProvider>
        <p>content</p>
      </ThemeProvider>,
    );

    await waitFor(() => expect(document.documentElement.classList.contains("light")).toBe(true));
    expect(screen.getByText("content")).toBeInTheDocument();
  });
});

describe("TanstackQueryProvider", () => {
  it("lets children run queries through its client", async () => {
    function Probe() {
      const query = useQuery({
        queryKey: ["probe"],
        queryFn: async () => "data",
      });
      return <p>{query.data ?? "loading"}</p>;
    }

    render(
      <TanstackQueryProvider>
        <Probe />
      </TanstackQueryProvider>,
    );

    await waitFor(() => expect(screen.getByText("data")).toBeInTheDocument());
    expect(screen.getByText("data")).toHaveTextContent("data");
  });
});

describe("DigitalFingerPrintProvider", () => {
  it("captures the fingerprint, browser and OS into the store", async () => {
    render(
      <DigitalFingerPrintProvider>
        <p>content</p>
      </DigitalFingerPrintProvider>,
    );

    await waitFor(() =>
      expect(useDigitalFingerprintStore.getState().fingerprint).toBe("fingerprint-123"),
    );
    expect(useDigitalFingerprintStore.getState().browser).toBe("Chrome");
    expect(useDigitalFingerprintStore.getState().os).toBe("Linux");
    expect(screen.getByText("content")).toBeInTheDocument();
  });
});
