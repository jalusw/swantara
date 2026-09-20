import { screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { MapView } from "@/components/map-view";
import { renderWithProviders } from "@/lib/tests";

type MapOptions = {
  container: HTMLElement;
  style: string | Record<string, unknown>;
  center: [number, number];
  zoom: number;
};

const mapClass = vi.hoisted(() => {
  class MaplibreMap {
    constructor(readonly options: MapOptions) {}
    addControl = vi.fn();
    remove = vi.fn();
    addTo = vi.fn();
  }
  class Marker {
    setLngLat = vi.fn().mockReturnThis();
    setPopup = vi.fn().mockReturnThis();
    addTo = vi.fn();
  }
  class Popup {
    setText = vi.fn().mockReturnThis();
  }
  return {
    Map: vi.fn(MaplibreMap),
    Marker: vi.fn(Marker),
    Popup: vi.fn(Popup),
    NavigationControl: class NavigationControl {},
  };
});

vi.mock("maplibre-gl", () => ({
  Map: mapClass.Map,
  Marker: mapClass.Marker,
  Popup: mapClass.Popup,
  NavigationControl: mapClass.NavigationControl,
}));

const POSITRON_STYLE_URL = "https://tiles.openfreemap.org/styles/positron";

function assertPositronStyle(style: string | Record<string, unknown> | undefined) {
  expect(style).toBe(POSITRON_STYLE_URL);
}

beforeEach(() => {
  mapClass.Map.mockClear();
  mapClass.Marker.mockClear();
  mapClass.Popup.mockClear();
});

describe("MapView", () => {
  it("renders a labeled container with an OpenFreeMap Positron basemap", async () => {
    renderWithProviders(<MapView aria-label="Warehouse location" />);

    expect(screen.getByRole("region", { name: "Warehouse location" })).toBeInTheDocument();
    await waitFor(() => expect(mapClass.Map).toHaveBeenCalledTimes(1));

    const options = mapClass.Map.mock.lastCall?.[0];
    expect(options).toHaveProperty("style");
    assertPositronStyle(options?.style);
    expect(options?.center).toEqual([0, 0]);
    expect(options?.zoom).toBe(2);
  });

  it("renders markers with popups", async () => {
    renderWithProviders(
      <MapView
        center={{ lng: 106.8456, lat: -6.2088 }}
        zoom={12}
        markers={[
          { lng: 106.8456, lat: -6.2088, popup: "Main warehouse" },
          { lng: 106.8, lat: -6.2 },
        ]}
      />,
    );

    await waitFor(() => expect(mapClass.Marker).toHaveBeenCalledTimes(2));
    expect(mapClass.Marker).toHaveBeenCalledWith();
    expect(mapClass.Popup).toHaveBeenCalledTimes(1);
  });
});
