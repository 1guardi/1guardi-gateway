import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { ACTION_STYLES } from '@/lib/styles.ts'
import { useGuardrailRules, useUpdateRule, useCreateRule, useGuardrailEvents, type GuardrailRuleResponse, type CreateRuleRequest } from '../api/guardrails'

// Local UI shape used throughout this file.
interface UIRule {
  id: string
  priority: number
  name: string
  scope: string[]
  action: string
  mode: string
  managed: boolean
  managedId: string
  enabled: boolean
  fires24h: number
  agentId?: string
}

function toUIRule(r: GuardrailRuleResponse): UIRule {
  return {
    id: String(r.ID),
    priority: r.priority,
    name: r.name,
    scope: r.scope ? r.scope.split(',').map((s) => s.trim()).filter(Boolean) : [],
    action: r.action,
    mode: r.mode,
    managed: r.managed,
    managedId: r.managed_id ?? '',
    enabled: r.enabled,
    fires24h: r.fires24h ?? 0,
    agentId: r.agent_id != null ? String(r.agent_id) : undefined,
  }
}

function SubGroupLabel({ label }: { label: string }) {
  return (
    <div className="px-4 py-1.5 bg-muted/30 border-b border-border/30">
      <span className="font-mono text-[9px] tracking-widest text-muted-foreground/60">{label}</span>
    </div>
  )
}

