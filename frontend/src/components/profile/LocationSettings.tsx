"use client";

import { MapPin } from "lucide-react";
import { toast } from "sonner";
import { useLocationPreference } from "@/hooks/use-location-preference";
import { ProfileSection } from "@/components/profile/ProfileSection";
import { cn } from "@/lib/utils";

export default function LocationSettings({ framed = true }: { framed?: boolean }) {
  const { enabled, setEnabled } = useLocationPreference();

  function toggle(next: boolean) {
    setEnabled(next);
    if (next) {
      toast.message("Location on", {
        description: "Nearby posts will use your location when the browser allows it.",
      });
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          () => undefined,
          () => {
            toast.message("Location blocked", {
              description: "Allow location in browser settings, or turn this off anytime.",
            });
          },
          { enableHighAccuracy: true, timeout: 12_000, maximumAge: 60_000 }
        );
      }
    } else {
      toast.message("Location off", {
        description: "Your feed won’t use nearby ranking on this device.",
      });
    }
  }

  const body = (
    <div className="space-y-5">
      <p className="text-sm leading-relaxed text-fg-muted">
        When on, the home feed can mix in nearby places. Stored on this device only — we don’t
        sync this preference to the server.
      </p>
      <div className="flex items-center justify-between gap-4 rounded-2xl border border-border-subtle bg-bg px-4 py-3.5">
        <div className="flex min-w-0 items-start gap-3">
          <span className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl border border-brand/25 bg-brand/10 text-brand">
            <MapPin className="h-4 w-4" />
          </span>
          <div className="min-w-0">
            <p className="text-sm font-semibold text-fg">Use location for nearby</p>
            <p className="mt-0.5 text-xs text-fg-faint">
              {enabled ? "Enabled on this browser" : "Disabled — feed uses following / recent"}
            </p>
          </div>
        </div>
        <button
          type="button"
          role="switch"
          aria-checked={enabled}
          onClick={() => toggle(!enabled)}
          className={cn(
            "relative h-7 w-12 shrink-0 rounded-full transition-colors",
            enabled ? "bg-brand" : "bg-overlay ring-1 ring-border-subtle"
          )}
        >
          <span
            className={cn(
              "absolute top-0.5 left-0.5 h-6 w-6 rounded-full bg-white shadow transition-transform",
              enabled && "translate-x-5"
            )}
          />
        </button>
      </div>
    </div>
  );

  if (!framed) return body;

  return (
    <ProfileSection
      title="Location"
      description="Nearby feed ranking on this device."
    >
      {body}
    </ProfileSection>
  );
}
