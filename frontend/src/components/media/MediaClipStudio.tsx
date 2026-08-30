"use client";

import { useEffect, useId, useMemo, useRef, useState } from "react";
import { Loader2, Pause, Play, Scissors, X } from "lucide-react";
import ClipPreviewPlayer, {
  type ClipPreviewHandle,
} from "@/components/media/ClipPreviewPlayer";
import ClipTimeline from "@/components/media/ClipTimeline";
import {
  DEFAULT_MAX_CLIP_SEC,
  buildWaveformPeaks,
  encodeClipOrPassthrough,
  formatClipTime,
  initialClipRange,
  readMediaDuration,
  type ClipKind,
  type ClipProgress,
  type ClipRange,
  type ClipResult,
} from "@/lib/media";
import { cn } from "@/lib/utils";

/**
 * Full-screen clip studio: preview + timeline + encode.
 * Compose (and any future surface) only opens this and receives a File.
 */
export default function MediaClipStudio({
  open,
  file,
  kind,
  maxDurationSec = DEFAULT_MAX_CLIP_SEC,
  onCancel,
  onConfirm,
}: {
  open: boolean;
  file: File | null;
  kind: ClipKind;
  maxDurationSec?: number;
  onCancel: () => void;
  onConfirm: (result: ClipResult) => void;
}) {
  if (!open || !file) return null;
  return (
    <MediaClipStudioSession
      key={`${file.name}:${file.size}:${file.lastModified}:${kind}`}
      file={file}
      kind={kind}
      maxDurationSec={maxDurationSec}
      onCancel={onCancel}
      onConfirm={onConfirm}
    />
  );
}