function RuleRow({ rule, active, readOnly, onToggle, onClick }: {
  rule: UIRule
  active: boolean
  readOnly?: boolean
  onToggle: (id: string, enabled: boolean) => void
  onClick: () => void
}) {
  return (
    <div
      className={`flex items-center gap-4 px-4 py-3 cursor-pointer transition-colors hover:bg-muted/20 ${active ? 'bg-primary/4' : ''} ${!rule.enabled ? 'opacity-45' : ''}`}
      onClick={onClick}
    >
      <span className="font-mono text-[10px] font-bold w-5 text-center text-muted-foreground/50 flex-shrink-0">
        {rule.priority}
      </span>
      <Switch
        checked={rule.enabled}
        disabled={readOnly}
        onCheckedChange={readOnly ? undefined : () => onToggle(rule.id, !rule.enabled)}
        onClick={(e) => e.stopPropagation()}
        className="flex-shrink-0"
      />
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-semibold text-foreground truncate">{rule.name}</p>
          {rule.agentId ? (
            <Badge variant="outline" className="font-mono text-[9px] h-4 px-1 text-primary border-primary/20 bg-primary/5 uppercase">
              Agent: {rule.agentId}
            </Badge>
          ) : (
            <Badge variant="outline" className="font-mono text-[9px] h-4 px-1 text-muted-foreground border-muted-foreground/20 uppercase">
              Global
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-2 mt-0.5">
          {rule.scope.map((s) => (
            <span key={s} className="font-mono text-[9px] px-1.5 py-0.5 rounded bg-muted text-muted-foreground">{s}</span>
          ))}
          <span className="font-mono text-[9px] text-muted-foreground/40">{rule.mode}</span>
        </div>
      </div>
      <Badge variant="outline" className={`font-mono text-[10px] flex-shrink-0 ${ACTION_STYLES[rule.action]}`}>
        {rule.action}
      </Badge>
      <div className="text-right flex-shrink-0 w-12">
        <p className={`font-mono text-xs font-bold ${rule.fires24h > 0 ? 'text-warning' : 'text-muted-foreground/30'}`}>
          {rule.fires24h}
        </p>
        <p className="font-mono text-[9px] text-muted-foreground/30">fires/24h</p>
      </div>
    </div>
  )
}

function RuleDetail({ rule, tenantId, readOnly, onToggle, onClose }: {
  rule: UIRule
  tenantId: string | null
  readOnly?: boolean
  onToggle: (id: string, enabled: boolean) => void
  onClose: () => void
}) {
  const { data: events = [], isLoading: eventsLoading } = useGuardrailEvents(tenantId, rule.id, 20)

  const fields = [
    ['ID',        rule.id],
    ['SCOPE',     rule.agentId ? `AGENT: ${rule.agentId}` : 'GLOBAL'],
    ['PRIORITY',  String(rule.priority)],
    ['TARGET',    rule.scope.join(', ')],
    ['MODE',      rule.mode],
    ['MANAGED',   rule.managed ? 'yes' : 'no'],
    ['FIRES 24H', String(rule.fires24h)],
    ['STATUS',    rule.enabled ? 'enabled' : 'disabled'],
  ] as const

  return (
    <Card className="w-80 flex-shrink-0 self-start">
      <CardHeader className="p-6 pb-3 flex flex-row items-start justify-between space-y-0">
        <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">RULE DETAIL</CardTitle>
        <Button variant="ghost" size="icon" className="h-5 w-5 -mt-0.5" onClick={onClose}>✕</Button>
      </CardHeader>
      <CardContent className="p-6 pt-0 space-y-4">
        <p className="font-semibold text-foreground text-sm leading-snug">{rule.name}</p>
        <Badge variant="outline" className={`font-mono text-[10px] ${ACTION_STYLES[rule.action]}`}>{rule.action}</Badge>
        <Separator />
        <div className="space-y-0">
          {fields.map(([label, value]) => (
            <div key={label} className="flex justify-between py-1.5 border-b border-border/50">
              <span className="font-mono text-[9px] tracking-widest text-muted-foreground">{label}</span>
              <span className="font-mono text-xs text-foreground">{value}</span>
            </div>
          ))}
        </div>
        {rule.managedId === 'ml-injection-detection' && (
          <div className="rounded border border-warning/40 bg-warning/5 px-3 py-2 space-y-0.5">
            <p className="font-mono text-[9px] font-bold tracking-widest text-warning">LATENCY NOTE</p>
            <p className="font-mono text-[10px] text-muted-foreground leading-relaxed">
              Adds ~200ms per request. More accurate than regex — catches sophisticated injection attacks patterns miss.
            </p>
          </div>
        )}
        {readOnly ? (
          <p className="font-mono text-[10px] text-muted-foreground/50 text-center pt-1">
            switch to global view to edit
          </p>
        ) : (
          <Button
            variant="outline"
            className={`w-full font-mono text-xs ${rule.enabled ? 'text-error border-error/30 hover:bg-error/8' : 'text-primary border-primary/30 hover:bg-primary/8'}`}
            onClick={() => onToggle(rule.id, !rule.enabled)}
          >
            {rule.enabled ? 'Disable rule' : 'Enable rule'}
          </Button>
        )}
        <Separator />
        <div>
          <p className="font-mono text-[9px] tracking-widest text-muted-foreground mb-2">RECENT FIRES</p>
          {eventsLoading ? (
            <div className="space-y-1.5">
              {[...Array(3)].map((_, i) => <Skeleton key={i} className="h-10 w-full" />)}
            </div>
          ) : events.length === 0 ? (
            <p className="font-mono text-[10px] text-muted-foreground/40 text-center py-3">no events yet</p>
          ) : (
            <ScrollArea className="max-h-64">
              <div className="space-y-1">
                {events.map((ev, i) => (
                  <div key={i} className="rounded border border-border/50 px-2 py-1.5 space-y-0.5">
                    <div className="flex items-center justify-between gap-2">
                      <Badge variant="outline" className={`font-mono text-[9px] h-4 px-1 ${ACTION_STYLES[ev.action] ?? ''}`}>
                        {ev.action}
                      </Badge>
                      <span className="font-mono text-[9px] text-muted-foreground/60 truncate">
                        {ev.timestamp.replace('T', ' ').replace('Z', '')}
                      </span>
                    </div>
                    <p className="font-mono text-[10px] text-muted-foreground truncate" title={ev.reason}>
                      {ev.scope} · {ev.reason || '—'}
                    </p>
                  </div>
                ))}
              </div>
            </ScrollArea>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

function RuleRows({ items, selected, readOnly, onToggle, onSelect }: {
  items: UIRule[]
  selected: UIRule | null
  readOnly?: boolean
  onToggle: (id: string, enabled: boolean) => void
  onSelect: (rule: UIRule | null) => void
}) {
  return (
    <>
      {items.map((rule, i) => (
        <div key={rule.id}>
          <RuleRow
            rule={rule}
            active={selected?.id === rule.id}
            readOnly={readOnly}
            onToggle={onToggle}
            onClick={() => onSelect(selected?.id === rule.id ? null : rule)}
          />
          {i < items.length - 1 && <Separator className="bg-border/50" />}
        </div>
      ))}
    </>
  )
}

const ACTIONS = ['block', 'log', 'tag', 'shadow', 'rewrite'] as const

function NewRuleDialog({ open, onClose, tenantId, selectedAgent }: {
  open: boolean
  onClose: () => void
  tenantId: string | null
  selectedAgent: string
}) {
  const [name, setName] = useState('')
  const [scopeInput, setScopeInput] = useState(true)
  const [scopeOutput, setScopeOutput] = useState(false)
  const [action, setAction] = useState<string>('block')
  const [priority, setPriority] = useState('100')
  const [patterns, setPatterns] = useState('')
  const [matchAll, setMatchAll] = useState(false)
  const [agentScoped, setAgentScoped] = useState(false)

  const { mutate: createRule, isPending } = useCreateRule(tenantId)

  const reset = () => {
    setName(''); setScopeInput(true); setScopeOutput(false)
    setAction('block'); setPriority('100'); setPatterns(''); setMatchAll(false); setAgentScoped(false)
  }

  const isAgentMode = selectedAgent !== 'all'

  const valid = name.trim() !== '' && (scopeInput || scopeOutput) && patterns.trim() !== ''

  const handleSubmit = () => {
    if (!valid) return
    const scope = ([scopeInput && 'input', scopeOutput && 'output'] as (string | false)[]).filter(Boolean) as string[]
    const patternList = patterns.split('\n').map((p) => p.trim()).filter(Boolean)
    const req: CreateRuleRequest = {
      name: name.trim(),
      scope,
      action,
      priority: Number(priority) || 100,
      condition: { type: 'regex', patterns: patternList, match_all: matchAll },
      enabled: true,
    }
    if (isAgentMode && agentScoped) req.agent_id = Number(selectedAgent)
    createRule(req, { onSuccess: () => { reset(); onClose() } })
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) { reset(); onClose() } }}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="font-black tracking-tight">New Guardrail Rule</DialogTitle>
          <DialogDescription className="font-mono text-[10px] tracking-widest text-muted-foreground">
            REGEX CONDITION · CUSTOM
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-1">
          <div className="space-y-1.5">
            <Label className="font-mono text-[10px] tracking-widest text-muted-foreground">NAME</Label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. Block competitor mentions"
              className="font-mono text-xs"
            />
          </div>
          <div className="space-y-1.5">
            <Label className="font-mono text-[10px] tracking-widest text-muted-foreground">PATTERNS (one per line)</Label>
            <Textarea
              value={patterns}
              onChange={(e) => setPatterns(e.target.value)}
              rows={4}
              placeholder={'competitor\\.com\nrival[^s]'}
              className="font-mono text-xs resize-none"
            />
          </div>
          <div className="space-y-1.5">
            <Label className="font-mono text-[10px] tracking-widest text-muted-foreground">SCOPE</Label>
            <div className="flex gap-5">
              {(['input', 'output'] as const).map((s) => {
                const checked = s === 'input' ? scopeInput : scopeOutput
                const toggle = s === 'input' ? setScopeInput : setScopeOutput
                return (
                  <label key={s} className="flex items-center gap-2 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(e) => toggle(e.target.checked)}
                      className="accent-primary"
                    />
                    <span className="font-mono text-xs">{s}</span>
                  </label>
                )
              })}
            </div>
          </div>
          <div className="space-y-1.5">
            <Label className="font-mono text-[10px] tracking-widest text-muted-foreground">ACTION</Label>
            <Select value={action} onValueChange={setAction}>
              <SelectTrigger className="font-mono text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ACTIONS.map((a) => (
                  <SelectItem key={a} value={a} className="font-mono text-xs">{a}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label className="font-mono text-[10px] tracking-widest text-muted-foreground">PRIORITY</Label>
            <Input
              type="number"
              value={priority}
              onChange={(e) => setPriority(e.target.value)}
              className="font-mono text-xs w-28"
            />
          </div>
          <div className="flex items-center justify-between">
            <span className="font-mono text-[10px] tracking-widest text-muted-foreground">MATCH ALL PATTERNS (AND)</span>
            <Switch checked={matchAll} onCheckedChange={setMatchAll} />
          </div>
          {isAgentMode && (
            <div className="flex items-center justify-between">
              <span className="font-mono text-[10px] tracking-widest text-muted-foreground">SCOPE TO AGENT {selectedAgent}</span>
              <Switch checked={agentScoped} onCheckedChange={setAgentScoped} />
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" className="font-mono text-xs" onClick={() => { reset(); onClose() }}>Cancel</Button>
          <Button
            className="font-mono text-xs"
            disabled={!valid || isPending}
            onClick={handleSubmit}
          >
            {isPending ? 'Creating…' : 'Create rule'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default function Guardrails({ selectedAgent, tenantId }: { selectedAgent: string; tenantId: string | null }) {
  const [selected, setSelected] = useState<UIRule | null>(null)
  const [isNewRuleOpen, setIsNewRuleOpen] = useState(false)

  const { data: rawRules = [], isLoading, isError } = useGuardrailRules(tenantId)
  const { mutate: updateRule } = useUpdateRule(tenantId)

  const rules: UIRule[] = rawRules.map(toUIRule)

  const isAgentMode = selectedAgent !== 'all'

  const toggle = (id: string, enabled: boolean) => {
    updateRule({ ruleId: Number(id), enabled })
    // Optimistically update selected panel to avoid stale state.
    setSelected((prev) => prev?.id === id ? { ...prev, enabled } : prev)
  }

  const sortByPriority = (a: UIRule, b: UIRule) => b.priority - a.priority

  const globalManaged = rules.filter((r) => !r.agentId && r.managed).sort(sortByPriority)
  const globalCustom  = rules.filter((r) => !r.agentId && !r.managed).sort(sortByPriority)
  const agentManaged  = rules.filter((r) => r.agentId && r.managed  && (!isAgentMode || r.agentId === selectedAgent)).sort(sortByPriority)
  const agentCustom   = rules.filter((r) => r.agentId && !r.managed && (!isAgentMode || r.agentId === selectedAgent)).sort(sortByPriority)

  const globalTotal = globalManaged.length + globalCustom.length
  const agentTotal  = agentManaged.length + agentCustom.length

  const selectedIsGlobal = selected ? !selected.agentId : false

  if (isLoading) {
    return (
      <div className="p-6 space-y-5 max-w-7xl">
        <Skeleton className="h-14 w-64" />
        <div className="grid grid-cols-4 gap-3">
          {[...Array(4)].map((_, i) => <Skeleton key={i} className="h-24" />)}
        </div>
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (isError) {
    return (
      <div className="p-6">
        <p className="font-mono text-sm text-error">Failed to load guardrail rules.</p>
      </div>
    )
  }

  return (
    <>
    <NewRuleDialog
      open={isNewRuleOpen}
      onClose={() => setIsNewRuleOpen(false)}
      tenantId={tenantId}
      selectedAgent={selectedAgent}
    />
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Guardrails</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Controlled airspace · {rules.filter((r) => r.enabled).length} active rules</p>
        </div>
        <Button
          variant="outline"
          className="font-mono text-xs text-primary border-primary/30 hover:bg-primary/8"
          onClick={() => setIsNewRuleOpen(true)}
        >
          + New rule
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-3">
        {[
          { label: 'ACTIVE RULES', value: rules.filter((r) => r.enabled).length },
          { label: 'FIRES / 24H',  value: rules.reduce((s, r) => s + r.fires24h, 0) },
          { label: 'GLOBAL',       value: globalTotal },
          { label: 'AGENT-LEVEL',  value: agentTotal },
        ].map(({ label, value }) => (
          <Card key={label}>
            <CardHeader className="pb-2">
              <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">{label}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="font-mono text-2xl font-black text-foreground">{value}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Rules + detail */}
      <div className="flex gap-4">
        <div className="flex-1 space-y-3 min-w-0">

          {/* Agent Rules */}
          <Card className="overflow-hidden">
            <CardHeader className="pb-0 border-b border-border">
              <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground pb-3">
                {isAgentMode ? `AGENT RULES — ${selectedAgent}` : 'AGENT RULES'}
              </CardTitle>
            </CardHeader>
            <CardContent className="p-0">
              {agentManaged.length === 0 && agentCustom.length === 0 ? (
                <p className="font-mono text-[10px] text-muted-foreground/40 text-center py-6">no agent-level rules</p>
              ) : (
                <ScrollArea>
                  {agentManaged.length > 0 && (
                    <>
                      <SubGroupLabel label="MANAGED" />
                      <RuleRows items={agentManaged} selected={selected} onToggle={toggle} onSelect={setSelected} />
                    </>
                  )}
                  {agentCustom.length > 0 && (
                    <>
                      {agentManaged.length > 0 && <Separator className="bg-border/40" />}
                      <SubGroupLabel label="CUSTOM" />
                      <RuleRows items={agentCustom} selected={selected} onToggle={toggle} onSelect={setSelected} />
                    </>
                  )}
                </ScrollArea>
              )}
            </CardContent>
          </Card>

          {/* Global Rules */}
          <Card className="overflow-hidden">
            <CardHeader className="pb-0 border-b border-border">
              <div className="flex items-center justify-between pb-3">
                <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">GLOBAL RULES</CardTitle>
                {isAgentMode && (
                  <span className="font-mono text-[9px] text-muted-foreground/50 bg-muted px-2 py-0.5 rounded">reference only</span>
                )}
              </div>
            </CardHeader>
            <CardContent className="p-0">
              {globalManaged.length === 0 && globalCustom.length === 0 ? (
                <p className="font-mono text-[10px] text-muted-foreground/40 text-center py-6">no global rules</p>
              ) : (
                <ScrollArea>
                  {globalManaged.length > 0 && (
                    <>
                      <SubGroupLabel label="MANAGED" />
                      <RuleRows items={globalManaged} selected={selected} readOnly={isAgentMode} onToggle={toggle} onSelect={setSelected} />
                    </>
                  )}
                  {globalCustom.length > 0 && (
                    <>
                      {globalManaged.length > 0 && <Separator className="bg-border/40" />}
                      <SubGroupLabel label="CUSTOM" />
                      <RuleRows items={globalCustom} selected={selected} readOnly={isAgentMode} onToggle={toggle} onSelect={setSelected} />
                    </>
                  )}
                </ScrollArea>
              )}
            </CardContent>
          </Card>

        </div>

        {selected && (
          <RuleDetail
            rule={selected}
            tenantId={tenantId}
            readOnly={isAgentMode && selectedIsGlobal}
            onToggle={toggle}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
    </>
  )
}
