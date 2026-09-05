"use client";

import { useCallback, useSyncExternalStore } from "react";
import {
  LOCATION_PREF_EVENT,
  readLocationPref,
  writeLocationPref,
  type LocationPref,
} from "@/lib/locationPrefs";

function subscribe(onStoreChange: () => void) {
  window.addEventListener(LOCATION_PREF_EVENT, onStoreChange);
  window.addEventListener("storage", onStoreChange);
  return () => {
    window.removeEventListener(LOCATION_PREF_EVENT, onStoreChange);
    window.removeEventListener("storage", onStoreChange);
  };
}

export function useLocationPreference() {
  const pref = useSyncExternalStore(
    subscribe,
    readLocationPref,
    () => "on" as LocationPref
  );

  const setPref = useCallback((next: LocationPref) => {
    writeLocationPref(next);
  }, []);

  return {
    pref,
    enabled: pref === "on",
    setPref,
    setEnabled: (on: boolean) => setPref(on ? "on" : "off"),
  };
}
