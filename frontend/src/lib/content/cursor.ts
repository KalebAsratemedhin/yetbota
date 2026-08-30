/**
 * Build the same keyset cursor content-service expects:
 * base64url(`${unixNano}|${uuid}`) — posts older than this (newest-first lists).
 */
export function encodeContentCursor(createdAtIso: string, id: string): string {
  const ms = Date.parse(createdAtIso);
  if (!Number.isFinite(ms)) {
    throw new Error("Invalid created_at for cursor");
  }
  const ns = ms * 1_000_000;
  const raw = `${ns}|${id}`;
  const bytes = new TextEncoder().encode(raw);
  let bin = "";
  for (let i = 0; i < bytes.length; i++) {
    bin += String.fromCharCode(bytes[i]!);
  }
  return btoa(bin).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}
