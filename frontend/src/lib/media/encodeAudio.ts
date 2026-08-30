import { rangeDuration } from "@/lib/media/range";
import type { ClipEncodeOptions, ClipRange, ClipResult } from "@/lib/media/types";

/** Write a 16-bit PCM WAV from an AudioBuffer (no ffmpeg). */
function audioBufferToWav(buffer: AudioBuffer): Blob {
  const numChannels = buffer.numberOfChannels;
  const sampleRate = buffer.sampleRate;
  const format = 1; // PCM
  const bitDepth = 16;
  const samples = buffer.length;
  const blockAlign = (numChannels * bitDepth) / 8;
  const byteRate = sampleRate * blockAlign;
  const dataSize = samples * blockAlign;
  const headerSize = 44;
  const array = new ArrayBuffer(headerSize + dataSize);
  const view = new DataView(array);

  writeString(view, 0, "RIFF");
  view.setUint32(4, 36 + dataSize, true);
  writeString(view, 8, "WAVE");
  writeString(view, 12, "fmt ");
  view.setUint32(16, 16, true);
  view.setUint16(20, format, true);
  view.setUint16(22, numChannels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, byteRate, true);
  view.setUint16(32, blockAlign, true);
  view.setUint16(34, bitDepth, true);
  writeString(view, 36, "data");
  view.setUint32(40, dataSize, true);

  const channels: Float32Array[] = [];
  for (let c = 0; c < numChannels; c++) {
    channels.push(buffer.getChannelData(c));
  }

  let offset = 44;
  for (let i = 0; i < samples; i++) {
    for (let c = 0; c < numChannels; c++) {
      const sample = Math.max(-1, Math.min(1, channels[c]![i]!));
      view.setInt16(
        offset,
        sample < 0 ? sample * 0x8000 : sample * 0x7fff,
        true
      );
      offset += 2;
    }
  }

  return new Blob([array], { type: "audio/wav" });
}

function writeString(view: DataView, offset: number, str: string) {
  for (let i = 0; i < str.length; i++) {
    view.setUint8(offset + i, str.charCodeAt(i));
  }
}

/**
 * Clip audio with Web Audio (decode → slice → WAV).
 * Avoids ffmpeg.wasm codecs that are often missing (e.g. libmp3lame).
 */
export async function encodeAudioClip(
  file: File,
  range: ClipRange,
  options: ClipEncodeOptions
): Promise<ClipResult> {
  const { maxDurationSec, onProgress, signal } = options;
  const dur = rangeDuration(range);
  if (dur <= 0) throw new Error("Clip range is empty");
  if (dur > maxDurationSec + 0.05) {
    throw new Error(`Clip must be ≤ ${maxDurationSec}s`);
  }
  if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

  onProgress?.({ phase: "reading", message: "Reading audio…" });
  const ctx = new AudioContext();
  try {
    const raw = await file.arrayBuffer();
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");
    const decoded = await ctx.decodeAudioData(raw.slice(0));

    const sampleRate = decoded.sampleRate;
    const startSample = Math.max(0, Math.floor(range.start * sampleRate));
    const endSample = Math.min(
      decoded.length,
      Math.max(startSample + 1, Math.floor(range.end * sampleRate))
    );
    const frameCount = endSample - startSample;
    if (frameCount < 1) throw new Error("Clip range is empty");

    onProgress?.({ phase: "encoding", ratio: 0.2, message: "Cutting audio…" });

    const offline = new OfflineAudioContext(
      decoded.numberOfChannels,
      frameCount,
      sampleRate
    );
    const sliced = offline.createBuffer(
      decoded.numberOfChannels,
      frameCount,
      sampleRate
    );
    for (let c = 0; c < decoded.numberOfChannels; c++) {
      const src = decoded.getChannelData(c).subarray(startSample, endSample);
      sliced.copyToChannel(src, c);
    }

    const source = offline.createBufferSource();
    source.buffer = sliced;
    source.connect(offline.destination);
    source.start(0);

    onProgress?.({ phase: "encoding", ratio: 0.6, message: "Rendering clip…" });
    const rendered = await offline.startRendering();
    if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

    const wav = audioBufferToWav(rendered);
    const name = file.name.replace(/\.[^.]+$/, "") + "-clip.wav";
    const out = new File([wav], name, { type: "audio/wav" });
    const lengthSec = frameCount / sampleRate;

    onProgress?.({ phase: "done", ratio: 1, message: "Done" });
    return {
      file: out,
      range: {
        start: startSample / sampleRate,
        end: startSample / sampleRate + lengthSec,
      },
      durationSec: lengthSec,
      kind: "audio",
    };
  } finally {
    await ctx.close().catch(() => undefined);
  }
}
