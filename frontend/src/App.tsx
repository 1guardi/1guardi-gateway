import { useState } from 'react'
import type { ReactNode } from 'react'
import { TooltipProvider } from '@/components/ui/tooltip'
import Sidebar from './components/Sidebar.tsx'
import ComingSoon from './components/ComingSoon.tsx'
import Overview from './pages/Overview.tsx'
import Traces from './pages/Traces.tsx'
import Guardrails from './pages/Guardrails.tsx'
import PiiVault from './pages/PiiVault.tsx'
import Router from './pages/Router.tsx'
import APIKeys from './pages/APIKeys.tsx'
import Agents from './pages/Agents.tsx'
import Upstreams from './pages/Upstreams.tsx'
import { useTenants } from './api/tenants.ts'
import { useAgents } from './api/agents.ts'

export type Page = 'overview' | 'traces' | 'guardrails' | 'pii-vault' | 'router' | 'agents' | 'api-keys' | 'upstreams'

const COMING_SOON = import.meta.env.VITE_COMING_SOON !== 'false'

export const comingSoonPages: Set<Page> = COMING_SOON
  ? new Set(['overview', 'traces', 'guardrails', 'pii-vault'])
  : new Set()

export default function App() {
  const [page, setPage] = useState<Page>('router')
  const [selectedAgent, setSelectedAgent] = useState<string>('all')

  const { data: tenants } = useTenants()
  const tenant = tenants?.[0]
  const tenantId = tenant ? String(tenant.ID) : null
  const tenantName = tenant?.Name || '—'

  const { data: agents = [] } = useAgents(tenantId)

  const cs = (page: Page, label: string, tag: string, node: ReactNode) =>
    comingSoonPages.has(page) ? <ComingSoon pageName={label} tag={tag} /> : node

  const pages: Record<Page, ReactNode> = {
    overview:    cs('overview',   'Overview',   'TOWER VIEW', <Overview selectedAgent={selectedAgent} />),
    traces:      cs('traces',     'Traces',     'FLIGHT LOG', <Traces selectedAgent={selectedAgent} />),
    guardrails:  cs('guardrails', 'Guardrails', 'AIRSPACE',   <Guardrails selectedAgent={selectedAgent} />),
    'pii-vault': cs('pii-vault',  'PII Vault',  'CARGO',      <PiiVault selectedAgent={selectedAgent} />),
    router:      <Router selectedAgent={selectedAgent} tenantId={tenantId} />,
    agents:      <Agents tenantId={tenantId} agents={agents} />,
    'api-keys':  <APIKeys selectedAgent={selectedAgent} tenantId={tenantId} agents={agents} />,
    upstreams:   <Upstreams tenantId={tenantId} />,
  }

  return (
    <TooltipProvider>
      <div className="flex min-h-screen w-full bg-background text-foreground">
        <div
          className="fixed inset-0 pointer-events-none"
          style={{
            backgroundImage:
              'linear-gradient(var(--grid-color) 1px, transparent 1px), linear-gradient(90deg, var(--grid-color) 1px, transparent 1px)',
            backgroundSize: '48px 48px',
          }}
        />
        <Sidebar
          current={page}
          onChange={setPage}
          agent={selectedAgent}
          onAgentChange={setSelectedAgent}
          tenantName={tenantName}
          agents={agents}
        />
        <main className="relative flex-1 overflow-y-auto min-h-screen">
          {pages[page]}
        </main>
      </div>
    </TooltipProvider>
  )
}
