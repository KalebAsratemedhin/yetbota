"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { Loader2, Navigation, Search, X } from "lucide-react";
import { toast } from "sonner";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import {
  DEFAULT_CENTER,
  POI_MIN_ZOOM,
  fetchMapPois,
  looksLikeGoogleMapsLink,
  resolveGoogleMapsLink,
  reversePlace,
  searchPlaces,
  type GeoHit,
  type MapPoi,
  type MapView,
} from "@/lib/geo/nominatim";
import { cn } from "@/lib/utils";

const LocationMap = dynamic(() => import("@/components/compose/LocationMap"), {
  ssr: false,
  loading: () => (
    <div className="flex h-full items-center justify-center bg-surface-2 text-sm text-fg-faint">
      Loading map…
    </div>
  ),
});

export default function LocationPicker({
  lat,
  lng,
  onChange,
}: {
  lat: number | null;
  lng: number | null;
  onChange: (lat: number, lng: number, label?: string) => void;
}) {
  const [query, setQuery] = useState("");
  const [hits, setHits] = useState<GeoHit[]>([]);
  const [resolvedFor, setResolvedFor] = useState("");
  const [label, setLabel] = useState<string | null>(null);
  const [locating, setLocating] = useState(false);
  const [resolvingLink, setResolvingLink] = useState(false);
  const [searchOpen, setSearchOpen] = useState(true);
  const [pois, setPois] = useState<MapPoi[]>([]);
  const [view, setView] = useState<MapView | null>(null);
  const [poisLoading, setPoisLoading] = useState(false);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const mapsHandledRef = useRef<string | null>(null);
  const debounced = useDebouncedValue(query.trim(), 400);
  const debouncedView = useDebouncedValue(view, 550);
  const viewKeyRef = useRef("");

  const pinLat = lat ?? DEFAULT_CENTER.lat;
  const pinLng = lng ?? DEFAULT_CENTER.lng;
  const pinned =
    lat != null && lng != null && Number.isFinite(lat) && Number.isFinite(lng);

  const bias = {
    lat: view?.centerLat ?? pinLat,
    lng: view?.centerLng ?? pinLng,
  };

  useEffect(() => {
    if (!searchOpen || debounced.length < 2) return;

    if (looksLikeGoogleMapsLink(debounced)) {
      if (mapsHandledRef.current === debounced) {
        setResolvedFor(debounced);
        return;
      }
      let cancelled = false;
      setResolvingLink(true);
      setHits([]);
      void resolveGoogleMapsLink(debounced)
        .then(async (coords) => {
          if (cancelled) return;
          mapsHandledRef.current = debounced;
          await pick(coords.lat, coords.lng, coords.label);
          toast.success("Location set from Maps link");
        })
        .catch((err: unknown) => {
          if (cancelled) return;
          toast.error("Could not use Maps link", {
            description: err instanceof Error ? err.message : undefined,
          });
          setResolvedFor(debounced);
        })
        .finally(() => {
          if (!cancelled) setResolvingLink(false);
        });
      return () => {
        cancelled = true;
      };
    }

    let cancelled = false;
    searchPlaces(debounced, bias)
      .then((rows) => {
        if (cancelled) return;
        setHits(rows);
        setResolvedFor(debounced);
      })
      .catch(() => {
        if (cancelled) return;
        setHits([]);
        setResolvedFor(debounced);
      });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- search on query; bias/pick read at fire time
  }, [debounced, searchOpen]);

  useEffect(() => {
    if (!debouncedView || debouncedView.zoom < POI_MIN_ZOOM) return;

    const key = [
      debouncedView.zoom,
      debouncedView.south.toFixed(3),
      debouncedView.west.toFixed(3),
      debouncedView.north.toFixed(3),
      debouncedView.east.toFixed(3),
    ].join("|");
    if (key === viewKeyRef.current) return;
    viewKeyRef.current = key;

    let cancelled = false;
    void fetchMapPois(debouncedView)
      .then((rows) => {
        if (cancelled) return;
        setPois(rows);
        setPoisLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setPois([]);
        setPoisLoading(false);
      });
    queueMicrotask(() => {
      if (!cancelled) setPoisLoading(true);
    });
    return () => {
      cancelled = true;
    };
  }, [debouncedView]);

  const visiblePois =
    debouncedView && debouncedView.zoom >= POI_MIN_ZOOM ? pois : [];
  const showPoiStatus = Boolean(view && view.zoom >= POI_MIN_ZOOM);

  const onViewChange = useCallback((next: MapView) => {
    setView(next);
  }, []);

  const shownHits = searchOpen && debounced.length >= 2 ? hits : [];
  const searching =
    searchOpen &&
    debounced.length >= 2 &&
    (resolvingLink ||
      (!looksLikeGoogleMapsLink(debounced) && resolvedFor !== debounced));

  async function applyMapsLink(url: string) {
    const trimmed = url.trim();
    mapsHandledRef.current = trimmed;
    setResolvingLink(true);
    setHits([]);
    try {
      const coords = await resolveGoogleMapsLink(trimmed);
      await pick(coords.lat, coords.lng, coords.label);
      toast.success("Location set from Maps link");
    } catch (err: unknown) {
      toast.error("Could not use Maps link", {
        description: err instanceof Error ? err.message : undefined,
      });
      setResolvedFor(trimmed);
    } finally {
      setResolvingLink(false);
    }
  }

  function closeSearch() {
    setSearchOpen(false);
    setQuery("");
    setHits([]);
    setResolvedFor("");
  }

  function openSearch() {
    setSearchOpen(true);
    queueMicrotask(() => searchInputRef.current?.focus());
  }

  async function pick(nextLat: number, nextLng: number, knownLabel?: string) {
    if (!Number.isFinite(nextLat) || !Number.isFinite(nextLng)) {
      toast.error("Could not read that place’s coordinates");
      return;
    }
    onChange(nextLat, nextLng, knownLabel);
    if (knownLabel) {
      setLabel(knownLabel);
      setHits([]);
      setQuery("");
      setSearchOpen(false);
      return;
    }
    setLabel("Looking up address…");
    try {
      const name = await reversePlace(nextLat, nextLng);
      setLabel(name);
      onChange(nextLat, nextLng, name);
    } catch {
      setLabel(`${nextLat.toFixed(5)}, ${nextLng.toFixed(5)}`);
    }
  }

  function pickPoi(poi: MapPoi) {
    void pick(poi.lat, poi.lng, `${poi.name} (${poi.kind})`);
  }

  function useMyLocation() {
    if (!navigator.geolocation) {
      toast.error("Geolocation not available");
      return;
    }
    setLocating(true);
    navigator.geolocation.getCurrentPosition(
      (pos) => {
        setLocating(false);
        void pick(pos.coords.latitude, pos.coords.longitude);
      },
      () => {
        setLocating(false);
        toast.error("Could not read location");
      },
      { enableHighAccuracy: true, timeout: 12000 }
    );
  }

  return (
    <div className="relative z-0 h-full min-h-[16rem] overflow-hidden">
      <LocationMap
        lat={pinLat}
        lng={pinLng}
        pinned={pinned}
        pois={visiblePois}
        onPick={(a, b) => void pick(a, b)}
        onPickPoi={pickPoi}
        onViewChange={onViewChange}
      />

      <div className="pointer-events-none absolute inset-x-0 top-0 z-10 p-3 sm:p-4">
        <div
          className="pointer-events-auto relative mx-auto max-w-md"
          onMouseDown={(e) => e.stopPropagation()}
          onDoubleClick={(e) => e.stopPropagation()}
          onClick={(e) => e.stopPropagation()}
        >
          {searchOpen ? (
            <>
              <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-white/50" />
              <input
                ref={searchInputRef}
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onPaste={(e) => {
                  const text = e.clipboardData.getData("text")?.trim() ?? "";
                  if (!looksLikeGoogleMapsLink(text)) return;
                  e.preventDefault();
                  setQuery(text);
                  void applyMapsLink(text);
                }}
                placeholder="Search or paste a Google Maps link…"
                className="w-full rounded-xl border border-white/15 bg-black/75 py-2.5 pr-16 pl-9 text-sm text-white outline-none backdrop-blur-sm placeholder:text-white/40 focus:border-brand/50"
              />
              <div className="absolute top-1/2 right-2.5 flex -translate-y-1/2 items-center gap-0.5">
                {searching || resolvingLink ? (
                  <Loader2 className="h-4 w-4 animate-spin text-brand" />
                ) : null}
                <button
                  type="button"
                  onClick={closeSearch}
                  aria-label="Close search"
                  className="rounded-md p-1 text-white/60 hover:bg-white/10 hover:text-white"
                >
                  <X className="h-4 w-4" />
                </button>
              </div>
              {shownHits.length > 0 ? (
                <ul className="absolute inset-x-0 top-[calc(100%+6px)] max-h-56 overflow-y-auto rounded-xl border border-white/15 bg-black/90 shadow-lg backdrop-blur-sm">
                  {shownHits.map((h) => (
                    <li key={`${h.lat},${h.lng}:${h.label}`}>
                      <button
                        type="button"
                        data-lat={h.lat}
                        data-lng={h.lng}
                        onClick={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          const el = e.currentTarget;
                          const nextLat = Number(el.dataset.lat);
                          const nextLng = Number(el.dataset.lng);
                          void pick(nextLat, nextLng, h.label);
                        }}
                        className="w-full px-3 py-2.5 text-left text-xs leading-snug text-white/90 hover:bg-white/10"
                      >
                        <span className="block">{h.label}</span>
                        {h.kind ? (
                          <span className="mt-0.5 block text-[10px] text-white/45 capitalize">
                            {h.kind.replace(/_/g, " ")}
                          </span>
                        ) : null}
                      </button>
                    </li>
                  ))}
                </ul>
              ) : null}
            </>
          ) : (
            <button
              type="button"
              onClick={openSearch}
              aria-label="Open place search"
              className="inline-flex items-center gap-2 rounded-full border border-white/15 bg-black/75 px-3 py-2 text-xs font-semibold text-white backdrop-blur-sm hover:bg-black/90"
            >
              <Search className="h-3.5 w-3.5" />
              Search
            </button>
          )}
        </div>
      </div>

      <div
        className="absolute right-3 bottom-3 z-10 flex flex-col items-end gap-2 sm:right-4 sm:bottom-4"
        onMouseDown={(e) => e.stopPropagation()}
        onClick={(e) => e.stopPropagation()}
      >
        <button
          type="button"
          onClick={useMyLocation}
          disabled={locating}
          className="inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-black/75 px-3 py-2 text-xs font-semibold text-white backdrop-blur-sm hover:bg-black/90 disabled:opacity-60"
        >
          <Navigation className={cn("h-3.5 w-3.5", locating && "animate-pulse")} />
          {locating ? "Locating…" : "My location"}
        </button>
        <div className="max-w-[15rem] rounded-xl border border-white/15 bg-black/75 px-3 py-2 text-right backdrop-blur-sm">
          <p className="text-[10px] font-semibold tracking-wide text-white/50 uppercase">
            {pinned ? "Location" : "No pin yet"}
          </p>
          <p className="mt-0.5 line-clamp-2 text-[11px] text-white/90">
            {label ??
              (pinned
                ? `${lat!.toFixed(5)}, ${lng!.toFixed(5)}`
                : "Search, paste Maps link, or tap the map")}
          </p>
          {view && view.zoom < POI_MIN_ZOOM ? (
            <p className="mt-1 text-[10px] text-white/45">Zoom in to see places</p>
          ) : showPoiStatus && poisLoading ? (
            <p className="mt-1 text-[10px] text-white/45">Loading places…</p>
          ) : showPoiStatus && visiblePois.length > 0 ? (
            <p className="mt-1 text-[10px] text-white/45">
              {visiblePois.length} places nearby
            </p>
          ) : null}
        </div>
      </div>
    </div>
  );
}
