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

function RuleRow({ rule, active, onToggle, onClick }: {
  rule: GuardrailRule
  active: boolean
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
        onCheckedChange={() => onToggle(rule.id)}
        onClick={(e) => e.stopPropagation()}
        className="flex-shrink-0"
      />
      <div className="flex-1 min-w-0">
        <p className="text-sm font-semibold text-foreground truncate">{rule.name}</p>
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
        <p className={`font-mono text-xs font-bold ${rule.fires24h > 0 ? 'text-amber-400' : 'text-muted-foreground/30'}`}>
          {rule.fires24h}
        </p>
        <p className="font-mono text-[9px] text-muted-foreground/30">fires/24h</p>
      </div>
    </div>
  )
}

function RuleDetail({ rule, onToggle, onClose }: { rule: GuardrailRule; onToggle: (id: string) => void; onClose: () => void }) {
  const fields = [
    ['ID',        rule.id],
    ['PRIORITY',  String(rule.priority)],
    ['SCOPE',     rule.scope.join(', ')],
    ['MODE',      rule.mode],
    ['MANAGED',   rule.managed ? 'yes' : 'no'],
    ['FIRES 24H', String(rule.fires24h)],
    ['STATUS',    rule.enabled ? 'enabled' : 'disabled'],
  ] as const

  return (
    <Card className="w-72 flex-shrink-0 self-start">
      <CardHeader className="pb-3 flex-row items-start justify-between space-y-0">
        <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground">RULE DETAIL</CardTitle>
        <Button variant="ghost" size="icon" className="h-5 w-5 -mt-0.5" onClick={onClose}>✕</Button>
      </CardHeader>
      <CardContent className="space-y-4">
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
        <Button
          variant="outline"
          className={`w-full font-mono text-xs ${rule.enabled ? 'text-red-400 border-red-400/30 hover:bg-red-400/8' : 'text-primary border-primary/30 hover:bg-primary/8'}`}
          onClick={() => onToggle(rule.id)}
        >
          {rule.enabled ? 'Disable rule' : 'Enable rule'}
        </Button>
      </CardContent>
    </Card>
  )
}

export default function Guardrails() {
  const [rules, setRules] = useState<GuardrailRule[]>(guardrailRules)
  const [selected, setSelected] = useState<GuardrailRule | null>(null)

  const toggle = (id: string) => {
    setRules((prev) => prev.map((r) => r.id === id ? { ...r, enabled: !r.enabled } : r))
    setSelected((prev) => prev?.id === id ? { ...prev, enabled: !prev.enabled } : prev)
  }

  const managed = rules.filter((r) => r.managed)
  const custom  = rules.filter((r) => !r.managed)

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
          { label: 'MANAGED',      value: managed.length },
          { label: 'CUSTOM',       value: custom.length },
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
          {[{ title: 'MANAGED RULES', items: managed }, { title: 'CUSTOM RULES', items: custom }].map(({ title, items }) => (
            <Card key={title} className="overflow-hidden">
              <CardHeader className="pb-0 border-b border-border">
                <CardTitle className="font-mono text-[10px] tracking-widest text-muted-foreground pb-3">{title}</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <ScrollArea>
                  {items.map((rule, i) => (
                    <div key={rule.id}>
                      <RuleRow
                        rule={rule}
                        active={selected?.id === rule.id}
                        onToggle={toggle}
                        onClick={() => setSelected(selected?.id === rule.id ? null : rule)}
                      />
                      {i < items.length - 1 && <Separator className="bg-border/50" />}
                    </div>
                  ))}
                </ScrollArea>
              </CardContent>
            </Card>
          ))}
        </div>

        {selected && (
          <RuleDetail
            rule={selected}
            onToggle={toggle}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
  )
}
