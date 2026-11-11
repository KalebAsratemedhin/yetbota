"use client";

import { Suspense, useCallback, useEffect, useMemo, useRef, useState } from "react";
import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import { AudioLines, Film, ImagePlus, Music2, Scissors, X } from "lucide-react";
import { toast } from "sonner";
import RequireAuth from "@/components/auth/RequireAuth";
import LocationPicker from "@/components/compose/LocationPicker";
import { Navbar } from "@/components/landing";
import {
  useCreatePostMutation,
  useGetPostQuery,
  useSignMediaMutation,
  useUpdatePostMutation,
} from "@/lib/api/contentApi";
import { getApiErrorMessage } from "@/lib/auth/errors";
import {
  DEFAULT_MAX_UPLOAD_BYTES,
  formatMaxUploadLabel,
  parseHashtags,
  uploadToCloudinary,
} from "@/lib/content/helpers";
import type { CreateMedia, MediaKind, MediaMode, UpdatePostRequest } from "@/lib/content/types";
import {
  DEFAULT_MAX_CLIP_SEC,
  type ClipKind,
  type ClipResult,
} from "@/lib/media";
import { useAppSelector } from "@/lib/hooks";
import { cn } from "@/lib/utils";

const MediaClipStudio = dynamic(() => import("@/components/media/MediaClipStudio"), {
  ssr: false,
});

const MAX_IMAGES = 10;
const MEDIA_MODES: MediaMode[] = ["images", "images_audio", "video"];

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

type UploadStatus = "pending" | "uploading" | "ready" | "error";

type Preview = {
  id: string;
  /** Present for newly picked / clipped files. */
  file?: File;
  kind: MediaKind;
  url: string;
  /** Original pick — re-open clip studio without losing source. */
  sourceFile?: File;
  /** Existing Cloudinary asset kept without re-upload. */
  existing?: CreateMedia;
  /** Uploaded asset for new files. */
  uploaded?: CreateMedia;
  uploadStatus: UploadStatus;
  uploadProgress: number;
  uploadError?: string;
};

type ClipSession = {
  file: File;
  kind: ClipKind;
};

function isMediaMode(v: string): v is MediaMode {
  return (MEDIA_MODES as string[]).includes(v);
}

function isMediaKind(v: string): v is MediaKind {
  return v === "image" || v === "video" || v === "audio";
}

function newPreviewId() {
  return crypto.randomUUID();
}

function ComposeContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const editId = searchParams.get("edit")?.trim() || "";
  const isEdit = Boolean(editId);
  const viewerId = useAppSelector((s) => s.auth.userId);

  const postQ = useGetPostQuery(editId, { skip: !editId });
  const [hydratedEditId, setHydratedEditId] = useState<string | null>(null);

  const fileRef = useRef<HTMLInputElement>(null);
  const abortById = useRef(new Map<string, AbortController>());
  const maxUploadBytesRef = useRef(DEFAULT_MAX_UPLOAD_BYTES);

  const [mode, setMode] = useState<MediaMode>("images");
  const [previews, setPreviews] = useState<Preview[]>([]);
  const [mediaDirty, setMediaDirty] = useState(false);
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
  const [updatePost] = useUpdatePostMutation();

  const patchPreview = useCallback((id: string, patch: Partial<Preview>) => {
    setPreviews((prev) => prev.map((p) => (p.id === id ? { ...p, ...patch } : p)));
  }, []);

  const abortUpload = useCallback((id: string) => {
    const ctrl = abortById.current.get(id);
    if (ctrl) {
      ctrl.abort();
      abortById.current.delete(id);
    }
  }, []);

  const abortAllUploads = useCallback(() => {
    abortById.current.forEach((ctrl) => ctrl.abort());
    abortById.current.clear();
  }, []);

  const startUpload = useCallback(
    async (id: string, file: File, kind: MediaKind) => {
      abortUpload(id);
      const ctrl = new AbortController();
      abortById.current.set(id, ctrl);

      const knownMax = maxUploadBytesRef.current || DEFAULT_MAX_UPLOAD_BYTES;
      if (file.size > knownMax) {
        const msg = `Max file size is ${formatMaxUploadLabel(knownMax)}.`;
        patchPreview(id, {
          uploadStatus: "error",
          uploadProgress: 0,
          uploadError: msg,
        });
        toast.error(msg);
        abortById.current.delete(id);
        return;
      }

      patchPreview(id, {
        uploadStatus: "uploading",
        uploadProgress: 0,
        uploadError: undefined,
        uploaded: undefined,
      });

      try {
        const sign = await signMedia({ kind }).unwrap();
        if (ctrl.signal.aborted) return;

        const maxBytes =
          sign.max_upload_bytes || sign.max_file_size || DEFAULT_MAX_UPLOAD_BYTES;
        maxUploadBytesRef.current = maxBytes;

        if (file.size > maxBytes) {
          const msg = `Max file size is ${formatMaxUploadLabel(maxBytes)}.`;
          patchPreview(id, {
            uploadStatus: "error",
            uploadProgress: 0,
            uploadError: msg,
          });
          toast.error(msg);
          return;
        }

        const uploaded = await uploadToCloudinary(file, sign, kind, {
          signal: ctrl.signal,
          onProgress: (pct) => patchPreview(id, { uploadProgress: pct }),
        });

        if (ctrl.signal.aborted) return;

        const asset: CreateMedia = {
          kind,
          cloudinary_public_id: uploaded.public_id || sign.public_id,
          url: uploaded.secure_url,
          duration_ms: uploaded.duration
            ? Math.round(uploaded.duration * 1000)
            : undefined,
          width: uploaded.width,
          height: uploaded.height,
          mime: file.type,
          bytes: file.size,
        };

        patchPreview(id, {
          uploadStatus: "ready",
          uploadProgress: 100,
          uploaded: asset,
          uploadError: undefined,
        });
      } catch (err) {
        if (ctrl.signal.aborted || (err instanceof DOMException && err.name === "AbortError")) {
          return;
        }
        const msg =
          err instanceof Error ? err.message : "Upload failed. Try again.";
        patchPreview(id, {
          uploadStatus: "error",
          uploadProgress: 0,
          uploadError: msg,
        });
        toast.error("Upload failed", {
          description: getApiErrorMessage(err, msg),
        });
      } finally {
        abortById.current.delete(id);
      }
    },
    [abortUpload, patchPreview, signMedia]
  );

  useEffect(() => {
    return () => abortAllUploads();
  }, [abortAllUploads]);

  useEffect(() => {
    if (!isEdit || !postQ.data || hydratedEditId === postQ.data.id) return;
    const post = postQ.data;
    if (viewerId && post.author_id !== viewerId) {
      toast.error("You can only edit your own posts");
      router.replace(`/posts/${post.id}`);
      return;
    }
    const nextMode = isMediaMode(post.media_mode) ? post.media_mode : "images";
    setMode(nextMode);
    setTitle(post.title);
    setCaption(post.caption);
    setTagsRaw(post.hashtags.map((t) => `#${t}`).join(" "));
    setLat(post.lat);
    setLng(post.lng);
    setPlaceLabel(post.title.trim() || null);
    setMediaDirty(false);
    abortAllUploads();
    setPreviews((prev) => {
      prev.forEach((p) => {
        if (p.file || p.sourceFile) URL.revokeObjectURL(p.url);
      });
      return [...post.media]
        .sort((a, b) => a.sort_order - b.sort_order)
        .filter((m) => isMediaKind(m.kind))
        .map((m) => ({
          id: newPreviewId(),
          kind: m.kind as MediaKind,
          url: m.url,
          existing: {
            kind: m.kind as MediaKind,
            cloudinary_public_id: m.cloudinary_public_id,
            url: m.url,
            sort_order: m.sort_order,
            duration_ms: m.duration_ms || undefined,
            width: m.width || undefined,
            height: m.height || undefined,
            mime: m.mime || undefined,
          },
          uploadStatus: "ready" as const,
          uploadProgress: 100,
        }));
    });
    setHydratedEditId(post.id);
  }, [isEdit, postQ.data, hydratedEditId, viewerId, router, abortAllUploads]);

  const accept = useMemo(() => {
    if (mode === "video") return "video/*";
    if (mode === "images_audio") return "image/*,audio/*";
    return "image/*";
  }, [mode]);

  function resetMedia() {
    abortAllUploads();
    setPreviews((prev) => {
      prev.forEach((p) => {
        if (p.file || p.sourceFile) URL.revokeObjectURL(p.url);
      });
      return [];
    });
    setMediaDirty(true);
  }

  function switchMode(next: MediaMode) {
    setMode(next);
    resetMedia();
    setClipSession(null);
  }

  function openClipStudio(file: File, kind: ClipKind) {
    const max = maxUploadBytesRef.current || DEFAULT_MAX_UPLOAD_BYTES;
    if (file.size > max) {
      toast.error(`Max file size is ${formatMaxUploadLabel(max)}.`);
      return;
    }
    setClipSession({ file, kind });
  }

  function onClipCancel() {
    setClipSession(null);
  }

  function onClipConfirm(result: ClipResult) {
    const source = clipSession?.file;
    setClipSession(null);
    const url = URL.createObjectURL(result.file);
    setMediaDirty(true);
    const id = newPreviewId();

    if (result.kind === "video") {
      setPreviews((prev) => {
        prev.forEach((p) => {
          abortUpload(p.id);
          if (p.file || p.sourceFile) URL.revokeObjectURL(p.url);
        });
        return [
          {
            id,
            file: result.file,
            kind: "video",
            url,
            sourceFile: source,
            uploadStatus: "pending",
            uploadProgress: 0,
          },
        ];
      });
      void startUpload(id, result.file, "video");
      return;
    }

    setPreviews((prev) => {
      const imgs = prev.filter((p) => p.kind === "image");
      const old = prev.find((p) => p.kind === "audio");
      if (old) {
        abortUpload(old.id);
        if (old.file || old.sourceFile) URL.revokeObjectURL(old.url);
      }
      return [
        ...imgs,
        {
          id,
          file: result.file,
          kind: "audio",
          url,
          sourceFile: source,
          uploadStatus: "pending",
          uploadProgress: 0,
        },
      ];
    });
    void startUpload(id, result.file, "audio");
  }

  function onPickFiles(files: FileList | null) {
    if (!files?.length) return;
    const incoming = Array.from(files);
    const max = maxUploadBytesRef.current || DEFAULT_MAX_UPLOAD_BYTES;

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
        setMediaDirty(true);
        setPreviews((prev) => {
          const keptImages = prev.filter((p) => p.kind === "image");
          const keptAudio = prev.find((p) => p.kind === "audio");
          let slots = MAX_IMAGES - keptImages.length;
          const nextImages = [...keptImages];
          const started: { id: string; file: File }[] = [];
          for (const file of images) {
            if (slots <= 0) break;
            if (file.size > max) {
              toast.error(`Max file size is ${formatMaxUploadLabel(max)}.`, {
                description: file.name,
              });
              continue;
            }
            const id = newPreviewId();
            nextImages.push({
              id,
              file,
              kind: "image",
              url: URL.createObjectURL(file),
              uploadStatus: "pending",
              uploadProgress: 0,
            });
            started.push({ id, file });
            slots -= 1;
          }
          queueMicrotask(() => {
            for (const s of started) void startUpload(s.id, s.file, "image");
          });
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

  function removePreview(id: string) {
    setMediaDirty(true);
    abortUpload(id);
    setPreviews((prev) => {
      const hit = prev.find((p) => p.id === id);
      if (hit && (hit.file || hit.sourceFile)) URL.revokeObjectURL(hit.url);
      return prev.filter((p) => p.id !== id);
    });
  }

  function retryUpload(p: Preview) {
    if (!p.file) return;
    void startUpload(p.id, p.file, p.kind);
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

  function collectMediaPayload(): CreateMedia[] {
    const images = previews.filter((p) => p.kind === "image");
    const video = previews.find((p) => p.kind === "video");
    const audio = previews.find((p) => p.kind === "audio");
    const queue: Preview[] =
      mode === "video"
        ? video
          ? [video]
          : []
        : mode === "images_audio"
          ? audio
            ? [...images, audio]
            : [...images]
          : images;

    return queue.map((item, order) => {
      const asset = item.uploaded ?? item.existing;
      if (!asset) {
        throw new Error("Media not uploaded yet");
      }
      return { ...asset, sort_order: order };
    });
  }

  function mediaReady(): boolean {
    const images = previews.filter((p) => p.kind === "image");
    const video = previews.find((p) => p.kind === "video");
    const audio = previews.find((p) => p.kind === "audio");
    if (mode === "images") return images.length >= 1;
    if (mode === "images_audio") return images.length >= 1 && Boolean(audio);
    return Boolean(video);
  }

  const uploadsInFlight = previews.some(
    (p) => p.uploadStatus === "pending" || p.uploadStatus === "uploading"
  );
  const uploadsFailed = previews.some((p) => p.uploadStatus === "error");
  const allMediaReady =
    previews.length > 0 &&
    previews.every((p) => p.uploadStatus === "ready" && (p.uploaded || p.existing));

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

    if (!isEdit || mediaDirty) {
      if (mode === "images" && !mediaReady()) {
        toast.error("Add at least one photo");
        return;
      }
      if (mode === "images_audio" && !mediaReady()) {
        toast.error("Add photos and one audio clip");
        return;
      }
      if (mode === "video" && !mediaReady()) {
        toast.error("Add a video");
        return;
      }
      if (uploadsInFlight) {
        toast.error("Wait for uploads to finish");
        return;
      }
      if (uploadsFailed || !allMediaReady) {
        toast.error("Fix failed uploads before publishing");
        return;
      }
    }

    setBusy(true);
    try {
      if (isEdit) {
        const body: UpdatePostRequest = {
          title: placeName,
          caption: caption.trim(),
          lat,
          lng,
          hashtags: parseHashtags(tagsRaw),
        };
        if (mediaDirty) {
          body.media_mode = mode;
          body.media = collectMediaPayload();
        }
        const post = await updatePost({ id: editId, body }).unwrap();
        toast.success("Post updated");
        router.push(`/posts/${post.id}`);
        return;
      }

      const media = collectMediaPayload();
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
      toast.error(isEdit ? "Could not update post" : "Could not publish", {
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
    !uploadsInFlight &&
    !uploadsFailed &&
    (isEdit && !mediaDirty
      ? true
      : mediaReady() && allMediaReady);

  if (isEdit && postQ.isLoading) {
    return (
      <p className="py-20 text-center text-sm text-fg-faint">Loading post…</p>
    );
  }

  if (isEdit && (postQ.isError || !postQ.data)) {
    return (
      <div className="mx-auto max-w-lg px-4 py-20 text-center">
        <p className="text-sm text-fg-muted">
          {getApiErrorMessage(postQ.error) || "Post not found."}
        </p>
      </div>
    );
  }

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
            <h1 className="text-xl font-bold tracking-tight text-fg">
              {isEdit ? "Edit post" : "New post"}
            </h1>
            <p className="mt-1 text-sm text-fg-muted">
              {isEdit
                ? "Update the place, caption, tags, or media. New media uploads as soon as you pick it."
                : "Name the place, add media (uploads start immediately), write a caption, and pin it on the map."}
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
                  {" · "}
                  Max {formatMaxUploadLabel(maxUploadBytesRef.current)}
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
                      key={p.id}
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

                      {p.uploadStatus === "uploading" || p.uploadStatus === "pending" ? (
                        <div className="absolute inset-x-0 bottom-0 bg-black/70 px-1.5 py-1.5">
                          <div className="h-1.5 overflow-hidden rounded-full bg-white/20">
                            <div
                              className="h-full rounded-full bg-brand transition-[width] duration-150"
                              style={{ width: `${p.uploadProgress}%` }}
                            />
                          </div>
                          <p className="mt-1 text-center text-[9px] font-semibold text-white">
                            {p.uploadStatus === "pending"
                              ? "Starting…"
                              : `${p.uploadProgress}%`}
                          </p>
                        </div>
                      ) : null}

                      {p.uploadStatus === "ready" && p.kind === "image" ? (
                        <span className="absolute bottom-1 left-1 rounded bg-brand/90 px-1.5 py-0.5 text-[9px] font-bold text-black">
                          Ready
                        </span>
                      ) : null}

                      {p.uploadStatus === "error" ? (
                        <button
                          type="button"
                          onClick={() => retryUpload(p)}
                          className="absolute inset-x-1 bottom-1 rounded bg-rose-600/90 px-1.5 py-1 text-[9px] font-semibold text-white"
                        >
                          Retry
                        </button>
                      ) : null}

                      {(p.kind === "video" || p.kind === "audio") &&
                      (p.sourceFile || p.file) &&
                      p.uploadStatus === "ready" ? (
                        <button
                          type="button"
                          onClick={() =>
                            openClipStudio(
                              (p.sourceFile ?? p.file) as File,
                              p.kind as ClipKind
                            )
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
                        onClick={() => removePreview(p.id)}
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
            {uploadsInFlight ? (
              <p className="mb-2 text-center text-xs text-fg-faint">
                Uploading media… Publish unlocks when all files are ready.
              </p>
            ) : null}
            <button
              type="submit"
              disabled={busy || !canPublish}
              className="inline-flex h-11 w-full items-center justify-center rounded-xl bg-brand text-sm font-semibold text-black transition-colors hover:bg-brand-dark disabled:opacity-40"
            >
              {busy
                ? isEdit
                  ? "Saving…"
                  : "Publishing…"
                : isEdit
                  ? "Save changes"
                  : "Publish"}
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
          <Suspense
            fallback={
              <p className="py-20 text-center text-sm text-fg-faint">Loading…</p>
            }
          >
            <ComposeContent />
          </Suspense>
        </div>
      </main>
    </RequireAuth>
  );
}
