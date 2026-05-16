import { useEffect, useState } from 'react'
import type { ReactNode } from 'react'
import { TooltipProvider } from '@/components/ui/tooltip'
import Sidebar from './components/Sidebar.tsx'
import ComingSoon from './components/ComingSoon.tsx'
import Login from './pages/Login.tsx'
import Overview from './pages/Overview.tsx'
import Traces from './pages/Traces.tsx'
import Guardrails from './pages/Guardrails.tsx'
import PiiVault from './pages/PiiVault.tsx'
import Router from './pages/Router.tsx'
import APIKeys from './pages/APIKeys.tsx'
import Members from './pages/Members.tsx'
import Agents from './pages/Agents.tsx'
import Upstreams from './pages/Upstreams.tsx'
import Tenants from './pages/Tenants.tsx'
import Onboarding from './pages/Onboarding.tsx'
import { useTenants } from './api/tenants.ts'
import { useAgents } from './api/agents.ts'
import { jwtDecode } from 'jwt-decode'
import { ADMIN_LOGOUT_EVENT } from './api/auth-storage'
import { consumeAuthCallback } from './api/oidc'

export type Page = 'overview' | 'traces' | 'guardrails' | 'pii-vault' | 'router' | 'agents' | 'api-keys' | 'upstreams' | 'tenants' | 'members'

interface JWTPayload {
  is_super_admin: boolean
  [key: string]: any
}

const COMING_SOON = import.meta.env.VITE_COMING_SOON !== 'false'

export const comingSoonPages: Set<Page> = COMING_SOON
  ? new Set(['overview', 'pii-vault'])
  : new Set()

function isTokenValid(token: string | null) {
  if (!token) {
    return false
  }

  try {
    const payload = jwtDecode<{ exp?: number }>(token)
    return !payload.exp || payload.exp * 1000 > Date.now()
  } catch {
    return false
  }
}

export default function App() {
  // Handle the OIDC redirect (/auth/callback#token=...) once, before first render.
  const [ssoCallback] = useState(() => consumeAuthCallback())
  const [isAuthenticated, setIsAuthenticated] = useState(
    () => ssoCallback.ok || isTokenValid(localStorage.getItem('admin_token')),
  )
  const [page, setPage] = useState<Page>('router')
  const [selectedAgent, setSelectedAgent] = useState<string>('all')
  const [activeTenantId, setActiveTenantId] = useState<string | null>(null)

  const token = localStorage.getItem('admin_token')
  const isSuperAdmin = token && isTokenValid(token)
    ? jwtDecode<JWTPayload>(token).is_super_admin
    : false

  const { data: tenants, isLoading: tenantsLoading } = useTenants()
  const tenant = activeTenantId
    ? tenants?.find(t => String(t.ID) === activeTenantId) ?? tenants?.[0]
    : tenants?.[0]
  const tenantId = tenant ? String(tenant.ID) : null
  const tenantName = tenant?.Name || '—'

  const { data: agents = [] } = useAgents(tenantId)

  const logout = () => {
    localStorage.removeItem('admin_token')
    setIsAuthenticated(false)
  }

  useEffect(() => {
    const handleLogout = () => setIsAuthenticated(false)
    window.addEventListener(ADMIN_LOGOUT_EVENT, handleLogout)
    return () => window.removeEventListener(ADMIN_LOGOUT_EVENT, handleLogout)
  }, [])

  if (!isAuthenticated) {
    return <Login onLogin={() => setIsAuthenticated(true)} ssoError={ssoCallback.error} />
  }

  // A user who belongs to no tenant must create an organization first.
  if (tenantsLoading) {
    return (
      <div className="flex-1 min-h-screen flex items-center justify-center bg-background">
        <p className="font-mono text-[10px] tracking-widest text-muted-foreground">LOADING...</p>
      </div>
    )
  }
  if (!tenants || tenants.length === 0) {
    return <Onboarding />
  }

  const cs = (page: Page, label: string, tag: string, node: ReactNode) =>
    comingSoonPages.has(page) ? <ComingSoon pageName={label} tag={tag} /> : node

  const pages: Record<Page, ReactNode> = {
    overview:    cs('overview',   'Overview',   'TOWER VIEW', <Overview selectedAgent={selectedAgent} />),
    traces:      cs('traces',     'Traces',     'FLIGHT LOG', <Traces selectedAgent={selectedAgent} tenantId={tenantId} />),
    guardrails:  cs('guardrails', 'Guardrails', 'AIRSPACE',   <Guardrails selectedAgent={selectedAgent} tenantId={tenantId} />),
    'pii-vault': cs('pii-vault',  'PII Vault',  'CARGO',      <PiiVault selectedAgent={selectedAgent} />),
    router:      <Router selectedAgent={selectedAgent} tenantId={tenantId} />,
    agents:      <Agents tenantId={tenantId} agents={agents} onViewTraces={(id) => { setSelectedAgent(id); setPage('traces') }} />,
    'api-keys':  <APIKeys selectedAgent={selectedAgent} tenantId={tenantId} agents={agents} />,
    members:     <Members tenantId={tenantId} />,
    upstreams:   <Upstreams tenantId={tenantId} />,
    tenants:     <Tenants activeTenantId={activeTenantId} onTenantSelect={(id) => { setActiveTenantId(id); setPage('router') }} />,
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
          onLogout={logout}
          isSuperAdmin={isSuperAdmin}
          tenantCount={tenants?.length || 0}
        />
        <main className="relative flex-1 overflow-y-auto min-h-screen">
          {pages[page]}
        </main>
      </div>
    </TooltipProvider>
  )
}
