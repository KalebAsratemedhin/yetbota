"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { MapPin, X } from "lucide-react";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useIsLg } from "@/hooks/use-is-lg";
import {
  POI_MIN_ZOOM,
  fetchMapPois,
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

function MapBody({ lat, lng }: { lat: number; lng: number }) {
  const [pois, setPois] = useState<MapPoi[]>([]);
  const [view, setView] = useState<MapView | null>(null);
  const debouncedView = useDebouncedValue(view, 550);
  const viewKeyRef = useRef("");

  const onViewChange = useCallback((next: MapView) => {
    setView(next);
  }, []);

  useEffect(() => {
    if (!debouncedView || debouncedView.zoom < POI_MIN_ZOOM) {
      setPois([]);
      return;
    }
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
        if (!cancelled) setPois(rows);
      })
      .catch(() => {
        if (!cancelled) setPois([]);
      });
    return () => {
      cancelled = true;
    };
  }, [debouncedView]);

  return (
    <div className="relative isolate h-full min-h-0 w-full overflow-hidden">
      <div className="absolute inset-0">
        <LocationMap
          lat={lat}
          lng={lng}
          pinned
          interactive={false}
          pois={pois}
          onViewChange={onViewChange}
        />
      </div>
    </div>
  );
}

function PanelChrome({
  onClose,
  children,
  className,
  grabber,
  showClose,
}: {
  onClose?: () => void;
  children: React.ReactNode;
  className?: string;
  grabber?: boolean;
  showClose?: boolean;
}) {
  return (
    <div className={cn("flex h-full min-h-0 flex-col bg-surface", className)}>
      {grabber ? (
        <div className="flex shrink-0 justify-center pt-2" aria-hidden>
          <span className="h-1 w-10 rounded-full bg-border-subtle" />
        </div>
      ) : null}
      <header className="flex shrink-0 items-center justify-between gap-3 px-4 py-2.5">
        <h2 className="inline-flex items-center gap-1.5 text-sm font-semibold text-fg">
          <MapPin className="h-3.5 w-3.5 text-brand" />
          Map
        </h2>
        {showClose && onClose ? (
          <button
            type="button"
            onClick={onClose}
            className="rounded-lg p-1.5 text-fg-muted hover:bg-surface-2 hover:text-fg"
            aria-label="Close map"
          >
            <X className="h-4 w-4" />
          </button>
        ) : (
          <span className="w-8" aria-hidden />
        )}
      </header>
      <div className="min-h-0 flex-1">{children}</div>
    </div>
  );
}

/**
 * Same side-slot as comments: desktop sticky column, mobile bottom sheet.
 * Only one map instance mounts (Leaflet breaks in zero-size hidden containers).
 */
export default function PostMapPanel({
  lat,
  lng,
  mobileOpen,
  onMobileClose,
  desktopOpen,
}: {
  lat: number;
  lng: number;
  mobileOpen: boolean;
  onMobileClose: () => void;
  desktopOpen: boolean;
}) {
  const isLg = useIsLg();

  useEffect(() => {
    if (!mobileOpen || isLg) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onMobileClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [mobileOpen, onMobileClose, isLg]);

  if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;

  const map = <MapBody key={`${lat},${lng}`} lat={lat} lng={lng} />;

  if (isLg) {
    if (!desktopOpen) return null;
    return (
      <aside
        className="sticky top-24 flex h-[calc(100dvh-6.5rem)] w-[28rem] min-w-[28rem] max-w-[28rem] shrink-0 flex-col overflow-hidden rounded-2xl border border-border-subtle"
        aria-label="Map"
      >
        <PanelChrome className="h-full min-h-0 w-full min-w-0 overflow-hidden">
          {map}
        </PanelChrome>
      </aside>
    );
  }

  if (!mobileOpen) return null;

  return (
    <div
      className="fixed inset-x-0 bottom-0 z-40 flex h-[min(52vh,28rem)] flex-col overflow-hidden rounded-t-2xl border border-b-0 border-border-subtle shadow-[0_-12px_40px_-20px_rgba(0,0,0,0.45)]"
      role="dialog"
      aria-modal="false"
      aria-label="Map"
    >
      <PanelChrome onClose={onMobileClose} grabber showClose className="h-full">
        {map}
      </PanelChrome>
    </div>
  );
}
