export { DEFAULT_MAX_CLIP_SEC } from "@/lib/media/constants";
export { readMediaDuration } from "@/lib/media/duration";
export {
  clampClipRange,
  formatClipTime,
  initialClipRange,
  rangeDuration,
} from "@/lib/media/range";
export { encodeAudioClip } from "@/lib/media/encodeAudio";
export { encodeClip, encodeClipOrPassthrough } from "@/lib/media/encodeClip";
export { buildWaveformPeaks } from "@/lib/media/waveform";
export type {
  ClipEncodeOptions,
  ClipKind,
  ClipProgress,
  ClipRange,
  ClipResult,
} from "@/lib/media/types";
