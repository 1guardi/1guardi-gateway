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

export default function App() {
  const [page, setPage] = useState<Page>('overview')
  const [selectedAgent, setSelectedAgent] = useState<string | 'all'>('all')

  const pages: Record<Page, ReactNode> = {
    overview: <Overview selectedAgent={selectedAgent} />,
    traces: <Traces selectedAgent={selectedAgent} />,
    guardrails: <Guardrails selectedAgent={selectedAgent} />,
    'pii-vault': <PiiVault selectedAgent={selectedAgent} />,
    router: <Router selectedAgent={selectedAgent} />,
    'api-keys': <APIKeys selectedAgent={selectedAgent} />,
  }

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
        <Sidebar current={page} onChange={setPage} agent={selectedAgent} onAgentChange={setSelectedAgent} />
        <main className="relative flex-1 overflow-y-auto min-h-screen">
          {pages[page]}
        </main>
      </div>
    </TooltipProvider>
  )
}
