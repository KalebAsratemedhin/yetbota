"use client";

import { Moon, Sun } from "lucide-react";
import { useTheme } from "@/hooks/use-theme";

export default function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  return (
    <div
      role="radiogroup"
      aria-label="Theme"
      className="inline-flex items-center rounded-full border border-border-subtle bg-surface-2 p-1"
    >
      <button
        type="button"
        role="radio"
        aria-checked={theme === "light"}
        onClick={() => setTheme("light")}
        className={`inline-flex h-9 items-center gap-2 rounded-full px-4 text-sm font-semibold transition-colors ${
          theme === "light"
            ? "bg-brand text-black shadow-sm"
            : "text-fg-muted hover:text-fg"
        }`}
      >
        <Sun className="h-4 w-4" />
        Light
      </button>
      <button
        type="button"
        role="radio"
        aria-checked={theme === "dark"}
        onClick={() => setTheme("dark")}
        className={`inline-flex h-9 items-center gap-2 rounded-full px-4 text-sm font-semibold transition-colors ${
          theme === "dark"
            ? "bg-brand text-black shadow-sm"
            : "text-fg-muted hover:text-fg"
        }`}
      >
        <Moon className="h-4 w-4" />
        Dark
      </button>
    </div>
  );
}
