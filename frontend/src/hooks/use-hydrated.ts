"use client";

import { useSyncExternalStore } from "react";

const emptySubscribe = () => () => {};

/** `false` on the server and during hydration; `true` after the client commits. */
export function useHydrated() {
  return useSyncExternalStore(emptySubscribe, () => true, () => false);
}
