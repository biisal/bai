import { siteName } from '@/config'

export default function Intro() {
  return (
    <section className="w-full min-h-[60vh] flex flex-col items-center justify-center px-4 py-16 text-center">
      <div className="inline-flex items-center gap-2 px-3 py-1 rounded-sm border border-border bg-muted/60 text-xs font-['JetBrains_Mono'] text-primary mb-6">
        <span className="w-2 h-2 rounded-sm bg-primary animate-pulse" />
        Intro Section Ready
      </div>
      <h1 className="text-4xl md:text-6xl font-bold tracking-tight text-white mb-4">
        {siteName}
      </h1>
      <p className="text-muted-foreground max-w-xl text-base md:text-lg">
        Waiting for your design reference image...
      </p>
    </section>
  )
}
