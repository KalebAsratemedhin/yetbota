"use client";

import ThemeToggle from "@/components/shared/ThemeToggle";
import { ProfileSection } from "@/components/profile/ProfileSection";

export default function AppearanceSettings({ framed = true }: { framed?: boolean }) {
  const body = (
    <div className="space-y-4">
      <p className="text-sm text-fg-muted">
        Dark is the default for Yet Bota. Switch anytime — we remember your choice locally.
      </p>
      <ThemeToggle />
    </div>
  );

  if (!framed) return body;

  return (
    <ProfileSection
      title="Appearance"
      description="Prefer light or dark. Saves on this device."
    >
      {body}
    </ProfileSection>
  );
}
