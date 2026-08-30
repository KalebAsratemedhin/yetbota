"use client";

import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import { Volume2, VolumeX } from "lucide-react";
import { formatClipTime } from "@/lib/media";
import { cn } from "@/lib/utils";

export type PostInlinePlayerHandle = {
  toggle: () => void;
  pause: () => void;
  play: () => void;
};

/**
 * Minimal looping player: seek + mute on hover, center tap pauses/resumes.
 * When `active` is false, media is unloaded to free memory.
 */
const PostInlinePlayer = forwardRef<
  PostInlinePlayerHandle,
  {
    src: string;
    kind: "audio" | "video";
    className?: string;
    overlay?: boolean;
    /** When false, pause and drop the media src (feed windowing). */
    active?: boolean;
  }
>(function PostInlinePlayer(
  { src, kind, className, overlay, active = true },
  ref
) {
  const elRef = useRef<HTMLVideoElement | HTMLAudioElement | null>(null);
  const mutedRef = useRef(false);
  const [current, setCurrent] = useState(0);
  const [duration, setDuration] = useState(0);
  const [muted, setMuted] = useState(false);
  const [playing, setPlaying] = useState(false);
  const [ready, setReady] = useState(false);
  const [chromeVisible, setChromeVisible] = useState(false);
  const hideTimer = useRef<number | null>(null);

  useEffect(() => {
    mutedRef.current = muted;
  }, [muted]);

  useEffect(() => {
    const el = elRef.current;
    if (!el) return;

    if (!active) {
      el.pause();
      el.removeAttribute("src");
      el.load();
      queueMicrotask(() => {
        setPlaying(false);
        setReady(false);
        setCurrent(0);
      });
      return;
    }

    el.src = src;
    el.muted = mutedRef.current;
    el.load();

    const tryPlay = () => {
      void el.play().then(
        () => setPlaying(true),
        () => {
          el.muted = true;
          setMuted(true);
          void el.play().then(
            () => setPlaying(true),
            () => setPlaying(false)
          );
        }
      );
    };

    const onMeta = () => {
      if (Number.isFinite(el.duration)) setDuration(el.duration);
      setReady(true);
      tryPlay();
    };
    const onTime = () => setCurrent(el.currentTime);
    const onPlay = () => setPlaying(true);
    const onPause = () => setPlaying(false);
    const onEnded = () => {
      el.currentTime = 0;
      void el.play().catch(() => undefined);
    };

    el.addEventListener("loadedmetadata", onMeta);
    el.addEventListener("timeupdate", onTime);
    el.addEventListener("play", onPlay);
    el.addEventListener("pause", onPause);
    el.addEventListener("ended", onEnded);

    return () => {
      el.removeEventListener("loadedmetadata", onMeta);
      el.removeEventListener("timeupdate", onTime);
      el.removeEventListener("play", onPlay);
      el.removeEventListener("pause", onPause);
      el.removeEventListener("ended", onEnded);
      el.pause();
    };
  }, [src, active]);

  useEffect(() => {
    const el = elRef.current;
    if (!el || !active) return;
    el.muted = muted;
  }, [muted, active]);

  useImperativeHandle(ref, () => ({
    toggle() {
      const el = elRef.current;
      if (!el || !active) return;
      if (el.paused) void el.play().catch(() => undefined);
      else el.pause();
    },
    pause() {
      elRef.current?.pause();
    },
    play() {
      void elRef.current?.play().catch(() => undefined);
    },
  }));

  function showChrome() {
    setChromeVisible(true);
    if (hideTimer.current != null) window.clearTimeout(hideTimer.current);
  }

  function scheduleHideChrome() {
    if (hideTimer.current != null) window.clearTimeout(hideTimer.current);
    hideTimer.current = window.setTimeout(() => setChromeVisible(false), 1200);
  }

  function onSeek(e: React.ChangeEvent<HTMLInputElement>) {
    const el = elRef.current;
    if (!el) return;
    const t = Number(e.target.value);
    if (!Number.isFinite(t)) return;
    el.currentTime = t;
    setCurrent(t);
    void el.play().catch(() => undefined);
  }

  function onSurfaceClick(e: React.MouseEvent) {
    if ((e.target as HTMLElement).closest("[data-player-chrome]")) return;
    const el = elRef.current;
    if (!el || !active) return;
    if (el.paused) void el.play().catch(() => undefined);
    else el.pause();
  }

  const bar = (
    <div
      data-player-chrome
      className={cn(
        "flex items-center gap-2 transition-opacity duration-200",
        chromeVisible ? "opacity-100" : "pointer-events-none opacity-0",
        overlay
          ? "rounded-full bg-black/55 px-2.5 py-1.5 backdrop-blur-md"
          : "rounded-full bg-surface-2 px-2.5 py-1.5"
      )}
    >
      <input
        type="range"
        min={0}
        max={duration > 0 ? duration : 0}
        step={0.05}
        value={Math.min(current, duration || 0)}
        disabled={!ready || duration <= 0}
        onChange={onSeek}
        aria-label="Seek"
        className="yb-seek h-1 min-w-0 flex-1 cursor-pointer appearance-none bg-transparent"
      />
      <span
        className={cn(
          "w-10 shrink-0 text-right text-[10px] tabular-nums",
          overlay ? "text-white/80" : "text-fg-faint"
        )}
      >
        {formatClipTime(current)}
      </span>
      <button
        type="button"
        onClick={(e) => {
          e.stopPropagation();
          setMuted((m) => !m);
        }}
        aria-label={muted ? "Unmute" : "Mute"}
        className={cn(
          "shrink-0 rounded-full p-1",
          overlay ? "text-white hover:bg-white/10" : "text-fg hover:bg-surface"
        )}
      >
        {muted ? <VolumeX className="h-4 w-4" /> : <Volume2 className="h-4 w-4" />}
      </button>
    </div>
  );

  if (kind === "audio") {
    return (
      <div
        className={cn("relative", className)}
        onMouseEnter={showChrome}
        onMouseLeave={scheduleHideChrome}
        onFocusCapture={showChrome}
        onPointerDown={showChrome}
      >
        <audio
          ref={(n) => {
            elRef.current = n;
          }}
          loop
          playsInline
          preload={active ? "auto" : "none"}
          controls={false}
          controlsList="nodownload noplaybackrate"
          onContextMenu={(e) => e.preventDefault()}
          className="hidden"
        />
        {/* Always show a thin hit area; bar fades in on hover */}
        <div className="min-h-8">{bar}</div>
      </div>
    );
  }

  return (
    <div
      className={cn("relative cursor-pointer bg-black", className)}
      onClick={onSurfaceClick}
      onMouseEnter={showChrome}
      onMouseLeave={scheduleHideChrome}
      onFocusCapture={showChrome}
    >
      <video
        ref={(n) => {
          elRef.current = n;
        }}
        loop
        playsInline
        preload={active ? "auto" : "none"}
        controls={false}
        controlsList="nodownload noplaybackrate noremoteplayback"
        disablePictureInPicture
        onContextMenu={(e) => e.preventDefault()}
        className="max-h-[min(78vh,52rem)] w-full object-contain"
      />
      {!playing && active ? (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <span className="rounded-full bg-black/45 px-3 py-1 text-xs font-semibold text-white/90 backdrop-blur-sm">
            Paused
          </span>
        </div>
      ) : null}
      <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/55 to-transparent px-3 pt-10 pb-3">
        <div className="pointer-events-auto">{bar}</div>
      </div>
    </div>
  );
});

export default PostInlinePlayer;
