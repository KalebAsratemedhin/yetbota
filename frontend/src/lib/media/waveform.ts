/**
 * Downsample an audio file into peak bars for a lightweight waveform UI.
 * Pure Web Audio — no ffmpeg.
 */
export async function buildWaveformPeaks(
  file: File,
  barCount = 64
): Promise<number[]> {
  const ctx = new AudioContext();
  try {
    const buf = await file.arrayBuffer();
    const audio = await ctx.decodeAudioData(buf.slice(0));
    const channel = audio.getChannelData(0);
    const block = Math.max(1, Math.floor(channel.length / barCount));
    const peaks: number[] = [];
    for (let i = 0; i < barCount; i++) {
      const start = i * block;
      let max = 0;
      for (let j = start; j < start + block && j < channel.length; j++) {
        const v = Math.abs(channel[j]!);
        if (v > max) max = v;
      }
      peaks.push(max);
    }
    const peak = Math.max(...peaks, 0.01);
    return peaks.map((p) => p / peak);
  } finally {
    await ctx.close().catch(() => undefined);
  }
}
