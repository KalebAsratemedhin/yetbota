"use client";

import { useCallback, useRef } from "react";
import {
  clampClipRange,
  formatClipTime,
  rangeDuration,
  type ClipRange,
} from "@/lib/media";
import { cn } from "@/lib/utils";

type DragMode = "start" | "end" | "window" | null;

/**
 * Dual-handle clip window over a timeline (Instagram/TikTok-style).
 * Presentational — parent owns range state.
 */
export default function ClipTimeline({
  durationSec,
  range,
  maxSec,
  playheadSec,
  peaks,
  onChange,
  onCommit,
  className,
}: {
  durationSec: number;
  range: ClipRange;
  maxSec: number;
  playheadSec?: number;
  /** Optional normalized 0–1 waveform peaks (audio). */
  peaks?: number[];
  onChange: (next: ClipRange) => void;
  /** Fired when a drag ends — use this to seek preview, not `onChange`. */
  onCommit?: (next: ClipRange) => void;
  className?: string;
}) {
  const trackRef = useRef<HTMLDivElement>(null);
  const dragRef = useRef<{
    mode: DragMode;
    originX: number;
    originRange: ClipRange;
  } | null>(null);

  const pct = useCallback(
    (sec: number) => {
      if (durationSec <= 0) return 0;
      return (sec / durationSec) * 100;
    },
    [durationSec]
  );

  const secFromClientX = useCallback(
    (clientX: number) => {
      const el = trackRef.current;
      if (!el || durationSec <= 0) return 0;
      const rect = el.getBoundingClientRect();
      const t = (clientX - rect.left) / rect.width;
      return Math.min(durationSec, Math.max(0, t * durationSec));
    },
    [durationSec]
  );

  function startDrag(mode: Exclude<DragMode, null>, e: React.PointerEvent) {
    e.preventDefault();
    e.stopPropagation();
    const track = trackRef.current;
    if (track) {
      try {
        track.setPointerCapture(e.pointerId);
      } catch {
        /* ignore */
      }
    }
    dragRef.current = {
      mode,
      originX: e.clientX,
      originRange: range,
    };
  }

  function onPointerMove(e: React.PointerEvent) {
    const drag = dragRef.current;
    if (!drag?.mode) return;
    const el = trackRef.current;
    if (!el) return;
    const dx = e.clientX - drag.originX;
    const dSec = (dx / el.getBoundingClientRect().width) * durationSec;

    if (drag.mode === "window") {
      const len = rangeDuration(drag.originRange);
      onChange(
        clampClipRange(
          {
            start: drag.originRange.start + dSec,
            end: drag.originRange.start + dSec + len,
          },
          durationSec,
          maxSec,
          "window"
        )
      );
      return;
    }

    if (drag.mode === "start") {
      onChange(
        clampClipRange(
          {
            start: drag.originRange.start + dSec,
            end: drag.originRange.end,
          },
          durationSec,
          maxSec,
          "start"
        )
      );
      return;
    }

    onChange(
      clampClipRange(
        {
          start: drag.originRange.start,
          end: drag.originRange.end + dSec,
        },
        durationSec,
        maxSec,
        "end"
      )
    );
  }

  function endDrag(e: React.PointerEvent) {
    if (!dragRef.current) return;
    const track = trackRef.current;
    try {
      track?.releasePointerCapture(e.pointerId);
    } catch {
      /* ignore */
    }
    dragRef.current = null;
    onCommit?.(range);
  }

  const left = pct(range.start);
  const width = pct(range.end - range.start);
  const playhead =
    playheadSec != null && durationSec > 0 ? pct(playheadSec) : null;

  return (
    <div className={cn("select-none", className)}>
      <div className="mb-1.5 flex items-center justify-between text-[11px] text-fg-faint">
        <span>{formatClipTime(range.start)}</span>
        <span className="font-medium text-fg-muted">
          {formatClipTime(rangeDuration(range))} / {maxSec}s max
        </span>
        <span>{formatClipTime(range.end)}</span>
      </div>

      <div
        ref={trackRef}
        className="relative h-14 touch-none rounded-xl bg-surface-2"
        onPointerMove={onPointerMove}
        onPointerUp={endDrag}
        onPointerCancel={endDrag}
      >
        {/* Waveform or filmstrip placeholder */}
        <div className="absolute inset-0 flex items-end gap-px overflow-hidden rounded-xl px-1 py-2 opacity-50">
          {(peaks && peaks.length > 0
            ? peaks
            : Array.from({ length: 48 }, (_, i) => 0.25 + ((i * 17) % 10) / 20)
          ).map((p, i) => (
            <div
              key={i}
              className="min-w-0 flex-1 rounded-sm bg-fg-faint/80"
              style={{ height: `${Math.max(12, p * 100)}%` }}
            />
          ))}
        </div>

        {/* Dim outside selection */}
        <div
          className="pointer-events-none absolute inset-y-0 left-0 rounded-l-xl bg-black/55"
          style={{ width: `${left}%` }}
        />
        <div
          className="pointer-events-none absolute inset-y-0 right-0 rounded-r-xl bg-black/55"
          style={{ width: `${Math.max(0, 100 - left - width)}%` }}
        />

        {/* Selection window */}
        <div
          role="slider"
          aria-valuemin={0}
          aria-valuemax={durationSec}
          aria-valuenow={range.start}
          tabIndex={0}
          className="absolute inset-y-0 cursor-grab rounded-md border-2 border-brand active:cursor-grabbing"
          style={{ left: `${left}%`, width: `${width}%` }}
          onPointerDown={(e) => startDrag("window", e)}
          onKeyDown={(e) => {
            const step = e.shiftKey ? 1 : 0.1;
            if (e.key === "ArrowLeft") {
              e.preventDefault();
              const next = clampClipRange(
                { start: range.start - step, end: range.end - step },
                durationSec,
                maxSec,
                "window"
              );
              onChange(next);
              onCommit?.(next);
            }
            if (e.key === "ArrowRight") {
              e.preventDefault();
              const next = clampClipRange(
                { start: range.start + step, end: range.end + step },
                durationSec,
                maxSec,
                "window"
              );
              onChange(next);
              onCommit?.(next);
            }
          }}
        >
          <button
            type="button"
            aria-label="Trim start"
            className="absolute top-0 bottom-0 left-0 z-10 w-3 -translate-x-1/2 cursor-ew-resize rounded-full bg-brand"
            onPointerDown={(e) => startDrag("start", e)}
          />
          <button
            type="button"
            aria-label="Trim end"
            className="absolute top-0 right-0 bottom-0 z-10 w-3 translate-x-1/2 cursor-ew-resize rounded-full bg-brand"
            onPointerDown={(e) => startDrag("end", e)}
          />
        </div>

        {playhead != null ? (
          <div
            className="pointer-events-none absolute top-0 bottom-0 z-20 w-0.5 bg-white shadow"
            style={{ left: `${playhead}%` }}
          />
        ) : null}

        {/* Click track to jump window */}
        <button
          type="button"
          aria-label="Move clip window"
          className="absolute inset-0 z-0"
          onPointerDown={(e) => {
            if ((e.target as HTMLElement).closest('[role="slider"]')) return;
            const center = secFromClientX(e.clientX);
            const len = rangeDuration(range);
            onChange(
              clampClipRange(
                { start: center - len / 2, end: center + len / 2 },
                durationSec,
                maxSec,
                "window"
              )
            );
            startDrag("window", e);
          }}
        />
      </div>
    </div>
  );
}
