"use client";

import { Suspense, type ReactNode } from "react";
import { Navbar } from "@/components/landing";
import ExploreHome from "@/components/feed/ExploreHome";
import { useHydrated } from "@/hooks/use-hydrated";

function HomeShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg text-fg">
      <Navbar />
      {children}
    </main>
  );
}

/** Home: Posts feed + Place Directory tabs; sidebar always visible. */
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
        <Suspense
          fallback={
            <div className="flex min-h-[40vh] items-center justify-center text-sm text-fg-faint">
              Loading…
            </div>
          }
        >
          <ExploreHome />
        </Suspense>
      </div>
    </HomeShell>
  );
}
