import type { ReactNode } from 'react'
import { Database, GitBranch, Palette } from 'lucide-react'

function Card({ id, icon, title, children }: { id?: string; icon: ReactNode; title: string; children: ReactNode }) {
  return (
    <div id={id} className="rounded-sm border border-border bg-muted/30 backdrop-blur-sm p-4">
      <div className="flex items-center gap-2 mb-3 text-sm font-medium text-white">
        <span className="text-primary">{icon}</span>
        {title}
      </div>
      <div className="font-['JetBrains_Mono'] text-xs leading-5 text-muted-foreground">{children}</div>
    </div>
  )
}

const swatches = ['bg-background', 'bg-muted', 'bg-primary', 'bg-accent', 'bg-success', 'bg-warning']

export default function SideCards() {
  return (
    <div className="grid w-full gap-4 sm:grid-cols-3">
      <Card icon={<Database className="w-4 h-4" />} title="Sessions">
        <div className="flex items-center gap-2 text-white">
          <span className="w-1.5 h-1.5 rounded-full bg-success" />
          fix token validation
        </div>
        <div className="pl-3.5">add rate limiting</div>
        <div className="pl-3.5">split config loading</div>
      </Card>

      <Card icon={<Palette className="w-4 h-4" />} title="Themes">
        <div className="flex gap-1.5 mb-2">
          {swatches.map(c => (
            <span key={c} className={`h-5 w-5 rounded-sm border border-border ${c}`} />
          ))}
        </div>
        ~/.config/bai/themes/*.json
      </Card>

      <Card icon={<GitBranch className="w-4 h-4" />} title="Shadow Git">
        <div><span className="text-[#f9a8d4]">e4b2a1c</span> add rate limiting</div>
        <div><span className="text-[#f9a8d4]">f8c3d2e</span> refactor db pool</div>
        <div className="text-primary">.bai_git · repo untouched</div>
      </Card>
    </div>
  )
}
