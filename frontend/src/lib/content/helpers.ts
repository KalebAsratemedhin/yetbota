import type { CloudinaryUploadResult, MediaKind, SignMediaResponse } from "@/lib/content/types";

/** Default matches content-service / Cloudinary Free plan when sign has not returned yet. */
export const DEFAULT_MAX_UPLOAD_BYTES = 10 * 1024 * 1024;

export function formatMaxUploadLabel(bytes: number): string {
  const mb = bytes / (1024 * 1024);
  if (!Number.isFinite(mb) || mb <= 0) return "10 MB";
  return Number.isInteger(mb) ? `${mb} MB` : `${mb.toFixed(1)} MB`;
}

export type CloudinaryUploadOptions = {
  onProgress?: (percent: number) => void;
  signal?: AbortSignal;
};

/** Upload a file to Cloudinary using signed params from content-service. */
export function uploadToCloudinary(
  file: File,
  sign: SignMediaResponse,
  kind: MediaKind,
  opts?: CloudinaryUploadOptions
): Promise<CloudinaryUploadResult> {
  const form = new FormData();
  form.append("file", file);
  form.append("api_key", sign.api_key);
  form.append("timestamp", String(sign.timestamp));
  form.append("signature", sign.signature);
  form.append("folder", sign.folder);
  form.append("public_id", sign.public_id);

  const resourceType = sign.resource_type || (kind === "image" ? "image" : "video");
  const endpoint = `https://api.cloudinary.com/v1_1/${encodeURIComponent(sign.cloud_name)}/${resourceType}/upload`;

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("POST", endpoint);

    const onAbort = () => {
      xhr.abort();
      reject(new DOMException("Upload aborted", "AbortError"));
    };
    if (opts?.signal) {
      if (opts.signal.aborted) {
        onAbort();
        return;
      }
      opts.signal.addEventListener("abort", onAbort, { once: true });
    }

    xhr.upload.onprogress = (e) => {
      if (!opts?.onProgress || !e.lengthComputable || e.total <= 0) return;
      opts.onProgress(Math.min(100, Math.round((e.loaded / e.total) * 100)));
    };

    xhr.onload = () => {
      opts?.signal?.removeEventListener("abort", onAbort);
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          resolve(JSON.parse(xhr.responseText) as CloudinaryUploadResult);
        } catch {
          reject(new Error("Invalid Cloudinary response"));
        }
        return;
      }
      reject(
        new Error(xhr.responseText || `Cloudinary upload failed (${xhr.status})`)
      );
    };

    xhr.onerror = () => {
      opts?.signal?.removeEventListener("abort", onAbort);
      reject(new Error("Cloudinary upload failed (network error)"));
    };

    xhr.send(form);
  });
}

/** Split on spaces, commas, or `#` so `#a#b` and `#a #b` both work. */
export function parseHashtags(raw: string): string[] {
  const parts = raw
    .split(/[\s,#]+/)
    .map((t) => t.trim().toLowerCase())
    .filter(Boolean);
  return [...new Set(parts)];
}

/** Place title for cards; falls back if older payloads omit title. */
export function placeTitle(post: { title?: string; caption?: string }): string {
  const title = post.title?.trim();
  if (title) return title;
  const caption = post.caption?.trim();
  if (caption) return caption;
  return "Untitled place";
}

export function coverUrl(post: { media: { kind: string; url: string; sort_order: number }[] }): string | null {
  const sorted = [...post.media].sort((a, b) => a.sort_order - b.sort_order);
  const image = sorted.find((m) => m.kind === "image");
  if (image) return image.url;
  const video = sorted.find((m) => m.kind === "video");
  return video?.url ?? null;
}

export function formatCoord(lat: number, lng: number): string {
  return `${lat.toFixed(4)}°, ${lng.toFixed(4)}°`;
}

export function relativeTime(iso: string): string {
  const t = new Date(iso).getTime();
  if (Number.isNaN(t)) return "";
  const diff = Date.now() - t;
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  if (d < 30) return `${d}d ago`;
  return new Date(iso).toLocaleDateString();
}
