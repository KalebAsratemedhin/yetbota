"use client";

import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from "react";
import { Pause, Play, Volume2, VolumeX } from "lucide-react";
import { formatClipTime } from "@/lib/media";
import { cn } from "@/lib/utils";

export type PostInlinePlayerHandle = {
  toggle: () => void;
  pause: () => void;
  play: () => void;
};

/** Center play/pause cue — brief flash on toggle, idle play while paused. */
export function MediaTransportCue({
  playing,
  active,
  flash,
  flashKey = 0,
  className,
}: {
  playing: boolean;
  active: boolean;
  flash: "play" | "pause" | null;
  flashKey?: number;
  className?: string;
}) {
  if (!active) return null;

  return (
    <div
      className={cn(
        "pointer-events-none absolute inset-0 z-20 flex items-center justify-center",
        className
      )}
      aria-hidden
    >
      {flash ? (
        <div
          key={flashKey}
          className="media-transport-flash flex h-[4.25rem] w-[4.25rem] items-center justify-center rounded-full bg-black/55 text-white shadow-[0_8px_28px_-12px_rgba(0,0,0,0.65)] backdrop-blur-md"
        >
          {flash === "pause" ? (
            <Pause className="h-8 w-8 fill-current" strokeWidth={1.75} />
          ) : (
            <Play className="ml-0.5 h-8 w-8 fill-current" strokeWidth={1.75} />
          )}
        </div>
      ) : !playing ? (
        <div className="media-transport-idle flex h-14 w-14 items-center justify-center rounded-full bg-black/50 text-white backdrop-blur-sm">
          <Play className="ml-0.5 h-7 w-7 fill-current" strokeWidth={1.75} />
        </div>
      ) : null}
    </div>
  );
}

/**
 * Minimal looping player: seek + mute on hover, center tap pauses/resumes.
 * When `active` is false, playback pauses but `src` stays set so the browser
 * can reuse its HTTP cache on the next view.
 */
const PostInlinePlayer = forwardRef<
  PostInlinePlayerHandle,
  {
    src: string;
    kind: "audio" | "video";
    className?: string;
    overlay?: boolean;
    /** When false, pause playback (keep src for cache). */
    active?: boolean;
    /** Fill parent instead of max-height contain. */
    fill?: boolean;
    onPlayingChange?: (playing: boolean) => void;
    onTransportCue?: (cue: "play" | "pause") => void;
  }
>(function PostInlinePlayer(
  {
    src,
    kind,
    className,
    overlay,
    active = true,
    fill,
    onPlayingChange,
    onTransportCue,
  },
  ref
) {
  const elRef = useRef<HTMLVideoElement | HTMLAudioElement | null>(null);
  const mutedRef = useRef(false);
  const flashTimer = useRef<number | null>(null);
  const [current, setCurrent] = useState(0);
  const [duration, setDuration] = useState(0);
  const [muted, setMuted] = useState(false);
  const [playing, setPlaying] = useState(false);
  const [ready, setReady] = useState(false);
  const [chromeVisible, setChromeVisible] = useState(false);
  const [flash, setFlash] = useState<"play" | "pause" | null>(null);
  const [flashKey, setFlashKey] = useState(0);
  const hideTimer = useRef<number | null>(null);
  const onPlayingChangeRef = useRef(onPlayingChange);
  const onTransportCueRef = useRef(onTransportCue);
  onPlayingChangeRef.current = onPlayingChange;
  onTransportCueRef.current = onTransportCue;

  useEffect(() => {
    mutedRef.current = muted;
  }, [muted]);

  useEffect(() => {
    onPlayingChangeRef.current?.(playing);
  }, [playing]);

  useEffect(() => {
    return () => {
      if (flashTimer.current != null) window.clearTimeout(flashTimer.current);
      if (hideTimer.current != null) window.clearTimeout(hideTimer.current);
    };
  }, []);

  function bumpFlash(cue: "play" | "pause") {
    setFlash(cue);
    setFlashKey((k) => k + 1);
    onTransportCueRef.current?.(cue);
    if (flashTimer.current != null) window.clearTimeout(flashTimer.current);
    flashTimer.current = window.setTimeout(() => setFlash(null), 520);
  }

  function togglePlayback() {
    const el = elRef.current;
    if (!el || !active) return;
    if (el.paused) {
      void el.play().catch(() => undefined);
      bumpFlash("play");
    } else {
      el.pause();
      bumpFlash("pause");
    }
  }

  useEffect(() => {
    const el = elRef.current;
    if (!el) return;

    const tryPlay = () => {
      if (!active) return;
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
      if (active) void el.play().catch(() => undefined);
    };

    el.addEventListener("loadedmetadata", onMeta);
    el.addEventListener("timeupdate", onTime);
    el.addEventListener("play", onPlay);
    el.addEventListener("pause", onPause);
    el.addEventListener("ended", onEnded);

    if (!active) {
      el.pause();
      setFlash(null);
      return () => {
        el.removeEventListener("loadedmetadata", onMeta);
        el.removeEventListener("timeupdate", onTime);
        el.removeEventListener("play", onPlay);
        el.removeEventListener("pause", onPause);
        el.removeEventListener("ended", onEnded);
      };
    }

    if (el.readyState >= 1) {
      setReady(true);
      if (Number.isFinite(el.duration)) setDuration(el.duration);
      tryPlay();
    }

    return () => {
      el.removeEventListener("loadedmetadata", onMeta);
      el.removeEventListener("timeupdate", onTime);
      el.removeEventListener("play", onPlay);
      el.removeEventListener("pause", onPause);
      el.removeEventListener("ended", onEnded);
    };
  }, [src, active]);

  useEffect(() => {
    const el = elRef.current;
    if (!el) return;
    el.muted = muted;
  }, [muted]);

  useImperativeHandle(ref, () => ({
    toggle() {
      togglePlayback();
    },
    pause() {
      elRef.current?.pause();
    },
    play() {
      if (!active) return;
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
    if (active) void el.play().catch(() => undefined);
  }

  function onSurfaceClick(e: React.MouseEvent) {
    if ((e.target as HTMLElement).closest("[data-player-chrome]")) return;
    togglePlayback();
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
          src={src}
          loop
          playsInline
          preload={active ? "auto" : "metadata"}
          controls={false}
          controlsList="nodownload noplaybackrate"
          onContextMenu={(e) => e.preventDefault()}
          className="hidden"
        />
        <div className="min-h-8">{bar}</div>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "relative cursor-pointer bg-black",
        fill && "flex h-full w-full items-center justify-center",
        className
      )}
      onClick={onSurfaceClick}
      onMouseEnter={showChrome}
      onMouseLeave={scheduleHideChrome}
      onFocusCapture={showChrome}
    >
      <video
        ref={(n) => {
          elRef.current = n;
        }}
        src={src}
        loop
        playsInline
        preload={active ? "auto" : "metadata"}
        controls={false}
        controlsList="nodownload noplaybackrate noremoteplayback"
        disablePictureInPicture
        onContextMenu={(e) => e.preventDefault()}
        className={
          fill
            ? "h-full max-h-full w-full object-contain"
            : "max-h-[min(78vh,52rem)] w-full object-contain"
        }
      />
      <MediaTransportCue
        playing={playing}
        active={active}
        flash={flash}
        flashKey={flashKey}
      />
      <div className="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/55 to-transparent px-3 pt-10 pb-3">
        <div className="pointer-events-auto">{bar}</div>
      </div>
    </div>
  );
});

export default PostInlinePlayer;
