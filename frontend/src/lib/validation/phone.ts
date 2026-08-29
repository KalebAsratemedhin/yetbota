/** Empty allowed. Non-empty must be 7–15 characters (matches identity-service). */
export const PHONE_MIN_LEN = 7;
export const PHONE_MAX_LEN = 15;

export function isValidOptionalPhone(raw: string): boolean {
  const trimmed = raw.trim();
  if (!trimmed) return true;
  const n = [...trimmed].length;
  return n >= PHONE_MIN_LEN && n <= PHONE_MAX_LEN;
}

/** Trim only — no E.164 normalization. */
export function normalizePhone(raw: string): string {
  return raw.trim();
}

export const PHONE_INVALID_MESSAGE =
  "Phone must be between 7 and 15 characters.";
