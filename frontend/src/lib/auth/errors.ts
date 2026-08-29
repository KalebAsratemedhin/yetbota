import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";

export function getApiErrorMessage(error: unknown, fallback = "Something went wrong"): string {
  if (!error || typeof error !== "object") return fallback;

  const err = error as FetchBaseQueryError;
  if ("status" in err) {
    const data = err.data;
    if (data && typeof data === "object" && "error" in data && typeof data.error === "string") {
      return data.error;
    }
    if (err.status === "FETCH_ERROR") {
      return "Cannot reach the server. Is identity-service running?";
    }
    if (err.status === 401) return "Invalid username or password";
    if (err.status === 409) return "Username is already taken";
  }

  return fallback;
}

/** Only honor relative paths so ?next= cannot open-redirect. */
export function safeNextPath(raw: string | null): string {
  if (!raw) return "/";
  if (!raw.startsWith("/") || raw.startsWith("//")) return "/";
  return raw;
}
