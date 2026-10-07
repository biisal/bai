import { useState, useEffect } from 'react'
import { Search } from 'lucide-react'
import { Button } from './ui/button'
import { siteName } from '@/config'

const links = [
  { label: 'Home', href: '#home' },
  { label: 'Features', href: '#features' },
  { label: 'Sessions', href: '#sessions' },
  { label: 'Themes', href: '#themes' },
  { label: 'Docs', href: '#docs' },
]

export default function Navbar() {
  const [active, setActive] = useState('')

  useEffect(() => {
    const ids = links.map(l => l.href.slice(1))
    const els = ids.map(id => document.getElementById(id)).filter(Boolean) as HTMLElement[]

    const observer = new IntersectionObserver(
      entries => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            setActive(entry.target.id)
          }
        }
      },
      { rootMargin: '-40% 0px -55% 0px' },
    )

    els.forEach(el => observer.observe(el))
    return () => observer.disconnect()
  }, [])

  return (
    <nav className="fixed top-0 inset-x-0 z-50 flex backdrop-blur-sm items-center justify-between px-6 py-4">
      {/* Logo */}
      <div className="flex items-center gap-3">
        <svg
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-primary"
        >
          <path d="m9 18 6-6-6-6" />
        </svg>
        <span className="text-xl font-medium tracking-tight text-white">{siteName}</span>
      </div>

      {/* Links */}
      <div className="hidden md:flex items-center gap-8 text-sm">
        {links.map(link => (
          <a
            key={link.href}
            href={link.href}
            className={`transition-colors pb-1 ${
              active === link.href.slice(1)
                ? 'text-primary border-b border-primary'
                : 'text-muted-foreground hover:text-foreground border-b border-transparent'
            }`}
          >
            {link.label}
          </a>
        ))}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-6">
        <Button variant="ghost" size="icon" aria-label="Search">
          <Search className="w-5 h-5" />
        </Button>
        <Button variant="outline" asChild>
          <a href="https://github.com/biisal/bai">
            Get Started &rarr;
          </a>
        </Button>
      </div>
    </nav>
  )
}