import type { CloudinaryUploadResult, MediaKind, SignMediaResponse } from "@/lib/content/types";

/** Upload a file to Cloudinary using signed params from content-service. */
export async function uploadToCloudinary(
  file: File,
  sign: SignMediaResponse,
  kind: MediaKind
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

  const res = await fetch(endpoint, { method: "POST", body: form });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Cloudinary upload failed (${res.status})`);
  }
  return (await res.json()) as CloudinaryUploadResult;
}

export function parseHashtags(raw: string): string[] {
  const parts = raw
    .split(/[\s,]+/)
    .map((t) => t.replace(/^#/, "").trim().toLowerCase())
    .filter(Boolean);
  return [...new Set(parts)];
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
