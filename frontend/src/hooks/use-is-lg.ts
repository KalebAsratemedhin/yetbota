"use client";

import { useSyncExternalStore } from "react";

const LG_QUERY = "(min-width: 1024px)";

function subscribe(onChange: () => void) {
  const mq = window.matchMedia(LG_QUERY);
  mq.addEventListener("change", onChange);
  return () => mq.removeEventListener("change", onChange);
}

function getSnapshot() {
  return window.matchMedia(LG_QUERY).matches;
}

function getServerSnapshot() {
  return false;
}

/** True at Tailwind `lg` and up. */
export function useIsLg() {
  return useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);
}
