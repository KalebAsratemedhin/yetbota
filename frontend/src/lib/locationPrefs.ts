const PREF_KEY = "yetbota.location";
const PROMPT_KEY = "yetbota.location.prompt";
export const LOCATION_PREF_EVENT = "yetbota:location-pref-change";

export type LocationPref = "on" | "off";

export function readLocationPref(): LocationPref {
  if (typeof window === "undefined") return "on";
  return window.localStorage.getItem(PREF_KEY) === "off" ? "off" : "on";
}

export function writeLocationPref(pref: LocationPref) {
  window.localStorage.setItem(PREF_KEY, pref);
  window.dispatchEvent(new Event(LOCATION_PREF_EVENT));
}

/** One-shot: show the feed location toast after the next home load (post sign-in). */
export function markLocationPromptAfterAuth() {
  if (typeof window === "undefined") return;
  window.sessionStorage.setItem(PROMPT_KEY, "1");
}

export function consumeLocationPrompt(): boolean {
  if (typeof window === "undefined") return false;
  if (window.sessionStorage.getItem(PROMPT_KEY) !== "1") return false;
  window.sessionStorage.removeItem(PROMPT_KEY);
  return true;
}
