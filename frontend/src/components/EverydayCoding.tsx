import { Button } from './ui/button'
import { siteName } from '@/config'
import TerminalHeader from './ui/terminal-header'
import SideCards from './SideCards'

export default function EverydayCoding() {
  return (
    <section className="w-full max-w-7xl mx-auto px-6 py-32 flex flex-col lg:flex-row gap-16 relative">
      
      {/* Background Decorative Accents */}
      <div className="absolute top-1/2 left-0 -translate-y-1/2 -translate-x-1/4 w-[500px] h-[500px] bg-muted rounded-sm blur-[150px] opacity-80 pointer-events-none" />
      <div className="absolute bottom-0 right-0 translate-x-1/4 w-[400px] h-[400px] bg-primary rounded-sm blur-[180px] opacity-10 pointer-events-none" />

      {/* Left Copy */}
      <div className="w-full lg:w-1/3 flex flex-col justify-center z-10">
        <h2 className="text-5xl lg:text-6xl font-['Playfair_Display'] text-white leading-tight mb-6">
          Built for<br />everyday coding.
        </h2>
        <p className="text-muted-foreground text-lg leading-relaxed mb-10">
          Whether you're feeling stuck on a complex bug, tired of writing boilerplate, or just need a moment of focus — {siteName} is here to support your development workflow, anytime, anywhere.
        </p>
        <Button>
          Explore All Features &rarr;
        </Button>
      </div>

      {/* Right Grid / UI Showcase */}
      <div className="w-full lg:flex-1 flex flex-col gap-6 z-10">
        
        {/* Main Terminal Mockup Card */}
        <div className="flex-1 rounded-sm bg-muted/20 backdrop-blur-sm overflow-hidden">
          <TerminalHeader title="bai" />

          {/* Terminal body */}
          <div className="p-3 font-['JetBrains_Mono'] font-bold text-sm leading-6">
            {/* prompt + command */}
            <div className="mb-3">
              <span className="text-primary">~/my-project</span>
              <span className="text-muted-foreground"> ❯ </span>
              <span className="text-white/60">bai git log --oneline</span>
            </div>
            {/* git log output */}
            <div className="text-muted-foreground space-y-0.5">
              <div><span className="text-[#f9a8d4]">e4b2a1c</span> <span className="text-warning">feat:</span> <span className="text-white">Add rate limiting logic to API router</span></div>
              <div><span className="text-[#f9a8d4]">f8c3d2e</span> <span className="text-success">fix:</span> <span className="text-white">Refactor database connection pool settings</span></div>
              <div><span className="text-[#f9a8d4]">a1b2c3d</span> <span className="text-success">fix:</span> <span className="text-white">Null pointer in token validation</span></div>
              <div><span className="text-[#f9a8d4]">9d8c7b6</span> <span className="text-secondary">chore:</span> <span className="text-white">Initial scaffold of auth middleware</span></div>
              <div><span className="text-[#f9a8d4]">3a1f8cc</span> <span className="text-warning">feat:</span> <span className="text-white">Stream model response to stdout</span></div>
              <div><span className="text-[#f9a8d4]">bb02de1</span> <span className="text-destructive">refactor:</span> <span className="text-white">Split config loading into separate module</span></div>
            </div>
            {/* next prompt */}
            <div className="mt-3">
              <span className="text-primary">~/my-project</span>
              <span className="text-muted-foreground"> ❯ </span>
              <span className="w-2 h-[1em] bg-muted-foreground/60 animate-pulse inline-block align-middle" />
            </div>
          </div>
        </div>

        {/* Side Stacked Cards */}
        <SideCards />


      </div>
    </section>
  )
}
