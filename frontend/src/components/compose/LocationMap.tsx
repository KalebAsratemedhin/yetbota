"use client";

import { useEffect, useMemo } from "react";
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

function viewFromMap(map: L.Map): MapView {
  const b = map.getBounds();
  const c = map.getCenter();
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
  onPick: (lat: number, lng: number) => void;
  onViewChange?: (view: MapView) => void;
}) {
  const map = useMapEvents({
    click(e) {
      onPick(e.latlng.lat, e.latlng.lng);
    },
    moveend() {
      onViewChange?.(viewFromMap(map));
    },
    zoomend() {
      onViewChange?.(viewFromMap(map));
    },
  });

  useEffect(() => {
    const id = window.setTimeout(() => onViewChange?.(viewFromMap(map)), 0);
    return () => window.clearTimeout(id);
  }, [map, onViewChange]);

  return null;
}

function Recenter({ lat, lng }: { lat: number; lng: number }) {
  const map = useMap();
  useEffect(() => {
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) return;
    map.flyTo([lat, lng], Math.max(map.getZoom(), 16), { duration: 0.7 });
  }, [lat, lng, map]);
  return null;
}

export default function LocationMap({
  lat,
  lng,
  pinned,
  pois,
  onPick,
  onPickPoi,
  onViewChange,
}: {
  lat: number;
  lng: number;
  pinned: boolean;
  pois: MapPoi[];
  onPick: (lat: number, lng: number) => void;
  onPickPoi: (poi: MapPoi) => void;
  onViewChange: (view: MapView) => void;
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
      {/* CyclOSM: free OSM tiles with dense amenity labels; no API key */}
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> · <a href="https://www.cyclosm.org">CyclOSM</a>'
        url="https://{s}.tile-cyclosm.openstreetmap.fr/cyclosm/{z}/{x}/{y}.png"
        maxZoom={20}
      />
      <ZoomControl position="bottomleft" />
      <MapEvents onPick={onPick} onViewChange={onViewChange} />
      {showPin ? <Recenter lat={lat} lng={lng} /> : null}
      {pois.map((p) => (
        <Marker
          key={p.id}
          position={[p.lat, p.lng]}
          icon={icon}
          eventHandlers={{
            click: (e) => {
              L.DomEvent.stopPropagation(e.originalEvent);
              onPickPoi(p);
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
