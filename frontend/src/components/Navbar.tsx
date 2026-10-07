import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Menu, X } from 'lucide-react'
import { Button } from './ui/button'
import { siteName } from '@/config'

const links = [
  { label: 'Home', href: '#home' },
  { label: 'Features', href: '#features' },
  { label: 'Sessions', href: '#sessions' },
  { label: 'Tools', href: '#tools' },
  { label: 'Themes', href: '#themes' },
]

export default function Navbar() {
  const [active, setActive] = useState('')
  const [open, setOpen] = useState(false)

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

  const close = () => setOpen(false)

  return (
    <nav className="fixed top-0 inset-x-0 z-50 backdrop-blur-sm">
      <div className="flex items-center justify-between px-6 py-4">
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
          <Link
            to="/docs"
            className="border-b border-transparent pb-1 text-muted-foreground transition-colors hover:text-foreground"
          >
            Docs
          </Link>
        </div>

        <div className="flex items-center gap-3">
          <Button variant="outline" className="hidden sm:inline-flex" render={<Link to="/docs" />}>
            Get Started &rarr;
          </Button>
          <button
            type="button"
            className="md:hidden flex h-9 w-9 items-center justify-center rounded-sm border border-border text-foreground"
            aria-expanded={open}
            aria-label={open ? 'Close menu' : 'Open menu'}
            onClick={() => setOpen(v => !v)}
          >
            {open ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
          </button>
        </div>
      </div>

      {open && (
        <div className="md:hidden border-t border-border bg-background/95 px-6 py-4">
          <div className="flex flex-col gap-4 text-sm">
            {links.map(link => (
              <a
                key={link.href}
                href={link.href}
                onClick={close}
                className={active === link.href.slice(1) ? 'text-primary' : 'text-muted-foreground'}
              >
                {link.label}
              </a>
            ))}
            <Link to="/docs" onClick={close} className="text-muted-foreground">
              Docs
            </Link>
            <Link to="/docs" onClick={close} className="text-primary sm:hidden">
              Get Started &rarr;
            </Link>
          </div>
        </div>
      )}
    </nav>
  )
}
