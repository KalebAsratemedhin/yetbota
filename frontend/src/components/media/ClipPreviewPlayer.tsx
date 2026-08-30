"use client";

import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
} from "react";
import type { ClipKind, ClipRange } from "@/lib/media";
import { cn } from "@/lib/utils";

export type ClipPreviewHandle = {
  /** Call from a click handler so the browser allows sound. */
  play: () => Promise<void>;
  pause: () => void;
  /** Jump to a time in the source (clip in-point after a trim drag). */
  seek: (timeSec: number) => void;
};

/**
 * Loops playback inside the selected clip range for preview-before-export.
 * Audio uses Web Audio (reliable). Video uses HTMLVideoElement.
 */
const ClipPreviewPlayer = forwardRef<
  ClipPreviewHandle,
  {
    src: string;
    /** Original file — preferred for Web Audio decode (blob URLs can be flaky). */
    file?: File | null;
    kind: ClipKind;
    range: ClipRange;
    playing: boolean;
    onPlayhead?: (sec: number) => void;
    onPlayingChange?: (playing: boolean) => void;
    className?: string;
  }
>(function ClipPreviewPlayer(
  { src, file, kind, range, playing, onPlayhead, onPlayingChange, className },
  ref
) {
  const videoRef = useRef<HTMLVideoElement | null>(null);
  const rangeRef = useRef(range);
  rangeRef.current = range;
  const playingRef = useRef(playing);
  playingRef.current = playing;

  const audioCtxRef = useRef<AudioContext | null>(null);
  const audioBufRef = useRef<AudioBuffer | null>(null);
  const audioSrcRef = useRef<AudioBufferSourceNode | null>(null);
  const audioStartedAtRef = useRef(0);
  const audioOffsetRef = useRef(0);
  const rafRef = useRef<number | null>(null);

  function stopAudioGraph() {
    if (rafRef.current != null) {
      cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    }
    try {
      audioSrcRef.current?.stop();
    } catch {
      /* already stopped */
    }
    audioSrcRef.current?.disconnect();
    audioSrcRef.current = null;
  }

  function tickAudioPlayhead() {
    const ctx = audioCtxRef.current;
    if (!ctx || !playingRef.current) return;
    const r = rangeRef.current;
    const elapsed = ctx.currentTime - audioStartedAtRef.current;
    const t = audioOffsetRef.current + elapsed;
    if (t >= r.end - 0.02) {
      // Loop selection
      void startAudioFrom(r.start);
      return;
    }
    onPlayhead?.(t);
    rafRef.current = requestAnimationFrame(tickAudioPlayhead);
  }

  async function ensureAudioBuffer(): Promise<AudioBuffer> {
    if (audioBufRef.current) return audioBufRef.current;
    const ctx = audioCtxRef.current ?? new AudioContext();
    audioCtxRef.current = ctx;
    if (ctx.state === "suspended") {
      await ctx.resume();
    }
    const bytes = file
      ? await file.arrayBuffer()
      : await (await fetch(src)).arrayBuffer();
    const buf = await ctx.decodeAudioData(bytes.slice(0));
    audioBufRef.current = buf;
    return buf;
  }

  async function startAudioFrom(at: number) {
    const ctx = audioCtxRef.current ?? new AudioContext();
    audioCtxRef.current = ctx;
    if (ctx.state === "suspended") {
      await ctx.resume();
    }
    const buffer = await ensureAudioBuffer();
    stopAudioGraph();

    const offset = Math.min(Math.max(0, at), Math.max(0, buffer.duration - 0.01));
    const source = ctx.createBufferSource();
    source.buffer = buffer;
    source.connect(ctx.destination);
    // Play through end of file; RAF loops back at range.end
    source.start(0, offset);
    audioSrcRef.current = source;
    audioOffsetRef.current = offset;
    audioStartedAtRef.current = ctx.currentTime;
    onPlayhead?.(offset);
    rafRef.current = requestAnimationFrame(tickAudioPlayhead);
  }

  useImperativeHandle(ref, () => ({
    async play() {
      if (kind === "audio") {
        // Create + resume in the gesture turn before any heavy await.
        const ctx = audioCtxRef.current ?? new AudioContext();
        audioCtxRef.current = ctx;
        if (ctx.state === "suspended") {
          await ctx.resume();
        }
        playingRef.current = true;
        onPlayingChange?.(true);
        await startAudioFrom(rangeRef.current.start);
        return;
      }

      const el = videoRef.current;
      if (!el) throw new Error("Video not ready");
      el.muted = false;
      el.volume = 1;
      // Do not await seek — that drops the user gesture and browsers block play().
      try {
        el.currentTime = rangeRef.current.start;
      } catch {
        /* ignore */
      }
      await el.play();
      onPlayingChange?.(true);
    },
    pause() {
      if (kind === "audio") {
        playingRef.current = false;
        stopAudioGraph();
        onPlayingChange?.(false);
        return;
      }
      videoRef.current?.pause();
      onPlayingChange?.(false);
    },
    seek(timeSec: number) {
      const t = Math.max(0, timeSec);
      if (kind === "audio") {
        if (playingRef.current) {
          void startAudioFrom(t);
        } else {
          onPlayhead?.(t);
        }
        return;
      }
      const el = videoRef.current;
      if (!el) return;
      try {
        el.currentTime = t;
      } catch {
        /* ignore */
      }
      onPlayhead?.(t);
    },
  }));

  // Video: loop within selection
  useEffect(() => {
    if (kind !== "video") return;
    const el = videoRef.current;
    if (!el) return;

    const onTime = () => {
      const r = rangeRef.current;
      onPlayhead?.(el.currentTime);
      if (el.currentTime >= r.end - 0.05) {
        el.currentTime = r.start;
        if (!playingRef.current) el.pause();
      }
    };
    el.addEventListener("timeupdate", onTime);
    return () => el.removeEventListener("timeupdate", onTime);
  }, [kind, onPlayhead]);

  useEffect(() => {
    if (kind !== "video") return;
    const el = videoRef.current;
    if (!el) return;
    if (!playing) el.pause();
  }, [kind, playing]);

  // Invalidate decoded audio when source changes
  useEffect(() => {
    audioBufRef.current = null;
    stopAudioGraph();
    return () => {
      stopAudioGraph();
    };
  }, [src, file]);

  // Stop audio graph on unmount
  useEffect(() => {
    return () => {
      stopAudioGraph();
      void audioCtxRef.current?.close().catch(() => undefined);
      audioCtxRef.current = null;
    };
  }, []);

  // If parent clears playing (encode start), stop audio
  useEffect(() => {
    if (kind === "audio" && !playing) stopAudioGraph();
  }, [kind, playing]);

  if (kind === "audio") {
    return (
      <div
        className={cn(
          "flex flex-col items-center justify-center gap-4 rounded-2xl bg-surface-2 px-6 py-10",
          className
        )}
      >
        <div
          className={cn(
            "flex h-16 w-16 items-center justify-center rounded-full bg-brand/15 text-brand transition-transform",
            playing && "animate-pulse scale-105"
          )}
        >
          <span className="text-2xl font-bold">♪</span>
        </div>
        <p className="text-xs text-fg-faint">
          {playing ? "Playing selection…" : "Preview loops your selection"}
        </p>
      </div>
    );
  }

  return (
    <div className={cn("overflow-hidden rounded-2xl bg-black", className)}>
      <video
        ref={videoRef}
        src={src}
        playsInline
        preload="auto"
        className="aspect-[9/16] max-h-[50vh] w-full object-contain sm:aspect-video sm:max-h-[40vh]"
      />
    </div>
  );
});

export default ClipPreviewPlayer;
