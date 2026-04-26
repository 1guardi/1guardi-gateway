import { useState, useEffect } from 'react'
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

export type Page = 'overview' | 'traces' | 'guardrails' | 'pii-vault' | 'router' | 'agents' | 'api-keys'

const COMING_SOON = import.meta.env.VITE_COMING_SOON !== 'false'

export const comingSoonPages: Set<Page> = COMING_SOON
  ? new Set(['overview', 'traces', 'guardrails', 'pii-vault'])
  : new Set()

export interface AgentSummary {
  ID: number
  Name: string
  Description: string
  CreatedAt: string
}

export default function App() {
  const [page, setPage] = useState<Page>('router')
  const [selectedAgent, setSelectedAgent] = useState<string>('all')
  const [tenantId, setTenantId] = useState<string | null>(null)
  const [tenantName, setTenantName] = useState<string>('—')
  const [agents, setAgents] = useState<AgentSummary[]>([])

  const refreshAgents = (id: string) => {
    fetch(`/api/v1/tenants/${id}/agents`)
      .then((r) => r.json())
      .then((data: AgentSummary[]) => setAgents(data))
      .catch(() => {})
  }

  useEffect(() => {
    fetch('/api/v1/tenants')
      .then((r) => r.json())
      .then((tenants: { ID: number; Name: string }[]) => {
        if (!tenants.length) return
        const t = tenants[0]
        const id = String(t.ID)
        setTenantId(id)
        setTenantName(t.Name)
        refreshAgents(id)
      })
      .catch(() => {})
  }, [])

  const cs = (page: Page, label: string, tag: string, node: ReactNode) =>
    comingSoonPages.has(page) ? <ComingSoon pageName={label} tag={tag} /> : node

  const pages: Record<Page, ReactNode> = {
    overview:    cs('overview',   'Overview',   'TOWER VIEW', <Overview selectedAgent={selectedAgent} />),
    traces:      cs('traces',     'Traces',     'FLIGHT LOG', <Traces selectedAgent={selectedAgent} />),
    guardrails:  cs('guardrails', 'Guardrails', 'AIRSPACE',   <Guardrails selectedAgent={selectedAgent} />),
    'pii-vault': cs('pii-vault',  'PII Vault',  'CARGO',      <PiiVault selectedAgent={selectedAgent} />),
    router:      <Router selectedAgent={selectedAgent} />,
    agents:      <Agents tenantId={tenantId} agents={agents} onAgentCreated={() => tenantId && refreshAgents(tenantId)} />,
    'api-keys':  <APIKeys selectedAgent={selectedAgent} tenantId={tenantId} agents={agents} />,
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
