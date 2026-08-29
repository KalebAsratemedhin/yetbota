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

export default function HomePage() {
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
