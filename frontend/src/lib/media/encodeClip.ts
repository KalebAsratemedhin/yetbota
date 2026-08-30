import { fetchFile } from "@ffmpeg/util";
import { encodeAudioClip } from "@/lib/media/encodeAudio";
import { getFFmpeg } from "@/lib/media/ffmpegLoader";
import { rangeDuration } from "@/lib/media/range";
import type { ClipEncodeOptions, ClipKind, ClipRange, ClipResult } from "@/lib/media/types";

function extFor(file: File, kind: ClipKind): string {
  const fromName = file.name.split(".").pop()?.toLowerCase();
  if (fromName && /^[a-z0-9]{2,5}$/.test(fromName)) return fromName;
  if (kind === "audio") {
    if (file.type.includes("mpeg")) return "mp3";
    if (file.type.includes("wav")) return "wav";
    if (file.type.includes("ogg")) return "ogg";
    return "m4a";
  }
  if (file.type.includes("webm")) return "webm";
  if (file.type.includes("quicktime")) return "mov";
  return "mp4";
}

function toFile(data: Uint8Array | string, name: string, mime: string): File {
  const raw =
    typeof data === "string" ? new TextEncoder().encode(data) : data;
  const copy = new Uint8Array(raw.byteLength);
  copy.set(raw);
  return new File([copy], name, { type: mime });
}

async function execVideoStrategies(
  ffmpeg: Awaited<ReturnType<typeof getFFmpeg>>,
  inputName: string,
  start: number,
  length: number
): Promise<{ outName: string; mime: string; fileName: string }> {
  const startS = start.toFixed(3);
  const lenS = length.toFixed(3);

  const strategies: { outName: string; mime: string; fileName: string; args: string[] }[] = [
    {
      outName: "clip.mp4",
      mime: "video/mp4",
      fileName: "clip.mp4",
      args: [
        "-i",
        inputName,
        "-ss",
        startS,
        "-t",
        lenS,
        "-c:v",
        "libx264",
        "-preset",
        "ultrafast",
        "-crf",
        "28",
        "-c:a",
        "aac",
        "-b:a",
        "128k",
        "-movflags",
        "+faststart",
        "-pix_fmt",
        "yuv420p",
        "clip.mp4",
      ],
    },
    {
      outName: "clip.mp4",
      mime: "video/mp4",
      fileName: "clip.mp4",
      args: [
        "-i",
        inputName,
        "-ss",
        startS,
        "-t",
        lenS,
        "-c:v",
        "mpeg4",
        "-q:v",
        "7",
        "-c:a",
        "aac",
        "-b:a",
        "128k",
        "clip.mp4",
      ],
    },
    {
      outName: "clip.webm",
      mime: "video/webm",
      fileName: "clip.webm",
      args: [
        "-i",
        inputName,
        "-ss",
        startS,
        "-t",
        lenS,
        "-c:v",
        "libvpx",
        "-b:v",
        "1M",
        "-c:a",
        "libvorbis",
        "clip.webm",
      ],
    },
    {
      // Last resort — keyframe-aligned stream copy
      outName: "clip-copy.mp4",
      mime: "video/mp4",
      fileName: "clip-copy.mp4",
      args: [
        "-ss",
        startS,
        "-i",
        inputName,
        "-t",
        lenS,
        "-c",
        "copy",
        "-avoid_negative_ts",
        "make_zero",
        "clip-copy.mp4",
      ],
    },
  ];

  let lastErr = "Encoding failed";
  for (const strategy of strategies) {
    try {
      // Clear previous outputs if any
      try {
        await ffmpeg.deleteFile(strategy.outName);
      } catch {
        /* ignore */
      }
      const code = await ffmpeg.exec(strategy.args);
      if (code === 0) {
        return {
          outName: strategy.outName,
          mime: strategy.mime,
          fileName: strategy.fileName,
        };
      }
      lastErr = `Encoder exited with code ${code}`;
    } catch (err) {
      lastErr = err instanceof Error ? err.message : "Encoding failed";
    }
  }
  throw new Error(lastErr);
}

/**
 * Encode a clipped segment.
 * Audio → Web Audio WAV. Video → ffmpeg.wasm with codec fallbacks.
 */
export async function encodeClip(
  file: File,
  kind: ClipKind,
  range: ClipRange,
  options: ClipEncodeOptions
): Promise<ClipResult> {
  if (kind === "audio") {
    return encodeAudioClip(file, range, options);
  }

  const { maxDurationSec, onProgress, signal } = options;
  const dur = rangeDuration(range);
  if (dur <= 0) throw new Error("Clip range is empty");
  if (dur > maxDurationSec + 0.05) {
    throw new Error(`Clip must be ≤ ${maxDurationSec}s`);
  }
  if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

  onProgress?.({ phase: "loading_engine", message: "Loading clip engine…" });
  const ffmpeg = await getFFmpeg();
  if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

  const onProg = ({ progress }: { progress: number }) => {
    onProgress?.({
      phase: "encoding",
      ratio: Math.min(1, Math.max(0, progress)),
      message: "Encoding clip…",
    });
  };
  ffmpeg.on("progress", onProg);

  const inExt = extFor(file, "video");
  const inputName = `input.${inExt}`;
  const start = Math.max(0, range.start);
  const length = Math.max(0.05, dur);

  try {
    onProgress?.({ phase: "reading", message: "Reading media…" });
    await ffmpeg.writeFile(inputName, await fetchFile(file));
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    onProgress?.({ phase: "encoding", ratio: 0, message: "Encoding clip…" });
    const out = await execVideoStrategies(ffmpeg, inputName, start, length);
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    const data = await ffmpeg.readFile(out.outName);
    const base = file.name.replace(/\.[^.]+$/, "") || "video";
    const clipped = toFile(
      data instanceof Uint8Array ? data : String(data),
      `${base}-${out.fileName}`,
      out.mime
    );

    onProgress?.({ phase: "done", ratio: 1, message: "Done" });
    return {
      file: clipped,
      range: { start, end: start + length },
      durationSec: length,
      kind: "video",
    };
  } finally {
    try {
      ffmpeg.off("progress", onProg);
    } catch {
      /* older builds */
    }
    for (const name of [
      inputName,
      "clip.mp4",
      "clip.webm",
      "clip-copy.mp4",
    ]) {
      try {
        await ffmpeg.deleteFile(name);
      } catch {
        /* ignore */
      }
    }
  }
}

/**
 * If the source is already within max duration and the selection is the full
 * file, skip re-encode (fast path for short takes).
 */
export async function encodeClipOrPassthrough(
  file: File,
  kind: ClipKind,
  range: ClipRange,
  sourceDurationSec: number,
  options: ClipEncodeOptions
): Promise<ClipResult> {
  const coversWhole =
    range.start <= 0.05 && range.end >= sourceDurationSec - 0.05;
  const shortEnough = sourceDurationSec <= options.maxDurationSec + 0.25;

  if (coversWhole && shortEnough) {
    options.onProgress?.({ phase: "done", ratio: 1, message: "Ready" });
    return {
      file,
      range: { start: 0, end: sourceDurationSec },
      durationSec: sourceDurationSec,
      kind,
    };
  }

  return encodeClip(file, kind, range, options);
}
