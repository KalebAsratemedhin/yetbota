import type { ClipRange } from "@/lib/media/types";

export function rangeDuration(range: ClipRange): number {
  return Math.max(0, range.end - range.start);
}

/** Initial selection: first `maxSec` of the source (or full file if shorter). */
export function initialClipRange(durationSec: number, maxSec: number): ClipRange {
  const dur = Math.max(0, durationSec);
  const window = Math.min(maxSec, dur || maxSec);
  return { start: 0, end: window };
}

/**
 * Clamp a proposed range so it stays inside [0, duration] and ≤ maxSec.
 * Prefer keeping the dragged edge; shift the other if needed.
 */
export function clampClipRange(
  next: ClipRange,
  durationSec: number,
  maxSec: number,
  prefer: "start" | "end" | "window" = "window"
): ClipRange {
  const dur = Math.max(0, durationSec);
  const maxWin = Math.min(maxSec, dur || maxSec);
  let start = next.start;
  let end = next.end;

  if (prefer === "start") {
    start = clamp(start, 0, Math.max(0, dur - 0.1));
    end = Math.min(dur, Math.max(start + 0.1, Math.min(start + maxWin, end)));
    if (end - start > maxWin) end = start + maxWin;
  } else if (prefer === "end") {
    end = clamp(end, 0.1, dur);
    start = Math.max(0, Math.min(end - 0.1, Math.max(end - maxWin, start)));
    if (end - start > maxWin) start = end - maxWin;
  } else {
    const len = clamp(end - start, 0.1, maxWin);
    start = clamp(start, 0, Math.max(0, dur - len));
    end = start + len;
    if (end > dur) {
      end = dur;
      start = Math.max(0, end - len);
    }
  }

  start = clamp(start, 0, dur);
  end = clamp(end, start + 0.05, dur);
  if (end - start > maxWin) {
    if (prefer === "end") start = end - maxWin;
    else end = start + maxWin;
  }
  return { start, end };
}

function clamp(n: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, n));
}

export function formatClipTime(sec: number): string {
  if (!Number.isFinite(sec) || sec < 0) return "0:00";
  const m = Math.floor(sec / 60);
  const s = Math.floor(sec % 60);
  const tenths = Math.floor((sec % 1) * 10);
  if (m > 0) return `${m}:${String(s).padStart(2, "0")}.${tenths}`;
  return `${s}.${tenths}s`;
}
