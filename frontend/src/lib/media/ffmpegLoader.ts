import { FFmpeg } from "@ffmpeg/ffmpeg";
import { toBlobURL } from "@ffmpeg/util";
import { CLIP_ENGINE_CDN } from "@/lib/media/constants";

let ffmpegSingleton: FFmpeg | null = null;
let loadPromise: Promise<FFmpeg> | null = null;

/**
 * Lazily load a single shared FFmpeg instance (browser only).
 * Uses single-thread core via blob URLs — no COOP/COEP headers required.
 */
export async function getFFmpeg(
  onLog?: (message: string) => void
): Promise<FFmpeg> {
  if (typeof window === "undefined") {
    throw new Error("FFmpeg is only available in the browser");
  }
  if (ffmpegSingleton?.loaded) return ffmpegSingleton;
  if (loadPromise) return loadPromise;

  loadPromise = (async () => {
    const ffmpeg = new FFmpeg();
    if (onLog) {
      ffmpeg.on("log", ({ message }) => onLog(message));
    }
    await ffmpeg.load({
      coreURL: await toBlobURL(`${CLIP_ENGINE_CDN}/ffmpeg-core.js`, "text/javascript"),
      wasmURL: await toBlobURL(`${CLIP_ENGINE_CDN}/ffmpeg-core.wasm`, "application/wasm"),
    });
    ffmpegSingleton = ffmpeg;
    return ffmpeg;
  })();

  try {
    return await loadPromise;
  } catch (err) {
    loadPromise = null;
    ffmpegSingleton = null;
    throw err;
  }
}
