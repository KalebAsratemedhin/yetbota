import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { ApiErrorBody, ApiErrorPayload } from "@/lib/auth/types";

function isApiErrorPayload(value: unknown): value is ApiErrorPayload {
  return (
    !!value &&
    typeof value === "object" &&
    "code" in value &&
    "message" in value &&
    typeof (value as ApiErrorPayload).code === "string" &&
    typeof (value as ApiErrorPayload).message === "string"
  );
}

/** Extract OpenAPI-shaped `{ error: { code, message, details? } }` from an RTK error. */
export function getApiErrorPayload(error: unknown): ApiErrorPayload | null {
  if (!error || typeof error !== "object" || !("status" in error)) return null;
  const data = (error as FetchBaseQueryError).data;
  if (!data || typeof data !== "object" || !("error" in data)) return null;

  const body = data as ApiErrorBody | { error: string };
  if (typeof body.error === "string") {
    return { code: "INTERNAL_ERROR", message: body.error };
  }
  if (isApiErrorPayload(body.error)) return body.error;
  return null;
}

export function getApiErrorMessage(error: unknown, fallback = "Something went wrong"): string {
  const payload = getApiErrorPayload(error);
  if (payload) {
    const detail = payload.details?.[0]?.message;
    if (detail) return detail;
    return payload.message || fallback;
  }

  if (!error || typeof error !== "object") return fallback;

  const err = error as FetchBaseQueryError;
  if ("status" in err) {
    if (err.status === "FETCH_ERROR") {
      return "Cannot reach the server. Is the API running?";
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
