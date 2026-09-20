"use client";

import type { Map as MaplibreMap } from "maplibre-gl";
import { useEffect, useRef } from "react";

import { cn } from "@/lib/utils";

const OPENFREEMAP_POSITRON_STYLE = "https://tiles.openfreemap.org/styles/positron";
const EMPTY_MARKERS: MapMarker[] = [];

export type MapMarker = {
  lng: number;
  lat: number;
  popup?: string;
};

export type MapViewProps = {
  center?: { lng: number; lat: number };
  zoom?: number;
  markers?: MapMarker[];
  className?: string;
  "aria-label"?: string;
  onMapReady?: (map: MaplibreMap) => void;
};

const DEFAULT_CENTER = { lng: 0, lat: 0 };
const DEFAULT_ZOOM = 2;

export function MapView({
  center = DEFAULT_CENTER,
  zoom = DEFAULT_ZOOM,
  markers = EMPTY_MARKERS,
  className,
  "aria-label": ariaLabel = "Map",
  onMapReady,
}: MapViewProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<MaplibreMap | null>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    let disposed = false;
    let map: MaplibreMap | null = null;

    async function init() {
      try {
        await import("maplibre-gl/dist/maplibre-gl.css");
        const { Map: MapLibre, Marker, NavigationControl, Popup } = await import("maplibre-gl");
        if (disposed || !containerRef.current) {
          return;
        }

        map = new MapLibre({
          container: containerRef.current,
          style: OPENFREEMAP_POSITRON_STYLE,
          center: [center.lng, center.lat],
          zoom,
        });
        map.addControl(new NavigationControl(), "bottom-right");
        mapRef.current = map;

        for (const marker of markers) {
          const pin = new Marker().setLngLat([marker.lng, marker.lat]);
          if (marker.popup) {
            pin.setPopup(new Popup({ offset: 25 }).setText(marker.popup));
          }
          pin.addTo(map);
        }

        onMapReady?.(map);
      } catch {
        // map failed to load — silent fallback
      }
    }

    void init();

    return () => {
      disposed = true;
      map?.remove();
      mapRef.current = null;
    };
  }, [center, zoom, markers, onMapReady]);

  return (
    <section
      data-slot="map-view"
      aria-label={ariaLabel}
      className={cn(
        "relative h-[480px] w-full overflow-hidden rounded-xl bg-muted border border-border",
        className,
      )}
    >
      <div ref={containerRef} className="absolute inset-0 size-full" />
      {markers.length > 0 ? (
        <div className="sr-only">
          <h2>{ariaLabel}</h2>
          <ul>
            {markers.map((marker, index) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: marker list has no stable id
              <li key={index}>
                {marker.popup
                  ? `${marker.popup} (${marker.lat.toFixed(4)}, ${marker.lng.toFixed(4)})`
                  : `Marker ${index + 1} (${marker.lat.toFixed(4)}, ${marker.lng.toFixed(4)})`}
              </li>
            ))}
          </ul>
        </div>
      ) : null}
    </section>
  );
}
