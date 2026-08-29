import en from "../../content/en.json";

export type LandingContent = typeof en;

/** English landing/auth copy for now. Amharic can plug in later. */
export function useContent(): LandingContent {
  return en;
}
