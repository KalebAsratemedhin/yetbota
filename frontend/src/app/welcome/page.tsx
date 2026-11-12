"use client";

import {
  Navbar,
  HeroSection,
  DiscoverySection,
  ChampionsSection,
  CommunityQaSection,
  AssistantSection,
  GamificationSection,
  AccuracySection,
  Footer,
} from "@/components/landing";

/** Marketing / about landing. Product home stays at `/`. */
export default function WelcomePage() {
  return (
    <main className="min-h-screen overflow-x-clip bg-bg">
      <Navbar />
      <HeroSection />
      <DiscoverySection />
      <ChampionsSection />
      <CommunityQaSection />
      <AssistantSection />
      <GamificationSection />
      <AccuracySection />
      <Footer />
    </main>
  );
}
