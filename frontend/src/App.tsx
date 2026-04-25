import { useState } from 'react'
import type { ReactNode } from 'react'
import { TooltipProvider } from '@/components/ui/tooltip'
import Sidebar from './components/Sidebar.tsx'
import Overview from './pages/Overview.tsx'
import Traces from './pages/Traces.tsx'
import Guardrails from './pages/Guardrails.tsx'
import PiiVault from './pages/PiiVault.tsx'
import Router from './pages/Router.tsx'
import APIKeys from './pages/APIKeys.tsx'

export type Page = 'overview' | 'traces' | 'guardrails' | 'pii-vault' | 'router' | 'api-keys'

const pages: Record<Page, ReactNode> = {
  overview: <Overview />,
  traces: <Traces />,
  guardrails: <Guardrails />,
  'pii-vault': <PiiVault />,
  router: <Router />,
  'api-keys': <APIKeys />,
}

export default function App() {
  const [page, setPage] = useState<Page>('overview')

  return (
    <TooltipProvider>
      <div className="flex min-h-screen w-full bg-background text-foreground">
        {/* Grid overlay */}
        <div
          className="fixed inset-0 pointer-events-none"
          style={{
            backgroundImage:
              'linear-gradient(var(--grid-color) 1px, transparent 1px), linear-gradient(90deg, var(--grid-color) 1px, transparent 1px)',
            backgroundSize: '48px 48px',
          }}
        />
        <Sidebar current={page} onChange={setPage} />
        <main className="relative flex-1 overflow-y-auto min-h-screen">
          {pages[page]}
        </main>
      </div>
    </TooltipProvider>
  )
}
