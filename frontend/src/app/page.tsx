"use client";

import { type ReactNode } from "react";
import {
  Navbar,
  HeroSection,
  DiscoverySection,
  AssistantSection,
  NavigationSection,
  AccuracySection,
  GamificationSection,
  PrinciplesSection,
  ChampionsSection,
  Footer,
} from "@/components/landing";
import HomeFeed from "@/components/feed/HomeFeed";
import { useHydrated } from "@/hooks/use-hydrated";
import { useAppSelector } from "@/lib/hooks";

function MarketingHome() {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg">
      <Navbar />
      <HeroSection />
      <PrinciplesSection />
      <NavigationSection />
      <AccuracySection />
      <GamificationSection />
      <AssistantSection />
      <DiscoverySection />
      <ChampionsSection />
      <Footer />
    </main>
  );
}

function FeedHome() {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg text-fg">
      <Navbar />
      <div className="pt-20">
        <HomeFeed />
      </div>
    </main>
  );
}

function HomeShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg text-fg">
      <Navbar />
      {children}
    </main>
  );
}

export default function HomePage() {
  const accessToken = useAppSelector((s) => s.auth.accessToken);
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

  if (accessToken) {
    return <FeedHome />;
  }

  return <MarketingHome />;
}
