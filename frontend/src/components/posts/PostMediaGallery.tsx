"use client";

import { useEffect, useRef, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import PostInlinePlayer, {
  type PostInlinePlayerHandle,
} from "@/components/media/PostInlinePlayer";
import type { Media } from "@/lib/content/types";
import { cn } from "@/lib/utils";

export default function PostMediaGallery({
  media,
  mode,
  className,
  tall,
  fill,
  active = true,
}: {
  media: Media[];
  mode?: string;
  className?: string;
  tall?: boolean;
  /** Stretch to parent height (snap feed slides). */
  fill?: boolean;
  /** When false, pause A/V but keep media loaded for browser cache reuse. */
  active?: boolean;
}) {
  const sorted = [...media].sort((a, b) => a.sort_order - b.sort_order);
  const images = sorted.filter((m) => m.kind === "image");
  const video = sorted.find((m) => m.kind === "video");
  const audio = sorted.find((m) => m.kind === "audio");
  const scrollerRef = useRef<HTMLDivElement>(null);
  const compactVideoRef = useRef<HTMLVideoElement>(null);
  const audioRef = useRef<PostInlinePlayerHandle>(null);
  const [index, setIndex] = useState(0);

  useEffect(() => {
    const root = scrollerRef.current;
    if (!root || images.length <= 1) return;

    const slides = Array.from(root.querySelectorAll<HTMLElement>("[data-slide]"));
    if (slides.length === 0) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries
          .filter((e) => e.isIntersecting)
          .sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
        if (!visible) return;
        const i = slides.indexOf(visible.target as HTMLElement);
        if (i >= 0) setIndex(i);
      },
      { root, threshold: [0.6] }
    );

    slides.forEach((el) => observer.observe(el));
    return () => observer.disconnect();
  }, [images.length]);

  useEffect(() => {
    const el = compactVideoRef.current;
    if (!el) return;
    if (!active) el.pause();
    else void el.play().catch(() => undefined);
  }, [active]);

  function goTo(i: number) {
    const root = scrollerRef.current;
    if (!root) return;
    const slide = root.querySelectorAll<HTMLElement>("[data-slide]")[i];
    slide?.scrollIntoView({ behavior: "smooth", inline: "center", block: "nearest" });
    setIndex(i);
  }

  if (video) {
    if (tall || fill) {
      return (
        <PostInlinePlayer
          src={video.url}
          kind="video"
          className={cn(fill && "h-full w-full", className)}
          fill={fill}
          active={active}
        />
      );
    }
    return (
      <div className={cn("relative overflow-hidden bg-black", className)}>
        <video
          ref={compactVideoRef}
          src={video.url}
          muted
          loop
          playsInline
          autoPlay={active}
          preload={active ? "auto" : "metadata"}
          controls={false}
          controlsList="nodownload"
          disablePictureInPicture
          onContextMenu={(e) => e.preventDefault()}
          className="aspect-[4/5] max-h-[28rem] w-full object-cover"
        />
      </div>
    );
  }

  if (images.length === 0) {
    return (
      <div
        className={cn(
          "flex items-center justify-center bg-surface-2 text-sm text-fg-faint",
          fill ? "h-full w-full" : "aspect-[4/5]",
          className
        )}
      >
        No media
      </div>
    );
  }

  const hasAudio = Boolean(audio);
  const interactiveAudio = hasAudio && (tall || fill);

  const frame = fill
    ? "flex min-h-0 flex-1 flex-col"
    : tall
      ? "aspect-[4/5] w-full max-h-[min(78vh,52rem)] sm:aspect-square"
      : "aspect-[4/5] w-full";

  return (
    <div className={cn("relative w-full overflow-hidden bg-black", frame, className)}>
      <div
        ref={scrollerRef}
        className={cn(
          fill
            ? "flex min-h-0 flex-1 snap-x snap-mandatory overflow-x-auto overflow-y-hidden overscroll-x-contain"
            : "absolute inset-0 flex snap-x snap-mandatory overflow-x-auto overflow-y-hidden overscroll-x-contain",
          "scroll-smooth [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden"
        )}
      >
        {images.map((img, i) => (
          <div
            key={img.id || `${img.url}-${i}`}
            data-slide
            className={cn(
              "relative min-w-full shrink-0 snap-center snap-always",
              fill ? "flex min-h-0 flex-1" : "h-full w-full"
            )}
            onClick={(e) => {
              if (!interactiveAudio) return;
              if ((e.target as HTMLElement).closest("[data-player-chrome]")) return;
              if ((e.target as HTMLElement).closest("button")) return;
              audioRef.current?.toggle();
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={img.url}
              alt=""
              draggable={false}
              decoding="async"
              loading="eager"
              className={cn(
                "w-full object-cover select-none",
                fill ? "min-h-0 flex-1" : "h-full",
                interactiveAudio ? "cursor-pointer" : null
              )}
            />
          </div>
        ))}
      </div>

      {hasAudio && (interactiveAudio || mode === "images_audio") ? (
        <div className="absolute inset-x-3 bottom-10 z-10 sm:bottom-12">
          {interactiveAudio ? (
            <PostInlinePlayer
              ref={audioRef}
              src={audio!.url}
              kind="audio"
              overlay
              active={active}
            />
          ) : (
            <div className="h-1 overflow-hidden rounded-full bg-white/30" />
          )}
        </div>
      ) : null}

      {images.length > 1 ? (
        <>
          <div className="pointer-events-none absolute inset-x-0 top-3 z-10 flex justify-center">
            <span className="rounded-full bg-black/45 px-2.5 py-0.5 text-[11px] font-medium text-white/90 backdrop-blur-sm">
              {index + 1}/{images.length}
            </span>
          </div>

          <div className="pointer-events-none absolute inset-x-0 bottom-3 z-10 flex justify-center gap-1.5">
            {images.map((_, i) => (
              <button
                key={i}
                type="button"
                aria-label={`Photo ${i + 1}`}
                onClick={() => goTo(i)}
                className={cn(
                  "pointer-events-auto h-1.5 rounded-full transition-all",
                  i === index ? "w-4 bg-white" : "w-1.5 bg-white/45 hover:bg-white/75"
                )}
              />
            ))}
          </div>

          {index > 0 ? (
            <button
              type="button"
              aria-label="Previous photo"
              onClick={() => goTo(index - 1)}
              className="absolute top-1/2 left-2 z-10 hidden -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white backdrop-blur-sm transition-colors hover:bg-black/60 sm:inline-flex"
            >
              <ChevronLeft className="h-5 w-5" />
            </button>
          ) : null}
          {index < images.length - 1 ? (
            <button
              type="button"
              aria-label="Next photo"
              onClick={() => goTo(index + 1)}
              className="absolute top-1/2 right-2 z-10 hidden -translate-y-1/2 rounded-full bg-black/40 p-1.5 text-white backdrop-blur-sm transition-colors hover:bg-black/60 sm:inline-flex"
            >
              <ChevronRight className="h-5 w-5" />
            </button>
          ) : null}
        </>
      ) : null}
    </div>
  );
}
