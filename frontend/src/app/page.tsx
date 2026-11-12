"use client";

import { type ReactNode } from "react";
import { Navbar } from "@/components/landing";
import HomeFeed from "@/components/feed/HomeFeed";
import { useHydrated } from "@/hooks/use-hydrated";

function HomeShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg text-fg">
      <Navbar />
      {children}
    </main>
  );
}

/** Home feed is public; guests get anonymous /feed (recent → suggested). */
export default function HomePage() {
  const hydrated = useHydrated();

  if (!hydrated) {
    return (
      <HomeShell>
        <div className="flex min-h-[50vh] items-center justify-center pt-20 text-sm text-fg-faint">
          Loading…
        </div>
      </HomeShell>
    );
  }

  return (
    <HomeShell>
      <div className="pt-20">
        <HomeFeed />
      </div>
    </HomeShell>
  );
}
