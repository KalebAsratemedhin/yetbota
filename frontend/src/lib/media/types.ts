/** Shared media-clip domain types (UI + encoders). */

export type ClipKind = "video" | "audio";

export type ClipRange = {
  /** Inclusive start, seconds */
  start: number;
  /** Exclusive-ish end, seconds */
  end: number;
};

export type ClipProgress = {
  phase: "loading_engine" | "reading" | "encoding" | "done";
  /** 0–1 when known */
  ratio?: number;
  message?: string;
};

export type ClipResult = {
  file: File;
  range: ClipRange;
  durationSec: number;
  kind: ClipKind;
};

export type ClipEncodeOptions = {
  maxDurationSec: number;
  onProgress?: (p: ClipProgress) => void;
  signal?: AbortSignal;
};
