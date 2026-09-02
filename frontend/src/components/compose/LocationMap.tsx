"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import L from "leaflet";
import {
  MapContainer,
  Marker,
  TileLayer,
  Tooltip,
  ZoomControl,
  useMap,
  useMapEvents,
} from "react-leaflet";
import type { MapPoi, MapView } from "@/lib/geo/nominatim";
import "leaflet/dist/leaflet.css";

const CYCLOSM_URL =
  "https://{s}.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png";
const OSM_FALLBACK_URL = "https://tile.openstreetmap.de/{z}/{x}/{y}.png";

const pin = L.divIcon({
  className: "yb-map-pin",
  html: `<span class="yb-map-pin-inner"></span>`,
  iconSize: [32, 40],
  iconAnchor: [16, 38],
});

function poiIcon(labeled: boolean) {
  return L.divIcon({
    className: "yb-map-poi",
    html: labeled
      ? `<span class="yb-map-poi-dot"></span>`
      : `<span class="yb-map-poi-dot"></span>`,
    iconSize: [14, 14],
    iconAnchor: [7, 7],
  });
}

function viewFromMap(map: L.Map): MapView | null {
  const b = map.getBounds();
  const c = map.getCenter();
  if (
    !Number.isFinite(c.lat) ||
    !Number.isFinite(c.lng) ||
    !Number.isFinite(b.getSouth())
  ) {
    return null;
  }
  return {
    south: b.getSouth(),
    west: b.getWest(),
    north: b.getNorth(),
    east: b.getEast(),
    zoom: map.getZoom(),
    centerLat: c.lat,
    centerLng: c.lng,
  };
}

function MapEvents({
  onPick,
  onViewChange,
}: {
  onPick?: (lat: number, lng: number) => void;
  onViewChange?: (view: MapView) => void;
}) {
  const map = useMapEvents({
    click(e) {
      onPick?.(e.latlng.lat, e.latlng.lng);
    },
    moveend() {
      const view = viewFromMap(map);
      if (view) onViewChange?.(view);
    },
    zoomend() {
      const view = viewFromMap(map);
      if (view) onViewChange?.(view);
    },
  });

  useEffect(() => {
    const id = window.setTimeout(() => {
      const view = viewFromMap(map);
      if (view) onViewChange?.(view);
    }, 0);
    return () => window.clearTimeout(id);
  }, [map, onViewChange]);

  return null;
}

function MapSizeFix() {
  const map = useMap();
  useEffect(() => {
    const el = map.getContainer();
    const refresh = () => {
      try {
        map.invalidateSize({ pan: false });
      } catch {
        /* not ready */
      }
    };
    refresh();
    const t1 = window.setTimeout(refresh, 50);
    const t2 = window.setTimeout(refresh, 250);
    const ro = new ResizeObserver(refresh);
    ro.observe(el);
    return () => {
      window.clearTimeout(t1);
      window.clearTimeout(t2);
      ro.disconnect();
    };
  }, [map]);
  return null;
}

function Recenter({ lat, lng }: { lat: number; lng: number }) {
  const map = useMap();
  useEffect(() => {
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;
    const rawZoom = map.getZoom();
    const zoom = Number.isFinite(rawZoom) ? Math.max(rawZoom, 16) : 16;
    try {
      map.invalidateSize({ pan: false });
      map.setView([lat, lng], zoom, { animate: false });
    } catch {
      /* Map not ready yet */
    }
  }, [lat, lng, map]);
  return null;
}

/** CyclOSM (dense amenities) with OSM.de fallback if tiles fail. */
function BasemapTiles() {
  const [fallback, setFallback] = useState(false);
  const errors = useRef(0);

  if (fallback) {
    return (
      <TileLayer
        key="osmde"
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url={OSM_FALLBACK_URL}
        maxZoom={19}
      />
    );
  }

  return (
    <TileLayer
      key="cyclosm"
      attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> · <a href="https://www.cyclosm.org">CyclOSM</a>'
      url={CYCLOSM_URL}
      maxZoom={20}
      eventHandlers={{
        tileerror: () => {
          errors.current += 1;
          if (errors.current >= 4) setFallback(true);
        },
      }}
    />
  );
}

export default function LocationMap({
  lat,
  lng,
  pinned,
  pois = [],
  onPick,
  onPickPoi,
  onViewChange,
  interactive = true,
}: {
  lat: number;
  lng: number;
  pinned: boolean;
  pois?: MapPoi[];
  onPick?: (lat: number, lng: number) => void;
  onPickPoi?: (poi: MapPoi) => void;
  onViewChange?: (view: MapView) => void;
  /** When false, map is view-only (no pin picking). */
  interactive?: boolean;
}) {
  const safeLat = Number.isFinite(lat) ? lat : 9.0192;
  const safeLng = Number.isFinite(lng) ? lng : 38.7525;
  const showPin = pinned && Number.isFinite(lat) && Number.isFinite(lng);
  const icon = useMemo(() => poiIcon(false), []);

  return (
    <MapContainer
      center={[safeLat, safeLng]}
      zoom={15}
      scrollWheelZoom
      className="h-full w-full"
      zoomControl={false}
    >
      <BasemapTiles />
      <ZoomControl position="bottomleft" />
      <MapSizeFix />
      <MapEvents
        onPick={interactive ? onPick : undefined}
        onViewChange={onViewChange}
      />
      {showPin ? <Recenter lat={lat} lng={lng} /> : null}
      {pois.map((p) => (
        <Marker
          key={p.id}
          position={[p.lat, p.lng]}
          icon={icon}
          eventHandlers={{
            click: (e) => {
              L.DomEvent.stopPropagation(e.originalEvent);
              onPickPoi?.(p);
            },
          }}
        >
          <Tooltip direction="top" offset={[0, -6]} opacity={0.95}>
            <span className="text-xs font-medium">
              {p.name}
              <span className="text-[10px] opacity-60"> · {p.kind}</span>
            </span>
          </Tooltip>
        </Marker>
      ))}
      {showPin ? <Marker position={[lat, lng]} icon={pin} zIndexOffset={1000} /> : null}
    </MapContainer>
  );
}
