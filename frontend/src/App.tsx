import Navbar from '@/components/Navbar'
import Hero from '@/components/Hero'
import TuiPreview from '@/components/TuiPreview'
import FeatureBar from '@/components/FeatureBar'
import EverydayCoding from '@/components/EverydayCoding'
import WhatWeSkip from '@/components/WhatWeSkip'
import UndoSection from '@/components/UndoSection'
import FAQ from '@/components/FAQ'
import FooterShowcase from '@/components/FooterShowcase'

export default function App() {
  return (
    <div className="min-h-screen bg-background text-foreground antialiased selection:bg-primary/20 selection:text-primary overflow-x-hidden">
      <Navbar />
      <main>
        <section id="home"><Hero /></section>
        <TuiPreview />
        <section id="features"><FeatureBar /></section>
        <section id="sessions"><EverydayCoding /></section>
        <WhatWeSkip />
        <UndoSection />
        <FAQ />
        <FooterShowcase />
      </main>
    </div>
  )
}
