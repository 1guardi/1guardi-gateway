import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Sun, Moon, Monitor, Bot } from 'lucide-react'
import { useTheme } from './theme-provider.tsx'
import type { Page } from '../App.tsx'

interface NavItem { id: Page; label: string; tag: string; letter: string }

const navItems: NavItem[] = [
  { id: 'overview',   label: 'Overview',   tag: 'TOWER VIEW', letter: 'M' },
  { id: 'traces',     label: 'Traces',     tag: 'FLIGHT LOG', letter: 'T' },
  { id: 'guardrails', label: 'Guardrails', tag: 'AIRSPACE',   letter: 'G' },
  { id: 'pii-vault',  label: 'PII Vault',  tag: 'CARGO',      letter: 'P' },
  { id: 'router',     label: 'Router',     tag: 'ATC',        letter: 'R' },
  { id: 'api-keys',   label: 'API Keys',   tag: 'CREDENTIALS',letter: 'K' },
]

interface SidebarProps { 
  current: Page
  onChange: (p: Page) => void 
  agent: string
  onAgentChange: (a: string) => void
}

export default function Sidebar({ current, onChange, agent, onAgentChange }: SidebarProps) {
  const { theme, setTheme } = useTheme()

  return (
    <aside className="relative flex-shrink-0 w-56 flex flex-col border-r bg-sidebar border-sidebar-border">
      {/* Logo */}
      <div className="flex items-center gap-3 px-5 h-14 border-b border-sidebar-border">
        <div className="w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 bg-primary/8 border border-primary/20">
          <svg viewBox="0 0 32 32" fill="none" className="w-4 h-4">
            <circle cx="16" cy="16" r="14" stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.3" className="text-primary" />
            <circle cx="16" cy="16" r="8"  stroke="currentColor" strokeWidth="1.5" strokeOpacity="0.5" className="text-primary" />
            <circle cx="16" cy="16" r="3"  fill="currentColor" className="text-primary" />
            <path d="M16 4V16L24 24" stroke="currentColor" strokeWidth="2" strokeLinecap="round" className="text-primary" />
          </svg>
        </div>
        <div>
          <p className="font-mono font-black text-xs tracking-widest text-foreground">
            AI <span className="text-primary">GATEWAY</span>
          </p>
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground">TOWER ONLINE</p>
        </div>
      </div>

      {/* Tenant badge */}
      <div className="px-4 py-3 border-b border-sidebar-border space-y-3">
        <div className="rounded-md px-3 py-2 bg-primary/4 border border-primary/8">
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-0.5">TENANT</p>
          <p className="font-mono text-xs font-bold text-foreground uppercase">acme-corp</p>
        </div>

        <div className="space-y-1">
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground/60 px-1">ACTIVE SCOPE</p>
          <Select value={agent} onValueChange={onAgentChange}>
            <SelectTrigger className="h-8 bg-background border-sidebar-border font-mono text-[10px] uppercase">
              <SelectValue placeholder="Select Agent" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all" className="font-mono text-[10px] uppercase">All Agents (Global)</SelectItem>
              <SelectItem value="AGT-001" className="font-mono text-[10px] uppercase text-primary">AGT-001 Support</SelectItem>
              <SelectItem value="AGT-002" className="font-mono text-[10px] uppercase text-primary">AGT-002 Data</SelectItem>
              <SelectItem value="AGT-003" className="font-mono text-[10px] uppercase text-primary">AGT-003 Assistant</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 px-3 py-4 space-y-1">
        {navItems.map((item) => {
          const active = current === item.id
          return (
            <Button
              key={item.id}
              variant={active ? 'secondary' : 'ghost'}
              className="w-full justify-start gap-3 h-auto py-2.5 px-3"
              onClick={() => onChange(item.id)}
            >
              <div className={`w-6 h-6 rounded flex items-center justify-center font-mono text-xs font-bold flex-shrink-0 ${active ? 'bg-primary/12 text-primary' : 'bg-muted text-muted-foreground'}`}>
                {item.letter}
              </div>
              <div className="text-left">
                <p className={`text-xs font-semibold leading-none mb-0.5 ${active ? 'text-foreground' : 'text-muted-foreground'}`}>
                  {item.label}
                </p>
                <p className={`font-mono text-[9px] tracking-widest ${active ? 'text-primary' : 'text-muted-foreground/40'}`}>
                  {item.tag}
                </p>
              </div>
            </Button>
          )
        })}
      </nav>

      <Separator className="bg-sidebar-border" />

      {/* Theme Toggle & Bottom stats */}
      <div className="px-4 py-4 space-y-4">
        {/* Theme Toggle */}
        <div className="flex items-center justify-between bg-muted/50 p-1 rounded-lg border border-sidebar-border">
          {[
            { id: 'light',   icon: Sun },
            { id: 'dark',    icon: Moon },
            { id: 'system',  icon: Monitor },
          ].map(({ id, icon: Icon }) => (
            <Button
              key={id}
              variant="ghost"
              size="icon"
              className={`h-7 w-7 rounded-md ${theme === id ? 'bg-background shadow-sm text-primary' : 'text-muted-foreground/60'}`}
              onClick={() => setTheme(id as any)}
            >
              <Icon className="w-3.5 h-3.5" />
            </Button>
          ))}
        </div>

        <div className="space-y-2">
          {[
            { label: 'UPTIME',  value: '99.98%',    color: 'text-success' },
            { label: 'REGION',  value: 'us-east-1', color: 'text-primary' },
          ].map(({ label, value, color }) => (
            <div key={label} className="flex items-center justify-between">
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground/40">{label}</span>
              <span className={`font-mono text-[9px] font-bold ${color}`}>{value}</span>
            </div>
          ))}
        </div>
      </div>
    </aside>
  )
}