function MediaClipStudioSession({
  file,
  kind,
  maxDurationSec,
  onCancel,
  onConfirm,
}: {
  file: File;
  kind: ClipKind;
  maxDurationSec: number;
  onCancel: () => void;
  onConfirm: (result: ClipResult) => void;
}) {
  const titleId = useId();
  const previewRef = useRef<ClipPreviewHandle>(null);
  const src = useMemo(() => URL.createObjectURL(file), [file]);
  const [durationSec, setDurationSec] = useState(0);
  const [range, setRange] = useState<ClipRange>(() =>
    initialClipRange(maxDurationSec, maxDurationSec)
  );
  const [peaks, setPeaks] = useState<number[] | undefined>();
  const [playing, setPlaying] = useState(false);
  const [playhead, setPlayhead] = useState(0);
  const [metaReady, setMetaReady] = useState(false);
  const [encoding, setEncoding] = useState(false);
  const [progress, setProgress] = useState<ClipProgress | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [playHint, setPlayHint] = useState<string | null>(null);

  useEffect(() => {
    return () => URL.revokeObjectURL(src);
  }, [src]);

  useEffect(() => {
    let cancelled = false;
    void (async () => {
      try {
        const dur = await readMediaDuration(file);
        if (cancelled) return;
        setDurationSec(dur);
        setRange(initialClipRange(dur, maxDurationSec));
        setPlayhead(0);
        setMetaReady(true);
        if (kind === "audio") {
          const p = await buildWaveformPeaks(file, 56);
          if (!cancelled) setPeaks(p);
        }
      } catch {
        if (!cancelled) setError("Could not read this media file");
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [file, kind, maxDurationSec]);

  const encodingRef = useRef(encoding);
  encodingRef.current = encoding;
  const playingRef = useRef(playing);
  playingRef.current = playing;
  const onCancelRef = useRef(onCancel);
  onCancelRef.current = onCancel;

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape" && !encodingRef.current) onCancelRef.current();
      if (e.key === " " && !encodingRef.current) {
        e.preventDefault();
        void togglePlay();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  async function togglePlay() {
    const api = previewRef.current;
    if (!api) return;
    if (playingRef.current) {
      api.pause();
      setPlaying(false);
      setPlayHint(null);
      return;
    }
    try {
      setPlayHint(null);
      await api.play();
      setPlaying(true);
    } catch (err) {
      setPlaying(false);
      setPlayHint(
        err instanceof Error ? err.message : "Could not play preview"
      );
    }
  }

  async function handleUseClip() {
    setEncoding(true);
    setError(null);
    setPlaying(false);
    previewRef.current?.pause();
    try {
      const result = await encodeClipOrPassthrough(file, kind, range, durationSec, {
        maxDurationSec,
        onProgress: setProgress,
      });
      onConfirm(result);
    } catch (err) {
      if (err instanceof DOMException && err.name === "AbortError") return;
      const message =
        err instanceof Error ? err.message : "Could not create clip";
      setError(message);
    } finally {
      setEncoding(false);
    }
  }

  const loadingMeta = !metaReady && !error;
  const progressLabel =
    progress?.message ??
    (encoding ? "Working…" : null) ??
    (loadingMeta ? "Reading media…" : null);

  return (
    <div
      className="fixed inset-0 z-[80] flex items-end justify-center bg-black/70 sm:items-center sm:p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby={titleId}
    >
      <button
        type="button"
        className="absolute inset-0 cursor-default"
        aria-label="Close"
        disabled={encoding}
        onClick={() => {
          if (!encoding) onCancel();
        }}
      />

      <div
        className={cn(
          "relative z-10 flex max-h-[min(92vh,880px)] w-full max-w-lg flex-col overflow-hidden rounded-t-3xl border border-border-subtle bg-bg shadow-2xl sm:rounded-3xl"
        )}
      >
        <header className="flex items-center gap-3 border-b border-border-subtle px-4 py-3">
          <Scissors className="h-4 w-4 text-brand" />
          <div className="min-w-0 flex-1">
            <h2 id={titleId} className="text-sm font-semibold text-fg">
              {kind === "video" ? "Trim video" : "Trim audio"}
            </h2>
            <p className="truncate text-[11px] text-fg-faint">{file.name}</p>
          </div>
          <button
            type="button"
            disabled={encoding}
            onClick={onCancel}
            className="rounded-lg p-2 text-fg-muted hover:bg-surface-2 hover:text-fg disabled:opacity-40"
            aria-label="Cancel"
          >
            <X className="h-4 w-4" />
          </button>
        </header>

        <div className="yb-scrollbar min-h-0 flex-1 space-y-4 overflow-y-auto px-4 py-4">
          {!loadingMeta && !error ? (
            <ClipPreviewPlayer
              ref={previewRef}
              src={src}
              file={file}
              kind={kind}
              range={range}
              playing={playing}
              onPlayhead={setPlayhead}
              onPlayingChange={setPlaying}
            />
          ) : (
            <div className="flex aspect-video items-center justify-center rounded-2xl bg-surface-2 text-sm text-fg-faint">
              {error ?? "Loading…"}
            </div>
          )}

          <div className="flex items-center gap-2">
            <button
              type="button"
              disabled={loadingMeta || !!error || encoding}
              onClick={() => void togglePlay()}
              className="inline-flex h-10 w-10 items-center justify-center rounded-full bg-brand text-black disabled:opacity-40"
              aria-label={playing ? "Pause" : "Play"}
            >
              {playing ? <Pause className="h-4 w-4" /> : <Play className="ml-0.5 h-4 w-4" />}
            </button>
            <div className="min-w-0 flex-1 text-xs text-fg-muted">
              <p>
                Selection{" "}
                <span className="font-semibold text-fg">
                  {formatClipTime(range.end - range.start)}
                </span>
                {durationSec > 0 ? (
                  <span className="text-fg-faint"> · source {formatClipTime(durationSec)}</span>
                ) : null}
              </p>
              <p className="text-fg-faint">Drag handles to set in/out · Space to play</p>
              {playHint ? <p className="mt-1 text-red-400">{playHint}</p> : null}
            </div>
          </div>

          {!loadingMeta && durationSec > 0 ? (
            <ClipTimeline
              durationSec={durationSec}
              range={range}
              maxSec={maxDurationSec}
              playheadSec={playhead}
              peaks={peaks}
              onChange={(next) => setRange(next)}
              onCommit={(next) => {
                setPlayhead(next.start);
                previewRef.current?.seek(next.start);
              }}
            />
          ) : null}

          {error ? <p className="text-sm text-red-400">{error}</p> : null}
        </div>

        <footer className="border-t border-border-subtle px-4 py-3">
          {progressLabel ? (
            <div className="mb-3 flex items-center gap-2 text-xs text-fg-muted">
              <Loader2 className="h-3.5 w-3.5 animate-spin text-brand" />
              <span className="flex-1">{progressLabel}</span>
              {progress?.ratio != null ? (
                <span className="tabular-nums text-fg-faint">
                  {Math.round(progress.ratio * 100)}%
                </span>
              ) : null}
            </div>
          ) : null}
          {progress?.ratio != null && encoding ? (
            <div className="mb-3 h-1 overflow-hidden rounded-full bg-surface-2">
              <div
                className="h-full rounded-full bg-brand transition-[width]"
                style={{ width: `${Math.round(progress.ratio * 100)}%` }}
              />
            </div>
          ) : null}

          <div className="flex gap-2">
            <button
              type="button"
              disabled={encoding}
              onClick={onCancel}
              className="h-11 flex-1 rounded-xl border border-border-subtle text-sm font-medium text-fg hover:bg-surface-2 disabled:opacity-40"
            >
              Cancel
            </button>
            <button
              type="button"
              disabled={encoding || loadingMeta || !!error || durationSec <= 0}
              onClick={() => void handleUseClip()}
              className="h-11 flex-[1.4] rounded-xl bg-brand text-sm font-semibold text-black hover:bg-brand-dark disabled:opacity-40"
            >
              {encoding ? "Creating clip…" : "Use clip"}
            </button>
          </div>
        </footer>
      </div>
    </div>
  );
}
