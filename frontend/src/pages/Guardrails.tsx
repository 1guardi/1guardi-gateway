import { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ACTION_STYLES } from '@/lib/styles.ts'
import { guardrailRules } from '../data/mock.ts'
import type { GuardrailRule } from '../data/mock.ts'

function SubGroupLabel({ label }: { label: string }) {
  return (
    <div className="px-4 py-1.5 bg-muted/30 border-b border-border/30">
      <span className="font-mono text-[9px] tracking-widest text-muted-foreground/60">{label}</span>
    </div>
  )
}

function RuleRow({ rule, active, readOnly, onToggle, onClick }: {
  rule: GuardrailRule
  active: boolean
  readOnly?: boolean
  onToggle: (id: string) => void
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
        onCheckedChange={readOnly ? undefined : () => onToggle(rule.id)}
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

function RuleDetail({ rule, readOnly, onToggle, onClose }: {
  rule: GuardrailRule
  readOnly?: boolean
  onToggle: (id: string) => void
  onClose: () => void
}) {
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
        {readOnly ? (
          <p className="font-mono text-[10px] text-muted-foreground/50 text-center pt-1">
            switch to global view to edit
          </p>
        ) : (
          <Button
            variant="outline"
            className={`w-full font-mono text-xs ${rule.enabled ? 'text-error border-error/30 hover:bg-error/8' : 'text-primary border-primary/30 hover:bg-primary/8'}`}
            onClick={() => onToggle(rule.id)}
          >
            {rule.enabled ? 'Disable rule' : 'Enable rule'}
          </Button>
        )}
      </CardContent>
    </Card>
  )
}

function RuleRows({ items, selected, readOnly, onToggle, onSelect }: {
  items: GuardrailRule[]
  selected: GuardrailRule | null
  readOnly?: boolean
  onToggle: (id: string) => void
  onSelect: (rule: GuardrailRule | null) => void
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

export default function Guardrails({ selectedAgent }: { selectedAgent: string }) {
  const [rules, setRules] = useState<GuardrailRule[]>(guardrailRules)
  const [selected, setSelected] = useState<GuardrailRule | null>(null)

  const isAgentMode = selectedAgent !== 'all'

  const toggle = (id: string) => {
    setRules((prev) => prev.map((r) => r.id === id ? { ...r, enabled: !r.enabled } : r))
    setSelected((prev) => prev?.id === id ? { ...prev, enabled: !prev.enabled } : prev)
  }

  const sortByPriority = (a: GuardrailRule, b: GuardrailRule) => b.priority - a.priority

  const globalManaged = rules.filter((r) => !r.agentId && r.managed).sort(sortByPriority)
  const globalCustom  = rules.filter((r) => !r.agentId && !r.managed).sort(sortByPriority)
  const agentManaged  = rules.filter((r) => r.agentId && r.managed  && (!isAgentMode || r.agentId === selectedAgent)).sort(sortByPriority)
  const agentCustom   = rules.filter((r) => r.agentId && !r.managed && (!isAgentMode || r.agentId === selectedAgent)).sort(sortByPriority)

  const globalTotal = globalManaged.length + globalCustom.length
  const agentTotal  = agentManaged.length + agentCustom.length

  const selectedIsGlobal = selected ? !selected.agentId : false

  return (
    <div className="p-6 space-y-5 max-w-7xl">
      <div className="flex items-center justify-between h-14">
        <div>
          <h1 className="font-black text-xl text-foreground tracking-tight">Guardrails</h1>
          <p className="font-mono text-xs mt-0.5 text-muted-foreground">Controlled airspace · {rules.filter((r) => r.enabled).length} active rules</p>
        </div>
        <Button variant="outline" className="font-mono text-xs text-primary border-primary/30 hover:bg-primary/8">
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
            </CardContent>
          </Card>

        </div>

        {selected && (
          <RuleDetail
            rule={selected}
            readOnly={isAgentMode && selectedIsGlobal}
            onToggle={toggle}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
  )
}
