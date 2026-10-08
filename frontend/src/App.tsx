import Navbar from "@/components/Navbar";
import Hero from "@/components/Hero";
import TuiPreview from "@/components/TuiPreview";
import FeatureBar from "@/components/FeatureBar";
import WhatWeSkip from "@/components/WhatWeSkip";
import CustomTools from "@/components/CustomTools";
import Themes from "@/components/Themes";
import UndoSection from "@/components/UndoSection";
import FAQ from "@/components/FAQ";
import FooterShowcase from "@/components/FooterShowcase";

export default function App() {
  return (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/20 selection:text-primary overflow-x-hidden">
      <Navbar />
      <main>
        <section id="home">
          <Hero />
        </section>
        <TuiPreview />
        <section id="features">
          <FeatureBar />
        </section>
        <CustomTools />
        <WhatWeSkip />
        <Themes />
        <UndoSection />
        <FAQ />
        <FooterShowcase />
      </main>
    </div>
  );
}
