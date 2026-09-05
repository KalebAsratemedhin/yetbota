"use client";

import { useMemo, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { useRouter } from "next/navigation";
import { AudioLines, Film, ImagePlus, Music2, Scissors, X } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import LocationPicker from "@/components/compose/LocationPicker";
import { Navbar } from "@/components/landing";
import { useCreatePostMutation, useSignMediaMutation } from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import { parseHashtags, uploadToCloudinary } from "@/lib/content/helpers";
import type { CreateMedia, MediaKind, MediaMode } from "@/lib/content/types";
import {
  DEFAULT_MAX_CLIP_SEC,
  type ClipKind,
  type ClipResult,
} from "@/lib/media";
import { cn } from "@/lib/utils";

const MediaClipStudio = dynamic(() => import("@/components/media/MediaClipStudio"), {
  ssr: false,
});

const MAX_IMAGES = 10;

const MODES: {
  id: MediaMode;
  title: string;
  blurb: string;
  icon: React.ComponentType<{ className?: string }>;
}[] = [
  { id: "images", title: "Photos", blurb: "Up to 10", icon: ImagePlus },
  { id: "images_audio", title: "Photos + audio", blurb: "Audio ≤10s", icon: AudioLines },
  { id: "video", title: "Video", blurb: "Clip ≤10s", icon: Film },
];

type Preview = {
  file: File;
  kind: MediaKind;
  url: string;
  /** Original pick — re-open clip studio without losing source. */
  sourceFile?: File;
};

type ClipSession = {
  file: File;
  kind: ClipKind;
};

function ComposeContent() {
  const router = useRouter();
  const fileRef = useRef<HTMLInputElement>(null);
  const [mode, setMode] = useState<MediaMode>("images");
  const [previews, setPreviews] = useState<Preview[]>([]);
  const [caption, setCaption] = useState("");
  const [title, setTitle] = useState("");
  const [tagsRaw, setTagsRaw] = useState("");
  const [lat, setLat] = useState<number | null>(null);
  const [lng, setLng] = useState<number | null>(null);
  const [placeLabel, setPlaceLabel] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [clipSession, setClipSession] = useState<ClipSession | null>(null);
  const [signMedia] = useSignMediaMutation();
  const [createPost] = useCreatePostMutation();

  const accept = useMemo(() => {
    if (mode === "video") return "video/*";
    if (mode === "images_audio") return "image/*,audio/*";
    return "image/*";
  }, [mode]);

  function resetMedia() {
    setPreviews((prev) => {
      prev.forEach((p) => URL.revokeObjectURL(p.url));
      return [];
    });
  }

  function switchMode(next: MediaMode) {
    setMode(next);
    resetMedia();
    setClipSession(null);
  }

  function openClipStudio(file: File, kind: ClipKind) {
    setClipSession({ file, kind });
  }

  function onClipCancel() {
    setClipSession(null);
  }

  function onClipConfirm(result: ClipResult) {
    const source = clipSession?.file;
    setClipSession(null);
    const url = URL.createObjectURL(result.file);

    if (result.kind === "video") {
      setPreviews((prev) => {
        prev.forEach((p) => URL.revokeObjectURL(p.url));
        return [
          {
            file: result.file,
            kind: "video",
            url,
            sourceFile: source,
          },
        ];
      });
      return;
    }

    setPreviews((prev) => {
      const imgs = prev.filter((p) => p.kind === "image");
      const old = prev.find((p) => p.kind === "audio");
      if (old) URL.revokeObjectURL(old.url);
      return [
        ...imgs,
        {
          file: result.file,
          kind: "audio",
          url,
          sourceFile: source,
        },
      ];
    });
  }

  function onPickFiles(files: FileList | null) {
    if (!files?.length) return;
    const incoming = Array.from(files);

    try {
      if (mode === "video") {
        const file = incoming.find((f) => f.type.startsWith("video/"));
        if (!file) {
          toast.error("Pick a video file");
          return;
        }
        openClipStudio(file, "video");
        return;
      }

      const images = incoming.filter((f) => f.type.startsWith("image/"));
      const audios = incoming.filter((f) => f.type.startsWith("audio/"));

      if (images.length > 0) {
        setPreviews((prev) => {
          const keptImages = prev.filter((p) => p.kind === "image");
          const keptAudio = prev.find((p) => p.kind === "audio");
          const nextImages = [...keptImages];
          for (const file of images) {
            if (nextImages.length >= MAX_IMAGES) break;
            nextImages.push({
              file,
              kind: "image",
              url: URL.createObjectURL(file),
            });
          }
          return keptAudio ? [...nextImages, keptAudio] : nextImages;
        });
      }

      if (mode === "images_audio" && audios[0]) {
        openClipStudio(audios[0], "audio");
      }
    } catch (err) {
      toast.error("Could not add media", {
        description: err instanceof Error ? err.message : undefined,
      });
    } finally {
      if (fileRef.current) fileRef.current.value = "";
    }
  }

  function removePreview(url: string) {
    setPreviews((prev) => {
      const hit = prev.find((p) => p.url === url);
      if (hit) URL.revokeObjectURL(hit.url);
      return prev.filter((p) => p.url !== url);
    });
  }

  function setLocation(nextLat: number, nextLng: number, label?: string) {
    if (!Number.isFinite(nextLat) || !Number.isFinite(nextLng)) return;
    setLat(nextLat);
    setLng(nextLng);
    if (label) {
      setPlaceLabel(label);
      setTitle((prev) => (prev.trim() ? prev : label.slice(0, 200)));
    }
  }

  async function onPublish(e: React.FormEvent) {
    e.preventDefault();
    if (lat == null || lng == null || !Number.isFinite(lat) || !Number.isFinite(lng)) {
      toast.error("Pick a location on the map");
      return;
    }
    const placeName = title.trim();
    if (!placeName) {
      toast.error("Add a place title");
      return;
    }

    const images = previews.filter((p) => p.kind === "image");
    const video = previews.find((p) => p.kind === "video");
    const audio = previews.find((p) => p.kind === "audio");

    if (mode === "images" && images.length < 1) {
      toast.error("Add at least one photo");
      return;
    }
    if (mode === "images_audio" && (images.length < 1 || !audio)) {
      toast.error("Add photos and one audio clip");
      return;
    }
    if (mode === "video" && !video) {
      toast.error("Add a video");
      return;
    }

    setBusy(true);
    try {
      const media: CreateMedia[] = [];
      let order = 0;
      const queue: Preview[] =
        mode === "video"
          ? [video!]
          : mode === "images_audio"
            ? [...images, audio!]
            : images;

      for (const item of queue) {
        const sign = await signMedia({ kind: item.kind }).unwrap();
        const uploaded = await uploadToCloudinary(item.file, sign, item.kind);
        media.push({
          kind: item.kind,
          cloudinary_public_id: uploaded.public_id || sign.public_id,
          url: uploaded.secure_url,
          sort_order: order++,
          duration_ms: uploaded.duration
            ? Math.round(uploaded.duration * 1000)
            : undefined,
          width: uploaded.width,
          height: uploaded.height,
          mime: item.file.type,
        });
      }

      const post = await createPost({
        title: placeName,
        caption: caption.trim(),
        lat,
        lng,
        media_mode: mode,
        media,
        hashtags: parseHashtags(tagsRaw),
      }).unwrap();

      toast.success("Published");
      router.push(`/posts/${post.id}`);
    } catch (err) {
      toast.error("Could not publish", {
        description: getApiErrorMessage(err, "Check content-service and Cloudinary config."),
      });
    } finally {
      setBusy(false);
    }
  }

  const imageCount = previews.filter((p) => p.kind === "image").length;
  const hasAudio = previews.some((p) => p.kind === "audio");
  const canPublish =
    title.trim().length > 0 &&
    lat != null &&
    lng != null &&
    Number.isFinite(lat) &&
    Number.isFinite(lng) &&
    (mode === "video"
      ? Boolean(previews.find((p) => p.kind === "video"))
      : mode === "images_audio"
        ? imageCount > 0 && hasAudio
        : imageCount > 0);

  return (
    <>
      <form
        onSubmit={onPublish}
        className="mx-auto grid min-h-[calc(100vh-5rem)] w-full max-w-[1400px] lg:grid-cols-[minmax(20rem,24rem)_1fr] xl:grid-cols-[minmax(22rem,28rem)_1fr]"
      >
        <section className="relative z-0 isolate h-[42vh] min-h-[18rem] overflow-hidden border-b border-border-subtle lg:order-2 lg:sticky lg:top-20 lg:h-[calc(100vh-5rem)] lg:min-h-0 lg:border-b-0 lg:border-l">
          <LocationPicker lat={lat} lng={lng} onChange={setLocation} />
        </section>

        <aside className="relative z-10 flex flex-col bg-bg lg:order-1">
          <div className="border-b border-border-subtle px-5 py-5 sm:px-6">
            <h1 className="text-xl font-bold tracking-tight text-fg">New post</h1>
            <p className="mt-1 text-sm text-fg-muted">
              Name the place, add media, write a caption, and pin it on the map.
            </p>
          </div>

          <div className="yb-scrollbar flex-1 space-y-6 overflow-y-auto px-5 py-5 sm:px-6">
            <section>
              <div className="mb-2 flex items-center justify-between">
                <label className="text-sm font-medium text-fg">Media</label>
                <span className="text-xs text-fg-faint">
                  {mode === "video"
                    ? previews.some((p) => p.kind === "video")
                      ? "1 video"
                      : "None"
                    : mode === "images_audio"
                      ? `${imageCount}/${MAX_IMAGES} photos · ${hasAudio ? "audio clipped" : "no audio"}`
                      : `${imageCount}/${MAX_IMAGES}`}
                </span>
              </div>

              <div className="grid grid-cols-3 gap-1.5">
                {MODES.map((m) => {
                  const Icon = m.icon;
                  const active = mode === m.id;
                  return (
                    <button
                      key={m.id}
                      type="button"
                      onClick={() => switchMode(m.id)}
                      className={cn(
                        "rounded-xl border px-2.5 py-2.5 text-left transition-colors",
                        active
                          ? "border-brand/50 bg-brand/10"
                          : "border-border-subtle bg-surface hover:bg-surface-2"
                      )}
                    >
                      <Icon
                        className={cn("mb-1.5 h-4 w-4", active ? "text-brand" : "text-fg-faint")}
                      />
                      <p className="text-[11px] font-semibold text-fg">{m.title}</p>
                      <p className="text-[10px] text-fg-faint">{m.blurb}</p>
                    </button>
                  );
                })}
              </div>

              <input
                ref={fileRef}
                type="file"
                accept={accept}
                multiple={mode !== "video"}
                className="sr-only"
                onChange={(e) => onPickFiles(e.target.files)}
              />

              <button
                type="button"
                onClick={() => fileRef.current?.click()}
                className="mt-3 flex w-full items-center justify-center gap-2 rounded-xl border border-dashed border-border-subtle bg-surface px-4 py-6 text-sm text-fg-muted transition-colors hover:border-brand/40 hover:bg-surface-2 hover:text-fg"
              >
                <ImagePlus className="h-4 w-4 shrink-0" />
                {mode === "video"
                  ? "Choose video"
                  : mode === "images_audio"
                    ? "Choose photos or audio"
                    : "Choose photos"}
              </button>

              {previews.length > 0 ? (
                <ul className="mt-3 grid grid-cols-3 gap-2">
                  {previews.map((p) => (
                    <li
                      key={p.url}
                      className="relative aspect-square overflow-hidden rounded-lg border border-border-subtle bg-surface-2"
                    >
                      {p.kind === "image" ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={p.url} alt="" className="h-full w-full object-cover" />
                      ) : p.kind === "video" ? (
                        <video src={p.url} className="h-full w-full object-cover" muted />
                      ) : (
                        <div className="flex h-full flex-col items-center justify-center gap-1 bg-surface-2 text-fg-muted">
                          <Music2 className="h-5 w-5" />
                          <span className="text-[10px]">Audio</span>
                        </div>
                      )}
                      {p.kind === "video" || p.kind === "audio" ? (
                        <button
                          type="button"
                          onClick={() =>
                            openClipStudio(p.sourceFile ?? p.file, p.kind as ClipKind)
                          }
                          className="absolute bottom-1 left-1 inline-flex items-center gap-1 rounded-full bg-black/70 px-2 py-1 text-[10px] font-semibold text-white"
                          aria-label="Edit clip"
                        >
                          <Scissors className="h-3 w-3" />
                          Edit
                        </button>
                      ) : null}
                      <button
                        type="button"
                        onClick={() => removePreview(p.url)}
                        className="absolute top-1 right-1 rounded-full bg-black/70 p-1 text-white"
                        aria-label="Remove"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </li>
                  ))}
                </ul>
              ) : null}
            </section>

            <section className="space-y-3">
              <div>
                <label htmlFor="compose-title" className="mb-1.5 block text-sm font-medium text-fg">
                  Place title
                </label>
                <input
                  id="compose-title"
                  value={title}
                  onChange={(e) => setTitle(e.target.value.slice(0, 200))}
                  placeholder="e.g. Tomoca Coffee – Piazza"
                  maxLength={200}
                  className="w-full rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
                />
              </div>
              <div>
                <label htmlFor="compose-caption" className="mb-1.5 block text-sm font-medium text-fg">
                  Caption
                </label>
                <textarea
                  id="compose-caption"
                  rows={5}
                  value={caption}
                  onChange={(e) => setCaption(e.target.value)}
                  placeholder="What’s at this place?"
                  className="yb-scrollbar max-h-40 w-full resize-none overflow-y-auto rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
                />
              </div>
              <div>
                <label htmlFor="compose-tags" className="mb-1.5 block text-sm font-medium text-fg">
                  Hashtags
                </label>
                <input
                  id="compose-tags"
                  value={tagsRaw}
                  onChange={(e) => setTagsRaw(e.target.value)}
                  placeholder="#mercato #coffee"
                  className="w-full rounded-xl border border-border-subtle bg-surface px-3 py-2.5 text-sm text-fg outline-none placeholder:text-fg-faint focus:border-brand/50"
                />
              </div>
              {placeLabel ? (
                <p className="text-xs text-fg-muted">
                  <span className="text-fg-faint">Pinned: </span>
                  {placeLabel}
                </p>
              ) : (
                <p className="text-xs text-fg-faint">Pin a place on the map to continue.</p>
              )}
            </section>
          </div>

          <div className="sticky bottom-0 border-t border-border-subtle bg-bg px-5 py-4 sm:px-6">
            <button
              type="submit"
              disabled={busy || !canPublish}
              className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black transition-colors hover:bg-brand-dark disabled:opacity-40"
            >
              {busy ? "Publishing…" : "Publish"}
            </button>
          </div>
        </aside>
      </form>

      <MediaClipStudio
        open={Boolean(clipSession)}
        file={clipSession?.file ?? null}
        kind={clipSession?.kind ?? "video"}
        maxDurationSec={DEFAULT_MAX_CLIP_SEC}
        onCancel={onClipCancel}
        onConfirm={onClipConfirm}
      />
    </>
  );
}

export default function ComposePage() {
  return (
    <RequireAuth>
      <main className="min-h-screen overflow-x-clip bg-bg text-fg">
        <Navbar />
        <div className="pt-20">
          <ComposeContent />
        </div>
      </main>
    </RequireAuth>
  );
}
